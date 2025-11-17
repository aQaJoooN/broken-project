package redis_gateway

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type RedisClient struct {
	conn net.Conn
	addr string
}

func NewRedisClient(addr string) *RedisClient {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	return &RedisClient{
		conn: conn,
		addr: addr,
	}
}

func (r *RedisClient) Set(key, value string) error {
	cmd := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
		len(key), key, len(value), value)
	
	_, err := r.conn.Write([]byte(cmd))
	if err != nil {
		return err
	}

	reader := bufio.NewReader(r.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	if !strings.HasPrefix(response, "+OK") {
		return fmt.Errorf("Redis error: %s", response)
	}

	return nil
}

func (r *RedisClient) Get(key string) (string, error) {
	cmd := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key)
	
	_, err := r.conn.Write([]byte(cmd))
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(r.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(response, "$-1") {
		return "", fmt.Errorf("key not found")
	}

	if strings.HasPrefix(response, "$") {
		lengthStr := strings.TrimSpace(response[1:])
		length, _ := strconv.Atoi(lengthStr)
		
		value := make([]byte, length)
		_, err = reader.Read(value)
		if err != nil {
			return "", err
		}
		
		reader.ReadString('\n')
		
		return string(value), nil
	}

	return "", fmt.Errorf("unexpected response: %s", response)
}

func (r *RedisClient) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
