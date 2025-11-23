package func1

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

type Func1Stats struct {
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

func Func1Run(client *redis_gateway.RedisClient) (*Func1Stats, error) {
	log.Println("[FUNC-1] ========================================")
	log.Println("[FUNC-1] STARTING Func 1")
	log.Println("[FUNC-1] ========================================")
	log.Printf("[FUNC-1] Configuration:")
	log.Printf("[FUNC-1]   - Total Keys: %d", KeyCount)
	log.Printf("[FUNC-1]   - Key Length: %d characters", KeyLength)
	log.Printf("[FUNC-1]   - Value Length: %d characters", ValueLength)
	log.Printf("[FUNC-1]   - Batch Size: %d", BatchSize)
	log.Printf("[FUNC-1]   - Total Data Size: %.2f MB", float64(KeyCount*(KeyLength+ValueLength))/1024/1024)
	log.Println("[FUNC-1] ========================================")

	stats := &Func1Stats{
		TotalKeys: KeyCount,
		Keys:      make([]string, 0, KeyCount),
		Values:    make([]string, 0, KeyCount),
	}

	startTime := time.Now()
	log.Printf("[FUNC-1] Func 1 started at %s", startTime.Format(time.RFC3339))
	log.Printf("[FUNC-1] Initializing keys array with capacity %d", KeyCount)
	log.Printf("[FUNC-1] Initializing values array with capacity %d", KeyCount)

	for i := 0; i < KeyCount; i++ {
		batchNum := i / BatchSize
		keyNum := i % BatchSize

		if keyNum == 0 {
			log.Printf("[FUNC-1] Processing batch %d/%d (keys %d-%d)", 
				batchNum+1, (KeyCount+BatchSize-1)/BatchSize, i, min(i+BatchSize-1, KeyCount-1))
		}

		keyPrefix := fmt.Sprintf("func_key_%d_", i)
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

		log.Printf("[FUNC-1] Setting key #%d (key_len=%d, val_len=%d)", i+1, len(key), len(value))
		
		setStart := time.Now()
		err := client.Set(key, value)
		setDuration := time.Since(setStart)

		if err != nil {
			log.Printf("[FUNC-1] ERROR: Failed to set key #%d: %v", i+1, err)
			stats.FailedKeys++
		} else {
			stats.SuccessfulKeys++
			stats.TotalBytes += int64(len(key) + len(value))
			stats.Keys = append(stats.Keys, key)
			stats.Values = append(stats.Values, value)
			log.Printf("[FUNC-1] Key #%d set successfully in %v (stored key and value in arrays)", i+1, setDuration)
		}

		if (i+1)%100 == 0 {
			elapsed := time.Since(startTime).Seconds()
			rate := float64(i+1) / elapsed
			remaining := KeyCount - (i + 1)
			eta := time.Duration(float64(remaining)/rate) * time.Second
			
			log.Printf("[FUNC-1] Progress: %d/%d keys (%.1f%%) | Rate: %.1f keys/sec | ETA: %v",
				i+1, KeyCount, float64(i+1)/float64(KeyCount)*100, rate, eta)
		}
	}

	duration := time.Since(startTime)
	stats.DurationSeconds = duration.Seconds()
	stats.KeysPerSecond = float64(stats.SuccessfulKeys) / stats.DurationSeconds

	log.Println("[FUNC-1] ========================================")
	log.Println("[FUNC-1] Func 1 COMPLETED")
	log.Println("[FUNC-1] ========================================")
	log.Printf("[FUNC-1] Results:")
	log.Printf("[FUNC-1]   - Total Keys: %d", stats.TotalKeys)
	log.Printf("[FUNC-1]   - Successful: %d", stats.SuccessfulKeys)
	log.Printf("[FUNC-1]   - Failed: %d", stats.FailedKeys)
	log.Printf("[FUNC-1]   - Keys Stored in Array: %d", len(stats.Keys))
	log.Printf("[FUNC-1]   - Values Stored in Array: %d", len(stats.Values))
	log.Printf("[FUNC-1]   - Total Data: %.2f MB", float64(stats.TotalBytes)/1024/1024)
	log.Printf("[FUNC-1]   - Duration: %.2f seconds", stats.DurationSeconds)
	log.Printf("[FUNC-1]   - Throughput: %.2f keys/second", stats.KeysPerSecond)
	log.Printf("[FUNC-1]   - Data Rate: %.2f MB/second", float64(stats.TotalBytes)/1024/1024/stats.DurationSeconds)
	log.Printf("[FUNC-1]   - Keys Array Memory: %.2f MB", float64(len(stats.Keys)*KeyLength)/1024/1024)
	log.Printf("[FUNC-1]   - Values Array Memory: %.2f MB", float64(len(stats.Values)*ValueLength)/1024/1024)
	log.Printf("[FUNC-1]   - Total Array Memory: %.2f MB", float64(len(stats.Keys)*KeyLength+len(stats.Values)*ValueLength)/1024/1024)
	log.Println("[FUNC-1] ========================================")

	if stats.FailedKeys > 0 {
		return stats, fmt.Errorf("Func 1 completed with %d failures", stats.FailedKeys)
	}

	return stats, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
