# Comprehensive Logging Documentation

This application includes **VERY VERBOSE** logging to stdout for all operations.

## Log Format

All logs include:
- Date and time with microseconds
- Source file and line number
- Component prefix in brackets (e.g., `[REDIS]`, `[POSTGRES]`, `[HTTP]`)

## Log Categories

### 1. APPLICATION STARTUP (`[INIT]`)
```
[INIT] Reading environment variables...
[INIT] Redis configuration: host=redis, port=6379
[INIT] PostgreSQL configuration: host=postgres, port=5432, user=appuser, db=appdb
[INIT] Initializing metrics registry...
[INIT] Metrics registry initialized successfully
```

### 2. REDIS OPERATIONS (`[REDIS]`)

**Connection:**
```
[REDIS] Dialing TCP connection to redis:6379...
[REDIS] TCP connection established in 2.5ms
[REDIS] Local address: 172.18.0.4:45678
[REDIS] Remote address: 172.18.0.2:6379
```

**SET Command:**
```
[REDIS] Building RESP command for SET key='mykey' value='myvalue'
[REDIS] RESP command: "*3\r\n$3\r\nSET\r\n$5\r\nmykey\r\n$7\r\nmyvalue\r\n"
[REDIS] Command size: 38 bytes
[REDIS] Writing command to socket...
[REDIS] Wrote 38 bytes in 125µs
[REDIS] Reading response from Redis...
[REDIS] Received response in 1.2ms: "+OK\r\n"
[REDIS] SET operation successful
```

**GET Command:**
```
[REDIS] Building RESP command for GET key='mykey'
[REDIS] RESP command: "*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n"
[REDIS] Writing GET command to socket...
[REDIS] Wrote 24 bytes
[REDIS] Reading GET response...
[REDIS] Received response: "$7\r\n"
[REDIS] Value length: 7 bytes
[REDIS] Read 7 bytes of value data
[REDIS] GET operation successful, value='myvalue'
```

**Close:**
```
[REDIS] Closing connection to redis:6379...
[REDIS] Connection closed successfully
```

### 3. POSTGRESQL OPERATIONS (`[POSTGRES]`)

**Connection:**
```
[POSTGRES] Creating new PostgreSQL client
[POSTGRES] Configuration: host=postgres, port=5432, user=appuser, db=appdb
[POSTGRES] Initiating connection...
[POSTGRES] Dialing TCP connection to postgres:5432...
[POSTGRES] TCP connection established in 5.3ms
[POSTGRES] Local address: 172.18.0.4:56789
[POSTGRES] Remote address: 172.18.0.3:5432
[POSTGRES] Building startup message...
[POSTGRES] Startup message size: 45 bytes
[POSTGRES] Sending startup message...
[POSTGRES] Sent 45 bytes
[POSTGRES] Reading authentication response...
[POSTGRES] Message #1: type='R' (0x52)
[POSTGRES] Received authentication request
[POSTGRES] Authentication type: 0
[POSTGRES] Message #2: type='Z' (0x5A)
[POSTGRES] Received ReadyForQuery message - connection established
[POSTGRES] Connection handshake completed successfully
[POSTGRES] Client created successfully
```

**CREATE TABLE:**
```
[POSTGRES] Executing CREATE TABLE query...
[POSTGRES] Query: CREATE TABLE IF NOT EXISTS app_data (...)
[POSTGRES] Building query message...
[POSTGRES] Query message size: 156 bytes
[POSTGRES] Sending query to PostgreSQL...
[POSTGRES] Wrote 156 bytes in 234µs
[POSTGRES] Reading query response...
[POSTGRES] Response message #1: type='C' (0x43)
[POSTGRES] Message length: 15 bytes
[POSTGRES] Read 11 bytes of payload
[POSTGRES] Response message #2: type='Z' (0x5A)
[POSTGRES] Query completed successfully
```

**Close:**
```
[POSTGRES] Closing connection...
[POSTGRES] Sending termination message...
[POSTGRES] Connection closed successfully
```

### 4. HTTP REQUESTS (`[REQUEST:ID]`)

Each request gets a unique ID based on timestamp.

**Successful Request:**
```
[REQUEST:1700251234567890] Incoming POST request to /api/set from 172.18.0.5:45678
[REQUEST:1700251234567890] Headers: map[Accept:[*/*] Content-Type:[application/json]]
[REQUEST:1700251234567890] Decoded payload: key='testkey', value='testvalue'
[REQUEST:1700251234567890] Sending SET command to Redis...
[REQUEST:1700251234567890] Redis SET completed in 1.5ms
[REQUEST:1700251234567890] SUCCESS: Key 'testkey' set successfully
[REQUEST:1700251234567890] Sending response to client
[REQUEST:1700251234567890] Request completed successfully
```

**Failed Request (Method Not Allowed):**
```
[REQUEST:1700251234567891] Incoming GET request to /api/set from 172.18.0.5:45679
[REQUEST:1700251234567891] Headers: map[Accept:[*/*]]
[REQUEST:1700251234567891] ERROR: Method not allowed: GET
```

**Failed Request (Invalid JSON):**
```
[REQUEST:1700251234567892] Incoming POST request to /api/set from 172.18.0.5:45680
[REQUEST:1700251234567892] Headers: map[Content-Type:[application/json]]
[REQUEST:1700251234567892] ERROR: Failed to decode JSON body: invalid character 'x' looking for beginning of value
```

### 5. METRICS (`[METRICS]`)

**Counter Operations:**
```
[METRICS] Incrementing counter 'api_requests_total' with labels map[endpoint:/api/set method:POST status:200]
[METRICS] Counter 'api_requests_total' incremented to 5
```

**Gauge Operations:**
```
[METRICS] Setting gauge 'app_memory_usage_bytes' to 8388608.00 with labels map[type:alloc]
[METRICS] New gauge 'app_memory_usage_bytes' created with value 8388608.00
```

**Export:**
```
[METRICS:1700251234567893] Incoming request from 172.18.0.5:45681
[METRICS:1700251234567893] Exporting metrics...
[METRICS] Exporting metrics in Prometheus format
[METRICS] Exported 3 counters and 3 gauges
[METRICS:1700251234567893] Metrics exported, size: 456 bytes
[METRICS:1700251234567893] Metrics sent to client
```

### 6. MEMORY MONITORING (`[USAGE]`)

**Startup:**
```
[USAGE] Memory monitor goroutine started
[USAGE] Monitoring interval: 5 seconds
```

**Periodic Updates (every 5 seconds):**
```
[USAGE] Memory check iteration #1
[USAGE] Memory stats - Alloc: 2097152 bytes (2.00 MB)
[USAGE] Memory stats - TotalAlloc: 4194304 bytes (4.00 MB)
[USAGE] Memory stats - Sys: 8388608 bytes (8.00 MB)
[USAGE] Memory stats - NumGC: 2
[USAGE] Memory stats - NumGoroutine: 3
[USAGE] Metrics updated successfully
```

## Complete Startup Sequence Example

```
2025/11/17 18:30:45.123456 main.go:30: ========================================
2025/11/17 18:30:45.123457 main.go:31: APPLICATION STARTUP INITIATED
2025/11/17 18:30:45.123458 main.go:32: ========================================
2025/11/17 18:30:45.123459 main.go:34: [INIT] Reading environment variables...
2025/11/17 18:30:45.123460 main.go:37: [INIT] Redis configuration: host=redis, port=6379
2025/11/17 18:30:45.123461 main.go:44: [INIT] PostgreSQL configuration: host=postgres, port=5432, user=appuser, db=appdb
2025/11/17 18:30:45.123462 main.go:46: [INIT] Initializing metrics registry...
2025/11/17 18:30:45.123463 metrics.go:28: [METRICS] Creating new metrics registry
2025/11/17 18:30:45.123464 main.go:48: [INIT] Metrics registry initialized successfully
2025/11/17 18:30:45.123465 main.go:50: [REDIS] Attempting to connect to Redis at redis:6379...
2025/11/17 18:30:45.123466 redis_gateway.go:18: [REDIS] Dialing TCP connection to redis:6379...
2025/11/17 18:30:45.125789 redis_gateway.go:27: [REDIS] TCP connection established in 2.323ms
2025/11/17 18:30:45.125790 redis_gateway.go:28: [REDIS] Local address: 172.18.0.4:45678
2025/11/17 18:30:45.125791 redis_gateway.go:29: [REDIS] Remote address: 172.18.0.2:6379
2025/11/17 18:30:45.125792 main.go:54: [REDIS] Connected successfully in 2.327ms
2025/11/17 18:30:45.125793 main.go:56: [POSTGRES] Attempting to connect to PostgreSQL at postgres:5432...
2025/11/17 18:30:45.125794 pg_gateway.go:22: [POSTGRES] Creating new PostgreSQL client
2025/11/17 18:30:45.125795 pg_gateway.go:23: [POSTGRES] Configuration: host=postgres, port=5432, user=appuser, db=appdb
2025/11/17 18:30:45.125796 pg_gateway.go:33: [POSTGRES] Initiating connection...
2025/11/17 18:30:45.125797 pg_gateway.go:44: [POSTGRES] Dialing TCP connection to postgres:5432...
2025/11/17 18:30:45.131234 pg_gateway.go:53: [POSTGRES] TCP connection established in 5.437ms
2025/11/17 18:30:45.131235 pg_gateway.go:54: [POSTGRES] Local address: 172.18.0.4:56789
2025/11/17 18:30:45.131236 pg_gateway.go:55: [POSTGRES] Remote address: 172.18.0.3:5432
2025/11/17 18:30:45.131237 pg_gateway.go:57: [POSTGRES] Building startup message...
2025/11/17 18:30:45.131238 pg_gateway.go:59: [POSTGRES] Startup message size: 45 bytes
2025/11/17 18:30:45.131239 pg_gateway.go:60: [POSTGRES] Sending startup message...
2025/11/17 18:30:45.131456 pg_gateway.go:66: [POSTGRES] Sent 45 bytes
2025/11/17 18:30:45.131457 pg_gateway.go:68: [POSTGRES] Reading authentication response...
2025/11/17 18:30:45.132789 pg_gateway.go:78: [POSTGRES] Message #1: type='R' (0x52)
2025/11/17 18:30:45.132790 pg_gateway.go:80: [POSTGRES] Received authentication request
2025/11/17 18:30:45.132791 pg_gateway.go:83: [POSTGRES] Authentication type: 0
2025/11/17 18:30:45.133123 pg_gateway.go:78: [POSTGRES] Message #2: type='Z' (0x5A)
2025/11/17 18:30:45.133124 pg_gateway.go:87: [POSTGRES] Received ReadyForQuery message - connection established
2025/11/17 18:30:45.133125 pg_gateway.go:102: [POSTGRES] Connection handshake completed successfully
2025/11/17 18:30:45.133126 pg_gateway.go:40: [POSTGRES] Client created successfully
2025/11/17 18:30:45.133127 main.go:61: [POSTGRES] Connected successfully in 7.333ms
2025/11/17 18:30:45.133128 main.go:63: [POSTGRES] Creating database table if not exists...
2025/11/17 18:30:45.133129 pg_gateway.go:127: [POSTGRES] Executing CREATE TABLE query...
2025/11/17 18:30:45.133130 pg_gateway.go:133: [POSTGRES] Query: CREATE TABLE IF NOT EXISTS app_data (...)
2025/11/17 18:30:45.133131 pg_gateway.go:137: [POSTGRES] Building query message...
2025/11/17 18:30:45.133132 pg_gateway.go:139: [POSTGRES] Query message size: 156 bytes
2025/11/17 18:30:45.133133 pg_gateway.go:140: [POSTGRES] Sending query to PostgreSQL...
2025/11/17 18:30:45.134567 pg_gateway.go:148: [POSTGRES] Wrote 156 bytes in 1.434ms
2025/11/17 18:30:45.134568 pg_gateway.go:150: [POSTGRES] Reading query response...
2025/11/17 18:30:45.136789 pg_gateway.go:159: [POSTGRES] Response message #1: type='C' (0x43)
2025/11/17 18:30:45.136790 pg_gateway.go:163: [POSTGRES] Message length: 15 bytes
2025/11/17 18:30:45.136791 pg_gateway.go:167: [POSTGRES] Read 11 bytes of payload
2025/11/17 18:30:45.136792 pg_gateway.go:159: [POSTGRES] Response message #2: type='Z' (0x5A)
2025/11/17 18:30:45.136793 pg_gateway.go:171: [POSTGRES] Query completed successfully
2025/11/17 18:30:45.136794 main.go:67: [POSTGRES] Table created/verified successfully
2025/11/17 18:30:45.136795 main.go:70: [MONITOR] Starting memory monitoring goroutine...
2025/11/17 18:30:45.136796 usage.go:11: [USAGE] Memory monitor goroutine started
2025/11/17 18:30:45.136797 usage.go:12: [USAGE] Monitoring interval: 5 seconds
2025/11/17 18:30:45.136798 main.go:72: [MONITOR] Memory monitoring started
2025/11/17 18:30:45.136799 main.go:74: [HTTP] Registering /api/set endpoint...
2025/11/17 18:30:45.136800 main.go:120: [HTTP] /api/set endpoint registered
2025/11/17 18:30:45.136801 main.go:122: [HTTP] Registering /metrics endpoint...
2025/11/17 18:30:45.136802 main.go:135: [HTTP] /metrics endpoint registered
2025/11/17 18:30:45.136803 main.go:137: ========================================
2025/11/17 18:30:45.136804 main.go:138: API SERVER READY
2025/11/17 18:30:45.136805 main.go:139: Listening on :8080
2025/11/17 18:30:45.136806 main.go:140: Endpoints:
2025/11/17 18:30:45.136807 main.go:141:   - POST /api/set
2025/11/17 18:30:45.136808 main.go:142:   - GET  /metrics
2025/11/17 18:30:45.136809 main.go:143: ========================================
```

## Viewing Logs

### Docker Compose
```bash
# View all logs
docker-compose logs -f

# View only API logs
docker-compose logs -f api

# View last 100 lines
docker-compose logs --tail=100 api
```

### Direct Binary
```bash
cd broken-project/api
./api-server.exe
```

All logs go to stdout, so you can redirect them:
```bash
./api-server.exe > app.log 2>&1
./api-server.exe | tee app.log
```

## Log Volume

Expect approximately:
- **Startup**: 50-60 log lines
- **Per HTTP Request**: 10-15 log lines
- **Per Memory Check** (every 5s): 8 log lines
- **Per Metrics Export**: 5-10 log lines

For a busy system handling 100 requests/minute, expect ~1500 log lines per minute.
