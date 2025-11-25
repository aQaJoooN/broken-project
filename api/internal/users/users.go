package users

import (
	"api/internal/metrics"
	"api/internal/pg_gateway"
	"api/internal/redis_gateway"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type User struct {
	UserID        string `json:"user_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Age           int    `json:"age"`
	MaritalStatus bool   `json:"marital_status"`
}

type UsersManager struct {
	redisClient     *redis_gateway.RedisClient
	pgClient        *pg_gateway.PGClient
	metricsRegistry *metrics.Registry
}

func NewUsersManager(redisClient *redis_gateway.RedisClient, pgClient *pg_gateway.PGClient, metricsRegistry *metrics.Registry) *UsersManager {
	log.Println("[USERS] Creating new UsersManager")
	return &UsersManager{
		redisClient:     redisClient,
		pgClient:        pgClient,
		metricsRegistry: metricsRegistry,
	}
}

func (um *UsersManager) CreateUser(firstName, lastName string, age int, maritalStatus bool) (string, error) {
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	operationStart := time.Now()
	
	log.Printf("[USERS:%s] Creating user: first_name='%s', last_name='%s', age=%d, marital_status=%t", 
		requestID, firstName, lastName, age, maritalStatus)
	
	// Generate UUID for user
	userID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))
	redisKey := fmt.Sprintf("user:%s", userID)
	
	log.Printf("[USERS:%s] Generated user_id: %s", requestID, userID)
	log.Printf("[USERS:%s] Redis key: %s", requestID, redisKey)
	
	// Create JSON value for Redis
	userJSON := fmt.Sprintf(`{"first_name":"%s","last_name":"%s","age":%d,"marital_status":%t}`,
		firstName, lastName, age, maritalStatus)
	
	log.Printf("[USERS:%s] Storing to Redis...", requestID)
	setStart := time.Now()
	if err := um.redisClient.Set(redisKey, userJSON); err != nil {
		log.Printf("[USERS:%s] ERROR: Redis SET failed: %v", requestID, err)
		um.metricsRegistry.IncrementCounter("user_operations_total", map[string]string{
			"operation": "create", "status": "error", "source": "redis",
		})
		return "", err
	}
	setDuration := time.Since(setStart)
	log.Printf("[USERS:%s] Redis SET completed in %v", requestID, setDuration)
	
	log.Printf("[USERS:%s] Storing to PostgreSQL...", requestID)
	insertStart := time.Now()
	if err := um.pgClient.InsertUser(userID, firstName, lastName, age, maritalStatus); err != nil {
		log.Printf("[USERS:%s] ERROR: PostgreSQL INSERT failed: %v", requestID, err)
		um.metricsRegistry.IncrementCounter("user_operations_total", map[string]string{
			"operation": "create", "status": "error", "source": "postgres",
		})
		return "", fmt.Errorf("failed to insert into database")
	}
	insertDuration := time.Since(insertStart)
	log.Printf("[USERS:%s] PostgreSQL INSERT completed in %v", requestID, insertDuration)

	totalDuration := time.Since(operationStart)
	um.metricsRegistry.IncrementCounter("user_operations_total", map[string]string{
		"operation": "create", "status": "success",
	})
	um.metricsRegistry.SetGauge("user_operation_duration_seconds", totalDuration.Seconds(), map[string]string{
		"operation": "create",
	})
	
	log.Printf("[USERS:%s] SUCCESS: User '%s' created successfully (user_id: %s) in %v", 
		requestID, firstName+" "+lastName, userID, totalDuration)
	
	return userID, nil
}

func (um *UsersManager) GetUsers() ([]map[string]interface{}, error) {
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	operationStart := time.Now()
	
	log.Printf("[USERS:%s] Getting all users...", requestID)
	
	// First, try to get users from Redis
	log.Printf("[USERS:%s] Attempting to get users from Redis...", requestID)
	redisStart := time.Now()
	users, err := um.redisClient.GetAllUsers()
	redisDuration := time.Since(redisStart)
	
	if err != nil || len(users) == 0 {
		if err != nil {
			log.Printf("[USERS:%s] WARNING: Failed to get users from Redis: %v", requestID, err)
		} else {
			log.Printf("[USERS:%s] WARNING: No users found in Redis", requestID)
		}
		
		um.metricsRegistry.IncrementCounter("user_operations_total", map[string]string{
			"operation": "get", "status": "redis_empty",
		})
		
		log.Printf("[USERS:%s] ERROR:  Redis Does not Response Properly Waiting 30 seconds before querying PostgreSQL...", requestID)
		time.Sleep(30 * time.Second)
		log.Printf("[USERS:%s] Wait completed, querying PostgreSQL...", requestID)
		
		// Get users from PostgreSQL
		pgStart := time.Now()
		users, err = um.pgClient.GetAllUsers()
		pgDuration := time.Since(pgStart)
		
		if err != nil {
			log.Printf("[USERS:%s] ERROR: Failed to get users from PostgreSQL: %v", requestID, err)
			um.metricsRegistry.IncrementCounter("user_operations_total", map[string]string{
				"operation": "get", "status": "error", "source": "postgres",
			})
			return nil, err
		}
		
		log.Printf("[USERS:%s] Retrieved %d users from PostgreSQL in %v", requestID, len(users), pgDuration)
		um.metricsRegistry.SetGauge("user_operation_duration_seconds", pgDuration.Seconds(), map[string]string{
			"operation": "get", "source": "postgres",
		})
	} else {
		log.Printf("[USERS:%s] Retrieved %d users from Redis in %v", requestID, len(users), redisDuration)
		um.metricsRegistry.SetGauge("user_operation_duration_seconds", redisDuration.Seconds(), map[string]string{
			"operation": "get", "source": "redis",
		})
	}
	
	totalDuration := time.Since(operationStart)
	um.metricsRegistry.IncrementCounter("user_operations_total", map[string]string{
		"operation": "get", "status": "success",
	})
	um.metricsRegistry.SetGauge("users_retrieved_count", float64(len(users)), map[string]string{})
	um.metricsRegistry.SetGauge("user_operation_duration_seconds", totalDuration.Seconds(), map[string]string{
		"operation": "get_total",
	})
	
	log.Printf("[USERS:%s] SUCCESS: Retrieved %d users in %v", requestID, len(users), totalDuration)
	
	return users, nil
}
