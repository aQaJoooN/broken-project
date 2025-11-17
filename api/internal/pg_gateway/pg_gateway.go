package pg_gateway

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
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
	log.Printf("[POSTGRES] Creating new PostgreSQL client")
	log.Printf("[POSTGRES] Configuration: host=%s, port=%s, user=%s, db=%s", host, port, user, dbname)
	
	client := &PGClient{
		host: host,
		port: port,
		user: user,
		pass: password,
		db:   dbname,
	}

	log.Printf("[POSTGRES] Initiating connection...")
	if err := client.connect(); err != nil {
		log.Printf("[POSTGRES] FATAL: Failed to connect: %v", err)
		panic(fmt.Sprintf("Failed to connect to PostgreSQL: %v", err))
	}

	log.Printf("[POSTGRES] Client created successfully")
	return client
}

func (p *PGClient) connect() error {
	addr := p.host + ":" + p.port
	log.Printf("[POSTGRES] Dialing TCP connection to %s...", addr)
	startDial := time.Now()
	
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("[POSTGRES] ERROR: Failed to dial: %v", err)
		return err
	}
	p.conn = conn
	log.Printf("[POSTGRES] TCP connection established in %v", time.Since(startDial))
	log.Printf("[POSTGRES] Local address: %s", conn.LocalAddr())
	log.Printf("[POSTGRES] Remote address: %s", conn.RemoteAddr())

	log.Printf("[POSTGRES] Building startup message...")
	startupMsg := p.buildStartupMessage()
	log.Printf("[POSTGRES] Startup message size: %d bytes", len(startupMsg))
	log.Printf("[POSTGRES] Sending startup message...")
	
	bytesWritten, err := p.conn.Write(startupMsg)
	if err != nil {
		log.Printf("[POSTGRES] ERROR: Failed to send startup message: %v", err)
		return err
	}
	log.Printf("[POSTGRES] Sent %d bytes", bytesWritten)

	log.Printf("[POSTGRES] Reading authentication response...")
	reader := bufio.NewReader(p.conn)
	msgCount := 0
	
	for {
		msgType := make([]byte, 1)
		_, err := reader.Read(msgType)
		if err != nil {
			log.Printf("[POSTGRES] ERROR: Failed to read message type: %v", err)
			return err
		}
		msgCount++
		log.Printf("[POSTGRES] Message #%d: type='%c' (0x%02x)", msgCount, msgType[0], msgType[0])

		if msgType[0] == 'E' {
			log.Printf("[POSTGRES] Received error message")
			length := make([]byte, 4)
			reader.Read(length)
			msgLen := int(length[0])<<24 | int(length[1])<<16 | int(length[2])<<8 | int(length[3])
			if msgLen > 4 {
				payload := make([]byte, msgLen-4)
				reader.Read(payload)
				log.Printf("[POSTGRES] ERROR: %s", string(payload))
			}
			continue
		}

		if msgType[0] == 'R' {
			length := make([]byte, 4)
			bytesRead, err := reader.Read(length)
			if err != nil {
				log.Printf("[POSTGRES] ERROR: Failed to read auth length: %v", err)
				return err
			}
			log.Printf("[POSTGRES] Read %d bytes for auth length", bytesRead)
			
			msgLen := int(length[0])<<24 | int(length[1])<<16 | int(length[2])<<8 | int(length[3])
			log.Printf("[POSTGRES] Auth message length: %d bytes", msgLen)
			
			if msgLen > 4 {
				authData := make([]byte, msgLen-4)
				bytesRead, err := reader.Read(authData)
				if err != nil {
					log.Printf("[POSTGRES] ERROR: Failed to read auth data: %v", err)
					return err
				}
				log.Printf("[POSTGRES] Read %d bytes of auth data", bytesRead)
				
				if len(authData) >= 4 {
					authType := int(authData[0])<<24 | int(authData[1])<<16 | int(authData[2])<<8 | int(authData[3])
					log.Printf("[POSTGRES] Authentication type: %d", authType)
					
					if authType == 3 {
						log.Printf("[POSTGRES] Clear text password authentication required")
						passwordMsg := p.buildPasswordMessage()
						log.Printf("[POSTGRES] Sending password message (%d bytes)...", len(passwordMsg))
						_, err = p.conn.Write(passwordMsg)
						if err != nil {
							log.Printf("[POSTGRES] ERROR: Failed to send password: %v", err)
							return err
						}
						log.Printf("[POSTGRES] Password sent successfully")
					} else if authType == 0 {
						log.Printf("[POSTGRES] Authentication successful (trust mode)")
					}
				}
			}
			continue
		}

		if msgType[0] == 'S' || msgType[0] == 'K' {
			log.Printf("[POSTGRES] Received backend parameter/key data message")
			length := make([]byte, 4)
			reader.Read(length)
			msgLen := int(length[0])<<24 | int(length[1])<<16 | int(length[2])<<8 | int(length[3])
			if msgLen > 4 {
				payload := make([]byte, msgLen-4)
				reader.Read(payload)
				log.Printf("[POSTGRES] Read %d bytes of parameter data", len(payload))
			}
			continue
		}

		if msgType[0] == 'Z' {
			log.Printf("[POSTGRES] Received ReadyForQuery message - connection established")
			length := make([]byte, 4)
			reader.Read(length)
			msgLen := int(length[0])<<24 | int(length[1])<<16 | int(length[2])<<8 | int(length[3])
			if msgLen > 4 {
				payload := make([]byte, msgLen-4)
				reader.Read(payload)
			}
			break
		}

		length := make([]byte, 4)
		reader.Read(length)
		msgLen := int(length[0])<<24 | int(length[1])<<16 | int(length[2])<<8 | int(length[3])
		log.Printf("[POSTGRES] Message length: %d bytes", msgLen)
		
		if msgLen > 4 {
			payload := make([]byte, msgLen-4)
			bytesRead, _ := reader.Read(payload)
			log.Printf("[POSTGRES] Read %d bytes of payload", bytesRead)
		}
	}

	log.Printf("[POSTGRES] Connection handshake completed successfully")
	return nil
}

func (p *PGClient) buildStartupMessage() []byte {
	log.Printf("[POSTGRES] Building startup message with user='%s', database='%s'", p.user, p.db)
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
	
	log.Printf("[POSTGRES] Startup message built: %d bytes total", length)
	return msg
}

func (p *PGClient) buildPasswordMessage() []byte {
	log.Printf("[POSTGRES] Building password message")
	password := p.pass + "\x00"
	length := len(password) + 5
	
	msg := make([]byte, length)
	msg[0] = 'p'
	msg[1] = byte(length >> 24)
	msg[2] = byte(length >> 16)
	msg[3] = byte(length >> 8)
	msg[4] = byte(length)
	copy(msg[5:], password)
	
	log.Printf("[POSTGRES] Password message built: %d bytes total", length)
	return msg
}

func (p *PGClient) CreateTable() error {
	query := `CREATE TABLE IF NOT EXISTS app_data (
		id SERIAL PRIMARY KEY,
		key VARCHAR(255) UNIQUE NOT NULL,
		value TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	log.Printf("[POSTGRES] Executing CREATE TABLE query...")
	log.Printf("[POSTGRES] Query: %s", query)
	return p.executeQuery(query)
}

func (p *PGClient) executeQuery(query string) error {
	log.Printf("[POSTGRES] Building query message...")
	queryMsg := p.buildQueryMessage(query)
	log.Printf("[POSTGRES] Query message size: %d bytes", len(queryMsg))
	log.Printf("[POSTGRES] Sending query to PostgreSQL...")
	
	startWrite := time.Now()
	bytesWritten, err := p.conn.Write(queryMsg)
	if err != nil {
		log.Printf("[POSTGRES] ERROR: Failed to write query: %v", err)
		return err
	}
	log.Printf("[POSTGRES] Wrote %d bytes in %v", bytesWritten, time.Since(startWrite))

	log.Printf("[POSTGRES] Reading query response...")
	reader := bufio.NewReader(p.conn)
	msgCount := 0
	
	for {
		msgType := make([]byte, 1)
		_, err := reader.Read(msgType)
		if err != nil {
			log.Printf("[POSTGRES] ERROR: Failed to read message type: %v", err)
			return err
		}
		msgCount++
		log.Printf("[POSTGRES] Response message #%d: type='%c' (0x%02x)", msgCount, msgType[0], msgType[0])

		length := make([]byte, 4)
		reader.Read(length)
		msgLen := int(length[0])<<24 | int(length[1])<<16 | int(length[2])<<8 | int(length[3])
		log.Printf("[POSTGRES] Message length: %d bytes", msgLen)
		
		if msgLen > 4 {
			payload := make([]byte, msgLen-4)
			bytesRead, _ := reader.Read(payload)
			log.Printf("[POSTGRES] Read %d bytes of payload", bytesRead)
		}

		if msgType[0] == 'Z' {
			log.Printf("[POSTGRES] Query completed successfully")
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
	log.Printf("[POSTGRES] Closing connection...")
	if p.conn != nil {
		log.Printf("[POSTGRES] Sending termination message...")
		terminateMsg := []byte{'X', 0x00, 0x00, 0x00, 0x04}
		p.conn.Write(terminateMsg)
		
		err := p.conn.Close()
		if err != nil {
			log.Printf("[POSTGRES] ERROR: Failed to close connection: %v", err)
			return err
		}
		log.Printf("[POSTGRES] Connection closed successfully")
		return nil
	}
	log.Printf("[POSTGRES] No active connection to close")
	return nil
}
