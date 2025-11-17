package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	
	log.Println("========================================")
	log.Println("APPLICATION STARTUP INITIATED")
	log.Println("========================================")
	
	log.Println("[INIT] Reading environment variables...")
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	log.Printf("[INIT] Redis configuration: host=%s, port=%s", redisHost, redisPort)
	
	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort := getEnv("POSTGRES_PORT", "5432")
	pgUser := getEnv("POSTGRES_USER", "appuser")
	pgPass := getEnv("POSTGRES_PASSWORD", "apppass")
	pgDB := getEnv("POSTGRES_DB", "appdb")
	log.Printf("[INIT] PostgreSQL configuration: host=%s, port=%s, user=%s, db=%s", pgHost, pgPort, pgUser, pgDB)

	log.Println("[INIT] Initializing metrics registry...")
	metricsRegistry = metrics.NewRegistry()
	log.Println("[INIT] Metrics registry initialized successfully")
	
	log.Printf("[REDIS] Attempting to connect to Redis at %s:%s...", redisHost, redisPort)
	startTime := time.Now()
	redisClient := redis_gateway.NewRedisClient(redisHost + ":" + redisPort)
	defer redisClient.Close()
	log.Printf("[REDIS] Connected successfully in %v", time.Since(startTime))

	log.Printf("[POSTGRES] Attempting to connect to PostgreSQL at %s:%s...", pgHost, pgPort)
	startTime = time.Now()
	pgClient := pg_gateway.NewPGClient(pgHost, pgPort, pgUser, pgPass, pgDB)
	defer pgClient.Close()
	log.Printf("[POSTGRES] Connected successfully in %v", time.Since(startTime))

	log.Println("[POSTGRES] Creating database table if not exists...")
	if err := pgClient.CreateTable(); err != nil {
		log.Printf("[POSTGRES] WARNING: Could not create table: %v", err)
	} else {
		log.Println("[POSTGRES] Table created/verified successfully")
	}

	log.Println("[MONITOR] Starting memory monitoring goroutine...")
	go usage.MonitorMemory(metricsRegistry)
	log.Println("[MONITOR] Memory monitoring started")

	log.Println("[HTTP] Registering /api/set endpoint...")
	http.HandleFunc("/api/set", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		log.Printf("[REQUEST:%s] Incoming %s request to /api/set from %s", requestID, r.Method, r.RemoteAddr)
		log.Printf("[REQUEST:%s] Headers: %v", requestID, r.Header)
		
		if r.Method != http.MethodPost {
			log.Printf("[REQUEST:%s] ERROR: Method not allowed: %s", requestID, r.Method)
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/set", "status": "405",
			})
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[REQUEST:%s] ERROR: Failed to decode JSON body: %v", requestID, err)
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/set", "status": "400",
			})
			json.NewEncoder(w).Encode(Response{Success: false, Message: "Invalid request"})
			return
		}
		
		log.Printf("[REQUEST:%s] Decoded payload: key='%s', value='%s'", requestID, req.Key, req.Value)
		log.Printf("[REQUEST:%s] Sending SET command to Redis...", requestID)
		
		setStart := time.Now()
		if err := redisClient.Set(req.Key, req.Value); err != nil {
			log.Printf("[REQUEST:%s] ERROR: Redis SET failed: %v", requestID, err)
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/set", "status": "500",
			})
			json.NewEncoder(w).Encode(Response{Success: false, Message: err.Error()})
			return
		}
		log.Printf("[REQUEST:%s] Redis SET completed in %v", requestID, time.Since(setStart))

		metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
			"method": r.Method, "endpoint": "/api/set", "status": "200",
		})
		log.Printf("[REQUEST:%s] SUCCESS: Key '%s' set successfully", requestID, req.Key)
		log.Printf("[REQUEST:%s] Sending response to client", requestID)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Success: true, Message: "Key set successfully"})
		log.Printf("[REQUEST:%s] Request completed successfully", requestID)
	}))
	log.Println("[HTTP] /api/set endpoint registered")

	log.Println("[HTTP] Registering /metrics endpoint...")
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		log.Printf("[METRICS:%s] Incoming request from %s", requestID, r.RemoteAddr)
		log.Printf("[METRICS:%s] Exporting metrics...", requestID)
		
		metricsData := metricsRegistry.Export()
		log.Printf("[METRICS:%s] Metrics exported, size: %d bytes", requestID, len(metricsData))
		
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Write([]byte(metricsData))
		log.Printf("[METRICS:%s] Metrics sent to client", requestID)
	})
	log.Println("[HTTP] /metrics endpoint registered")

	log.Println("========================================")
	log.Println("API SERVER READY")
	log.Println("Listening on :8080")
	log.Println("Endpoints:")
	log.Println("  - POST /api/set")
	log.Println("  - GET  /metrics")
	log.Println("========================================")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
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
