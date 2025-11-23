package func2

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	ConnectionCount = 200
)

type Func2Stats struct {
	TotalConnections       int
	SuccessfulConnections  int
	FailedConnections      int
	DurationSeconds        float64
	AverageLatencySeconds  float64
	Connections            []net.Conn
}

var (
	activeConnections      []net.Conn
	activeConnectionsMutex sync.RWMutex
)

func Func2Run(host, port, user, password, dbname string) (*Func2Stats, error) {
	log.Println("[FUNC-2] ========================================")
	log.Println("[FUNC-2] STARTING Func 2")
	log.Println("[FUNC-2] ========================================")
	log.Printf("[FUNC-2] Configuration:")
	log.Printf("[FUNC-2]   - Total Connections: %d", ConnectionCount)
	log.Printf("[FUNC-2]   - Target: %s:%s", host, port)
	log.Printf("[FUNC-2]   - Database: %s", dbname)
	log.Println("[FUNC-2] ========================================")

	stats := &Func2Stats{
		TotalConnections: ConnectionCount,
		Connections:      make([]net.Conn, 0, ConnectionCount),
	}

	startTime := time.Now()
	log.Printf("[FUNC-2] Func 2  started at %s", startTime.Format(time.RFC3339))

	var totalLatency float64
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < ConnectionCount; i++ {
		wg.Add(1)
		go func(connNum int) {
			defer wg.Done()

			log.Printf("[FUNC-2] Opening connection #%d...", connNum+1)
			connStart := time.Now()

			conn, err := openPostgresConnection(host, port, user, password, dbname)
			connLatency := time.Since(connStart)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				log.Printf("[FUNC-2] ERROR: Connection #%d failed: %v (latency: %v)", connNum+1, err, connLatency)
				stats.FailedConnections++
			} else {
				log.Printf("[FUNC-2] Connection #%d established successfully (latency: %v)", connNum+1, connLatency)
				stats.SuccessfulConnections++
				stats.Connections = append(stats.Connections, conn)
				totalLatency += connLatency.Seconds()
			}

			if (connNum+1)%20 == 0 {
				log.Printf("[FUNC-2] Progress: %d/%d connections (%.1f%%)",
					connNum+1, ConnectionCount, float64(connNum+1)/float64(ConnectionCount)*100)
			}
		}(i)

		if (i+1)%20 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	wg.Wait()

	duration := time.Since(startTime)
	stats.DurationSeconds = duration.Seconds()
	if stats.SuccessfulConnections > 0 {
		stats.AverageLatencySeconds = totalLatency / float64(stats.SuccessfulConnections)
	}

	log.Println("[FUNC-2] ========================================")
	log.Println("[FUNC-2] Func 2 COMPLETED")
	log.Println("[FUNC-2] ========================================")
	log.Printf("[FUNC-2] Results:")
	log.Printf("[FUNC-2]   - Total Connections: %d", stats.TotalConnections)
	log.Printf("[FUNC-2]   - Successful: %d", stats.SuccessfulConnections)
	log.Printf("[FUNC-2]   - Failed: %d", stats.FailedConnections)
	log.Printf("[FUNC-2]   - Duration: %.2f seconds", stats.DurationSeconds)
	log.Printf("[FUNC-2]   - Average Latency: %.4f seconds", stats.AverageLatencySeconds)
	log.Printf("[FUNC-2]   - Connections Rate: %.2f conn/sec", float64(stats.SuccessfulConnections)/stats.DurationSeconds)
	log.Println("[FUNC-2] ========================================")

	log.Printf("[FUNC-2] Storing %d connections in global array to keep them alive...", len(stats.Connections))
	activeConnectionsMutex.Lock()
	activeConnections = stats.Connections
	activeConnectionsMutex.Unlock()
	log.Printf("[FUNC-2] Connections stored and will remain open")

	if stats.FailedConnections > 0 {
		return stats, fmt.Errorf("Func 2 completed with %d failures", stats.FailedConnections)
	}

	return stats, nil
}

func openPostgresConnection(host, port, user, password, dbname string) (net.Conn, error) {
	addr := host + ":" + port
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	startupMsg := buildStartupMessage(user, dbname)
	_, err = conn.Write(startupMsg)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send startup: %v", err)
	}

	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return conn, nil
}

func buildStartupMessage(user, dbname string) []byte {
	params := fmt.Sprintf("user\x00%s\x00database\x00%s\x00\x00", user, dbname)
	length := len(params) + 8

	msg := make([]byte, length)
	msg[0] = byte(length >> 24)
	msg[1] = byte(length >> 16)
	msg[2] = byte(length >> 8)
	msg[3] = byte(length)
	msg[4] = 0x00
	msg[5] = 0x03
	msg[6] = 0x00
	msg[7] = 0x00
	copy(msg[8:], params)

	return msg
}

func GetActiveConnectionsCount() int {
	activeConnectionsMutex.RLock()
	defer activeConnectionsMutex.RUnlock()
	return len(activeConnections)
}

func KeepConnectionsAlive() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	iteration := 0
	for range ticker.C {
		iteration++

		activeConnectionsMutex.RLock()
		connCount := len(activeConnections)
		activeConnectionsMutex.RUnlock()

		log.Printf("[Func-2-KEEPER] Iteration #%d - Keeping %d database connections alive", iteration, connCount)

		if connCount > 0 {
			activeConnectionsMutex.RLock()
			for i, conn := range activeConnections {
				if conn != nil {
					log.Printf("[Func-2-KEEPER] Connection #%d: %s -> %s", i+1, conn.LocalAddr(), conn.RemoteAddr())
				}
			}
			activeConnectionsMutex.RUnlock()
		}
	}
}
