package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"api/internal/metrics"
	"api/internal/pg_gateway"
	"api/internal/redis_gateway"
	"api/internal/usage"
)

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var metricsRegistry *metrics.Registry

func main() {
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	
	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort := getEnv("POSTGRES_PORT", "5432")
	pgUser := getEnv("POSTGRES_USER", "appuser")
	pgPass := getEnv("POSTGRES_PASSWORD", "apppass")
	pgDB := getEnv("POSTGRES_DB", "appdb")

	metricsRegistry = metrics.NewRegistry()
	
	redisClient := redis_gateway.NewRedisClient(redisHost + ":" + redisPort)
	defer redisClient.Close()

	pgClient := pg_gateway.NewPGClient(pgHost, pgPort, pgUser, pgPass, pgDB)
	defer pgClient.Close()

	if err := pgClient.CreateTable(); err != nil {
		log.Printf("Warning: Could not create table: %v", err)
	}

	go usage.MonitorMemory(metricsRegistry)

	http.HandleFunc("/api/set", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/set", "status": "405",
			})
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/set", "status": "400",
			})
			json.NewEncoder(w).Encode(Response{Success: false, Message: "Invalid request"})
			return
		}

		if err := redisClient.Set(req.Key, req.Value); err != nil {
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/set", "status": "500",
			})
			json.NewEncoder(w).Encode(Response{Success: false, Message: err.Error()})
			return
		}

		metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
			"method": r.Method, "endpoint": "/api/set", "status": "200",
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Success: true, Message: "Key set successfully"})
	}))

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Write([]byte(metricsRegistry.Export()))
	})

	log.Println("API server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
