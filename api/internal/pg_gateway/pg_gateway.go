package pg_gateway

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type PGClient struct {
	conn net.Conn
	host string
	port string
	user string
	pass string
	db   string
}

func NewPGClient(host, port, user, password, dbname string) *PGClient {
	client := &PGClient{
		host: host,
		port: port,
		user: user,
		pass: password,
		db:   dbname,
	}

	if err := client.connect(); err != nil {
		panic(fmt.Sprintf("Failed to connect to PostgreSQL: %v", err))
	}

	return client
}

func (p *PGClient) connect() error {
	conn, err := net.Dial("tcp", p.host+":"+p.port)
	if err != nil {
		return err
	}
	p.conn = conn

	startupMsg := p.buildStartupMessage()
	_, err = p.conn.Write(startupMsg)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(p.conn)
	for {
		msgType := make([]byte, 1)
		_, err := reader.Read(msgType)
		if err != nil {
			return err
		}

		if msgType[0] == 'R' {
			authData := make([]byte, 4)
			reader.Read(authData)
			continue
		}

		if msgType[0] == 'Z' {
			break
		}

		length := make([]byte, 4)
		reader.Read(length)
		msgLen := int(length[0])<<24 | int(length[1])<<16 | int(length[2])<<8 | int(length[3])
		if msgLen > 4 {
			payload := make([]byte, msgLen-4)
			reader.Read(payload)
		}
	}

	return nil
}

func (p *PGClient) buildStartupMessage() []byte {
	params := fmt.Sprintf("user\x00%s\x00database\x00%s\x00\x00", p.user, p.db)
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

func (p *PGClient) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS app_data (
		id SERIAL PRIMARY KEY,
		key VARCHAR(255) UNIQUE NOT NULL,
		value TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	return p.executeQuery(query)
}

func (p *PGClient) executeQuery(query string) error {
	queryMsg := p.buildQueryMessage(query)
	_, err := p.conn.Write(queryMsg)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(p.conn)
	for {
		msgType := make([]byte, 1)
		_, err := reader.Read(msgType)
		if err != nil {
			return err
		}

		length := make([]byte, 4)
		reader.Read(length)
		msgLen := int(length[0])<<24 | int(length[1])<<16 | int(length[2])<<8 | int(length[3])
		
		if msgLen > 4 {
			payload := make([]byte, msgLen-4)
			reader.Read(payload)
		}

		if msgType[0] == 'Z' {
			break
		}
	}

	return nil
}

func (p *PGClient) buildQueryMessage(query string) []byte {
	length := len(query) + 5
	msg := make([]byte, length+1)
	msg[0] = 'Q'
	msg[1] = byte(length >> 24)
	msg[2] = byte(length >> 16)
	msg[3] = byte(length >> 8)
	msg[4] = byte(length)
	copy(msg[5:], query)
	msg[length] = 0x00
	
	return msg
}

func (p *PGClient) Close() error {
	if p.conn != nil {
		terminateMsg := []byte{'X', 0x00, 0x00, 0x00, 0x04}
		p.conn.Write(terminateMsg)
		return p.conn.Close()
	}
	return nil
}
