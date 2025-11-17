package redis_gateway

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type RedisClient struct {
	conn net.Conn
	addr string
}

func NewRedisClient(addr string) *RedisClient {
	log.Printf("[REDIS] Dialing TCP connection to %s...", addr)
	startTime := time.Now()
	
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("[REDIS] FATAL: Failed to connect to Redis: %v", err)
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}
	
	log.Printf("[REDIS] TCP connection established in %v", time.Since(startTime))
	log.Printf("[REDIS] Local address: %s", conn.LocalAddr())
	log.Printf("[REDIS] Remote address: %s", conn.RemoteAddr())

	return &RedisClient{
		conn: conn,
		addr: addr,
	}
}

func (r *RedisClient) Set(key, value string) error {
	log.Printf("[REDIS] Building RESP command for SET key='%s' value='%s'", key, value)
	cmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
		len(key), key, len(value), value)
	
	log.Printf("[REDIS] RESP command: %q", cmd)
	log.Printf("[REDIS] Command size: %d bytes", len(cmd))
	log.Printf("[REDIS] Writing command to socket...", )
	
	startWrite := time.Now()
	bytesWritten, err := r.conn.Write([]byte(cmd))
	if err != nil {
		log.Printf("[REDIS] ERROR: Failed to write to socket: %v", err)
		return err
	}
	log.Printf("[REDIS] Wrote %d bytes in %v", bytesWritten, time.Since(startWrite))

	log.Printf("[REDIS] Reading response from Redis...")
	reader := bufio.NewReader(r.conn)
	startRead := time.Now()
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("[REDIS] ERROR: Failed to read response: %v", err)
		return err
	}
	log.Printf("[REDIS] Received response in %v: %q", time.Since(startRead), response)

	if !strings.HasPrefix(response, "+OK") {
		log.Printf("[REDIS] ERROR: Unexpected response from Redis: %s", response)
		return fmt.Errorf("Redis error: %s", response)
	}

	log.Printf("[REDIS] SET operation successful")
	return nil
}

func (r *RedisClient) Get(key string) (string, error) {
	log.Printf("[REDIS] Building RESP command for GET key='%s'", key)
	cmd := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
	
	log.Printf("[REDIS] RESP command: %q", cmd)
	log.Printf("[REDIS] Writing GET command to socket...")
	
	bytesWritten, err := r.conn.Write([]byte(cmd))
	if err != nil {
		log.Printf("[REDIS] ERROR: Failed to write GET command: %v", err)
		return "", err
	}
	log.Printf("[REDIS] Wrote %d bytes", bytesWritten)

	log.Printf("[REDIS] Reading GET response...")
	reader := bufio.NewReader(r.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("[REDIS] ERROR: Failed to read GET response: %v", err)
		return "", err
	}
	log.Printf("[REDIS] Received response: %q", response)

	if strings.HasPrefix(response, "$-1") {
		log.Printf("[REDIS] Key not found in Redis")
		return "", fmt.Errorf("key not found")
	}

	if strings.HasPrefix(response, "$") {
		lengthStr := strings.TrimSpace(response[1:])
		length, _ := strconv.Atoi(lengthStr)
		log.Printf("[REDIS] Value length: %d bytes", length)
		
		value := make([]byte, length)
		bytesRead, err := reader.Read(value)
		if err != nil {
			log.Printf("[REDIS] ERROR: Failed to read value: %v", err)
			return "", err
		}
		log.Printf("[REDIS] Read %d bytes of value data", bytesRead)
		
		reader.ReadString('\n')
		
		log.Printf("[REDIS] GET operation successful, value='%s'", string(value))
		return string(value), nil
	}

	log.Printf("[REDIS] ERROR: Unexpected response format: %s", response)
	return "", fmt.Errorf("unexpected response: %s", response)
}

func (r *RedisClient) Close() error {
	log.Printf("[REDIS] Closing connection to %s...", r.addr)
	if r.conn != nil {
		err := r.conn.Close()
		if err != nil {
			log.Printf("[REDIS] ERROR: Failed to close connection: %v", err)
			return err
		}
		log.Printf("[REDIS] Connection closed successfully")
		return nil
	}
	log.Printf("[REDIS] No active connection to close")
	return nil
}
