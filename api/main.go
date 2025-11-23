package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"api/internal/func1"
	"api/internal/func2"
	"api/internal/metrics"
	"api/internal/pg_gateway"
	"api/internal/redis_gateway"
	"api/internal/usage"
)

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UserRequest struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Age           int    `json:"age"`
	MaritalStatus bool   `json:"marital_status"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var (
	metricsRegistry *metrics.Registry
	loadedKeys      []string
	loadedKeysMutex sync.RWMutex
	loadedValues    []string
	loadedValuesMutex sync.RWMutex
)

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
	redisClient.SetMetricsRegistry(metricsRegistry)
	log.Println("[REDIS] Metrics registry attached to Redis client")

	log.Printf("[POSTGRES] Attempting to connect to PostgreSQL at %s:%s...", pgHost, pgPort)
	startTime = time.Now()
	pgClient := pg_gateway.NewPGClient(pgHost, pgPort, pgUser, pgPass, pgDB)
	defer pgClient.Close()
	log.Printf("[POSTGRES] Connected successfully in %v", time.Since(startTime))
	pgClient.SetMetricsRegistry(metricsRegistry)
	log.Println("[POSTGRES] Metrics registry attached to PostgreSQL client")

	log.Println("[POSTGRES] Creating database table if not exists...")
	if err := pgClient.CreateTable(); err != nil {
		log.Printf("[POSTGRES] WARNING: Could not create table: %v", err)
	} else {
		log.Println("[POSTGRES] Table created/verified successfully")
	}

	log.Println("[MONITOR] Starting memory monitoring goroutine...")
	go usage.MonitorMemory(metricsRegistry)
	log.Println("[MONITOR] Memory monitoring started")
	
	log.Println("[MONITOR] Starting array keeper goroutine to prevent GC...")
	go keepArraysAlive()
	log.Println("[MONITOR] Array keeper started")
	
	log.Println("[MONITOR] Starting database connections keeper goroutine...")
	go func2.KeepConnectionsAlive()
	log.Println("[MONITOR] Database connections keeper started")
	
	log.Println("[MONITOR] Starting database connection keeper goroutine...")
	go func2.KeepConnectionsAlive()
	log.Println("[MONITOR] Database connection keeper started")

	metricsRegistry.SetGauge("redis_connection_status", 1, map[string]string{})
	metricsRegistry.SetGauge("postgres_connection_status", 1, map[string]string{})

	log.Println("[HTTP] Registering /api/user endpoint...")
	http.HandleFunc("/api/user", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		requestStart := time.Now()
		
		log.Printf("[USER:%s] Incoming %s request to /api/user from %s", requestID, r.Method, r.RemoteAddr)
		log.Printf("[USER:%s] Headers: %v", requestID, r.Header)
		
		if r.Method != http.MethodPost {
			log.Printf("[USER:%s] ERROR: Method not allowed: %s", requestID, r.Method)
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/user", "status": "405",
			})
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req UserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[USER:%s] ERROR: Failed to decode JSON body: %v", requestID, err)
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/user", "status": "400",
			})
			json.NewEncoder(w).Encode(Response{Success: false, Message: "Invalid request"})
			return
		}
		
		log.Printf("[USER:%s] Decoded payload: first_name='%s', last_name='%s', age=%d, marital_status=%t", 
			requestID, req.FirstName, req.LastName, req.Age, req.MaritalStatus)
		
		// Generate UUID for user
		userID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))
		redisKey := fmt.Sprintf("user:%s", userID)
		
		log.Printf("[USER:%s] Generated user_id: %s", requestID, userID)
		log.Printf("[USER:%s] Redis key: %s", requestID, redisKey)
		
		// Create JSON value for Redis
		userJSON := fmt.Sprintf(`{"first_name":"%s","last_name":"%s","age":%d,"marital_status":%t}`,
			req.FirstName, req.LastName, req.Age, req.MaritalStatus)
		
		log.Printf("[USER:%s] Storing to Redis...", requestID)
		setStart := time.Now()
		if err := redisClient.Set(redisKey, userJSON); err != nil {
			log.Printf("[USER:%s] ERROR: Redis SET failed: %v", requestID, err)
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/user", "status": "500",
			})
			metricsRegistry.IncrementCounter("redis_operations_total", map[string]string{
				"operation": "set", "status": "error",
			})
			json.NewEncoder(w).Encode(Response{Success: false, Message: err.Error()})
			return
		}
		setDuration := time.Since(setStart)
		log.Printf("[USER:%s] Redis SET completed in %v", requestID, setDuration)
		
		log.Printf("[USER:%s] Storing to PostgreSQL...", requestID)
		insertStart := time.Now()
		if err := pgClient.InsertUser(userID, req.FirstName, req.LastName, req.Age, req.MaritalStatus); err != nil {
			log.Printf("[USER:%s] ERROR: PostgreSQL INSERT failed: %v", requestID, err)
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/user", "status": "500",
			})
			json.NewEncoder(w).Encode(Response{Success: false, Message: "Failed to insert into database"})
			return
		}
		insertDuration := time.Since(insertStart)
		log.Printf("[USER:%s] PostgreSQL INSERT completed in %v", requestID, insertDuration)

		metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
			"method": r.Method, "endpoint": "/api/user", "status": "200",
		})
		metricsRegistry.IncrementCounter("redis_operations_total", map[string]string{
			"operation": "set", "status": "success",
		})
		metricsRegistry.IncrementCounter("user_created_total", map[string]string{})
		metricsRegistry.SetGauge("http_request_duration_seconds", time.Since(requestStart).Seconds(), map[string]string{
			"endpoint": "/api/user",
		})
		metricsRegistry.SetGauge("app_goroutines", float64(runtime.NumGoroutine()), map[string]string{})
		
		log.Printf("[USER:%s] SUCCESS: User '%s' created successfully (user_id: %s)", requestID, req.FirstName+" "+req.LastName, userID)
		log.Printf("[USER:%s] Sending response to client", requestID)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "User created successfully",
			"user_id": userID,
		})
		log.Printf("[USER:%s] Request completed successfully", requestID)
	}))
	log.Println("[HTTP] /api/user endpoint registered")

	log.Println("[HTTP] Registering /api/set endpoint...")
	http.HandleFunc("/api/set", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		requestStart := time.Now()
		
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
			metricsRegistry.IncrementCounter("redis_operations_total", map[string]string{
				"operation": "set", "status": "error",
			})
			json.NewEncoder(w).Encode(Response{Success: false, Message: err.Error()})
			return
		}
		setDuration := time.Since(setStart)
		log.Printf("[REQUEST:%s] Redis SET completed in %v", requestID, setDuration)

		metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
			"method": r.Method, "endpoint": "/api/set", "status": "200",
		})
		metricsRegistry.IncrementCounter("redis_operations_total", map[string]string{
			"operation": "set", "status": "success",
		})
		metricsRegistry.SetGauge("http_request_duration_seconds", time.Since(requestStart).Seconds(), map[string]string{
			"endpoint": "/api/set",
		})
		metricsRegistry.SetGauge("app_goroutines", float64(runtime.NumGoroutine()), map[string]string{})
		
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

	log.Println("[HTTP] Registering /api/func1 endpoint...")
	http.HandleFunc("/api/func1", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		log.Printf("[FUNC1:%s] Incoming %s request to /api/func1 from %s", requestID, r.Method, r.RemoteAddr)
		
		if r.Method != http.MethodGet {
			log.Printf("[FUNC1:%s] ERROR: Method not allowed: %s", requestID, r.Method)
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/func1", "status": "405",
			})
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Printf("[FUNC1:%s] Starting func1 ..", requestID)
		metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
			"method": r.Method, "endpoint": "/api/func1", "status": "202",
		})
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(Response{Success: true, Message: "Func1 started"})
		log.Printf("[FUNC1:%s] Response sent, starting func1 in background", requestID)

		go func() {
			log.Printf("[FUNC1:%s] Background func1 initiated", requestID)
			funcStart := time.Now()
			
			metricsRegistry.IncrementCounter("func1_runs_total", map[string]string{"status": "started"})
			
			stats, err := func1.Func1Run(redisClient)
			funcDuration := time.Since(funcStart)
			
			if err != nil {
				log.Printf("[FUNC1:%s] ERROR: Func1 failed: %v", requestID, err)
				metricsRegistry.IncrementCounter("func1_runs_total", map[string]string{"status": "failed"})
				metricsRegistry.SetGauge("func1_failed_keys", float64(stats.FailedKeys), map[string]string{})
			} else {
				log.Printf("[FUNC1:%s] SUCCESS: Func1 completed", requestID)
				metricsRegistry.IncrementCounter("func1_runs_total", map[string]string{"status": "success"})
				metricsRegistry.SetGauge("func1_successful_keys", float64(stats.SuccessfulKeys), map[string]string{})
				metricsRegistry.SetGauge("func1_failed_keys", float64(stats.FailedKeys), map[string]string{})
				metricsRegistry.SetGauge("func1_duration_seconds", stats.DurationSeconds, map[string]string{})
				metricsRegistry.SetGauge("func1_throughput_keys_per_sec", stats.KeysPerSecond, map[string]string{})
				metricsRegistry.SetGauge("func1_total_bytes", float64(stats.TotalBytes), map[string]string{})
				
				log.Printf("[FUNC1:%s] Storing %d keys and %d values in application memory...", requestID, len(stats.Keys), len(stats.Values))
				loadedKeysMutex.Lock()
				loadedKeys = stats.Keys
				loadedKeysMutex.Unlock()
				
				loadedValuesMutex.Lock()
				loadedValues = stats.Values
				loadedValuesMutex.Unlock()
				
				log.Printf("[FUNC1:%s] Keys stored in application array (total: %d keys)", requestID, len(loadedKeys))
				log.Printf("[FUNC1:%s] Values stored in application array (total: %d values)", requestID, len(loadedValues))
				log.Printf("[FUNC1:%s] Keys array memory usage: %.2f MB", requestID, float64(len(loadedKeys)*4096)/1024/1024)
				log.Printf("[FUNC1:%s] Values array memory usage: %.2f MB", requestID, float64(len(loadedValues)*10000)/1024/1024)
				log.Printf("[FUNC1:%s] Total array memory usage: %.2f MB", requestID, float64(len(loadedKeys)*4096+len(loadedValues)*10000)/1024/1024)
				
				metricsRegistry.SetGauge("app_loaded_keys_count", float64(len(loadedKeys)), map[string]string{})
				metricsRegistry.SetGauge("app_loaded_values_count", float64(len(loadedValues)), map[string]string{})
			}
			
			log.Printf("[FUNC1:%s] Func1 completed in %v", requestID, funcDuration)
		}()
	}))
	log.Println("[HTTP] /api/func1 endpoint registered")

	log.Println("[HTTP] Registering /api/func2 endpoint...")
	http.HandleFunc("/api/func2", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		log.Printf("[FUNC-2:%s] Incoming %s request to /api/func2 from %s", requestID, r.Method, r.RemoteAddr)
		
		if r.Method != http.MethodGet {
			log.Printf("[FUNC-2:%s] ERROR: Method not allowed: %s", requestID, r.Method)
			metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
				"method": r.Method, "endpoint": "/api/func2", "status": "405",
			})
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Printf("[FUNC-2:%s] Starting Func 2 ...", requestID)
		metricsRegistry.IncrementCounter("api_requests_total", map[string]string{
			"method": r.Method, "endpoint": "/api/func2", "status": "202",
		})
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(Response{Success: true, Message: "Func 2 started"})
		log.Printf("[FUNC-2:%s] Response sent, starting Func 2 in background", requestID)

		go func() {
			log.Printf("[FUNC-2:%s] Background Func 2 initiated", requestID)
			funcStart := time.Now()
			
			metricsRegistry.IncrementCounter("func2_runs_total", map[string]string{"status": "started"})
			
			stats, err := func2.Func2Run(pgHost, pgPort, pgUser, pgPass, pgDB)
			funcDuration := time.Since(funcStart)
			
			if err != nil {
				log.Printf("[FUNC-2:%s] ERROR: Func 2 failed: %v", requestID, err)
				metricsRegistry.IncrementCounter("func2_runs_total", map[string]string{"status": "failed"})
			} else {
				log.Printf("[FUNC-2:%s] SUCCESS: Func 2 completed", requestID)
				metricsRegistry.IncrementCounter("func2_runs_total", map[string]string{"status": "success"})
				metricsRegistry.SetGauge("func2_connections", float64(stats.SuccessfulConnections), map[string]string{})
				metricsRegistry.SetGauge("func2_duration_seconds", stats.DurationSeconds, map[string]string{})
				metricsRegistry.SetGauge("func2_avg_latency_seconds", stats.AverageLatencySeconds, map[string]string{})
				metricsRegistry.SetGauge("db_active_connections_count", float64(func2.GetActiveConnectionsCount()), map[string]string{})
			}
			
			log.Printf("[FUNC-2:%s] Func 2 completed in %v", requestID, funcDuration)
		}()
	}))
	log.Println("[HTTP] /api/func2 endpoint registered")

	log.Println("========================================")
	log.Println("API SERVER READY")
	log.Println("Listening on :8080")
	log.Println("Endpoints:")
	log.Println("  - POST /api/user")
	log.Println("  - POST /api/set")
	log.Println("  - GET  /api/func1")
	log.Println("  - GET  /api/func2")
	log.Println("  - GET  /metrics")
	log.Println("========================================")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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

func keepArraysAlive() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	iteration := 0
	for range ticker.C {
		iteration++
		
		loadedKeysMutex.RLock()
		keysLen := len(loadedKeys)
		var keysSample string
		if keysLen > 0 {
			keysSample = loadedKeys[0][:min(50, len(loadedKeys[0]))]
		}
		loadedKeysMutex.RUnlock()
		
		loadedValuesMutex.RLock()
		valuesLen := len(loadedValues)
		var valuesSample string
		if valuesLen > 0 {
			valuesSample = loadedValues[0][:min(50, len(loadedValues[0]))]
		}
		loadedValuesMutex.RUnlock()
		
		log.Printf("[KEEPER] Iteration #%d - Keeping arrays alive", iteration)
		log.Printf("[KEEPER] Keys array: %d elements, sample: %s...", keysLen, keysSample)
		log.Printf("[KEEPER] Values array: %d elements, sample: %s...", valuesLen, valuesSample)
		log.Printf("[KEEPER] Total memory held: %.2f MB", float64(keysLen*4096+valuesLen*10000)/1024/1024)
		
		runtime.KeepAlive(loadedKeys)
		runtime.KeepAlive(loadedValues)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
