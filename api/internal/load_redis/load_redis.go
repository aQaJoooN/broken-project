package load_redis

import (
	"api/internal/redis_gateway"
	"crypto/rand"
	"fmt"
	"log"
	"time"
)

const (
	KeyCount      = 5000
	KeyLength     = 4096
	ValueLength   = 10000
	BatchSize     = 100
)

type LoadStats struct {
	TotalKeys       int
	SuccessfulKeys  int
	FailedKeys      int
	TotalBytes      int64
	DurationSeconds float64
	KeysPerSecond   float64
	Keys            []string
	Values          []string
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func LoadRedis(client *redis_gateway.RedisClient) (*LoadStats, error) {
	log.Println("[LOAD-REDIS] ========================================")
	log.Println("[LOAD-REDIS] STARTING REDIS LOAD TEST")
	log.Println("[LOAD-REDIS] ========================================")
	log.Printf("[LOAD-REDIS] Configuration:")
	log.Printf("[LOAD-REDIS]   - Total Keys: %d", KeyCount)
	log.Printf("[LOAD-REDIS]   - Key Length: %d characters", KeyLength)
	log.Printf("[LOAD-REDIS]   - Value Length: %d characters", ValueLength)
	log.Printf("[LOAD-REDIS]   - Batch Size: %d", BatchSize)
	log.Printf("[LOAD-REDIS]   - Total Data Size: %.2f MB", float64(KeyCount*(KeyLength+ValueLength))/1024/1024)
	log.Println("[LOAD-REDIS] ========================================")

	stats := &LoadStats{
		TotalKeys: KeyCount,
		Keys:      make([]string, 0, KeyCount),
		Values:    make([]string, 0, KeyCount),
	}

	startTime := time.Now()
	log.Printf("[LOAD-REDIS] Load test started at %s", startTime.Format(time.RFC3339))
	log.Printf("[LOAD-REDIS] Initializing keys array with capacity %d", KeyCount)
	log.Printf("[LOAD-REDIS] Initializing values array with capacity %d", KeyCount)

	for i := 0; i < KeyCount; i++ {
		batchNum := i / BatchSize
		keyNum := i % BatchSize

		if keyNum == 0 {
			log.Printf("[LOAD-REDIS] Processing batch %d/%d (keys %d-%d)", 
				batchNum+1, (KeyCount+BatchSize-1)/BatchSize, i, min(i+BatchSize-1, KeyCount-1))
		}

		keyPrefix := fmt.Sprintf("load_test_key_%d_", i)
		remainingLength := KeyLength - len(keyPrefix)
		if remainingLength < 0 {
			remainingLength = 0
		}
		key := keyPrefix + generateRandomString(remainingLength)
		
		// Ensure exact length
		if len(key) > KeyLength {
			key = key[:KeyLength]
		} else if len(key) < KeyLength {
			key = key + generateRandomString(KeyLength-len(key))
		}
		
		value := generateRandomString(ValueLength)

		log.Printf("[LOAD-REDIS] Setting key #%d (key_len=%d, val_len=%d)", i+1, len(key), len(value))
		
		setStart := time.Now()
		err := client.Set(key, value)
		setDuration := time.Since(setStart)

		if err != nil {
			log.Printf("[LOAD-REDIS] ERROR: Failed to set key #%d: %v", i+1, err)
			stats.FailedKeys++
		} else {
			stats.SuccessfulKeys++
			stats.TotalBytes += int64(len(key) + len(value))
			stats.Keys = append(stats.Keys, key)
			stats.Values = append(stats.Values, value)
			log.Printf("[LOAD-REDIS] Key #%d set successfully in %v (stored key and value in arrays)", i+1, setDuration)
		}

		if (i+1)%100 == 0 {
			elapsed := time.Since(startTime).Seconds()
			rate := float64(i+1) / elapsed
			remaining := KeyCount - (i + 1)
			eta := time.Duration(float64(remaining)/rate) * time.Second
			
			log.Printf("[LOAD-REDIS] Progress: %d/%d keys (%.1f%%) | Rate: %.1f keys/sec | ETA: %v",
				i+1, KeyCount, float64(i+1)/float64(KeyCount)*100, rate, eta)
		}
	}

	duration := time.Since(startTime)
	stats.DurationSeconds = duration.Seconds()
	stats.KeysPerSecond = float64(stats.SuccessfulKeys) / stats.DurationSeconds

	log.Println("[LOAD-REDIS] ========================================")
	log.Println("[LOAD-REDIS] LOAD TEST COMPLETED")
	log.Println("[LOAD-REDIS] ========================================")
	log.Printf("[LOAD-REDIS] Results:")
	log.Printf("[LOAD-REDIS]   - Total Keys: %d", stats.TotalKeys)
	log.Printf("[LOAD-REDIS]   - Successful: %d", stats.SuccessfulKeys)
	log.Printf("[LOAD-REDIS]   - Failed: %d", stats.FailedKeys)
	log.Printf("[LOAD-REDIS]   - Keys Stored in Array: %d", len(stats.Keys))
	log.Printf("[LOAD-REDIS]   - Values Stored in Array: %d", len(stats.Values))
	log.Printf("[LOAD-REDIS]   - Total Data: %.2f MB", float64(stats.TotalBytes)/1024/1024)
	log.Printf("[LOAD-REDIS]   - Duration: %.2f seconds", stats.DurationSeconds)
	log.Printf("[LOAD-REDIS]   - Throughput: %.2f keys/second", stats.KeysPerSecond)
	log.Printf("[LOAD-REDIS]   - Data Rate: %.2f MB/second", float64(stats.TotalBytes)/1024/1024/stats.DurationSeconds)
	log.Printf("[LOAD-REDIS]   - Keys Array Memory: %.2f MB", float64(len(stats.Keys)*KeyLength)/1024/1024)
	log.Printf("[LOAD-REDIS]   - Values Array Memory: %.2f MB", float64(len(stats.Values)*ValueLength)/1024/1024)
	log.Printf("[LOAD-REDIS]   - Total Array Memory: %.2f MB", float64(len(stats.Keys)*KeyLength+len(stats.Values)*ValueLength)/1024/1024)
	log.Println("[LOAD-REDIS] ========================================")

	if stats.FailedKeys > 0 {
		return stats, fmt.Errorf("load test completed with %d failures", stats.FailedKeys)
	}

	return stats, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
