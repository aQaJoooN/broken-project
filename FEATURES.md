# Application Features

## API Endpoints

### 1. POST /api/set
Set a single key-value pair in Redis.

**Request:**
```json
{
  "key": "mykey",
  "value": "myvalue"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Key set successfully"
}
```

### 2. GET /api/load
Trigger a Redis load test that creates 5,000 key-value pairs.

**Configuration:**
- Total Keys: 5,000
- Key Length: 4,096 characters
- Value Length: 10,000 characters
- Total Data Size: ~70 MB
- Batch Size: 100 keys per batch

**Response:**
```json
{
  "success": true,
  "message": "Load test started"
}
```

**Note:** The load test runs asynchronously in the background. Check server logs for progress.

**Memory Impact:**
- Keys stored in application array: ~20 MB
- Values stored in application array: ~48 MB
- Total memory held permanently: ~68 MB

### 3. GET /api/load-db
Trigger a database load test that creates 200 concurrent PostgreSQL connections.

**Configuration:**
- Total Connections: 200
- Connection Type: TCP with PostgreSQL wire protocol
- Connections kept alive permanently

**Response:**
```json
{
  "success": true,
  "message": "Database load test started"
}
```

**Note:** All 200 connections remain open and are kept alive to increase database latency.

### 4. GET /metrics
Prometheus-compatible metrics endpoint.

## Metrics Available

### Counters
1. **api_requests_total** - Total number of API requests
   - Labels: method, endpoint, status

2. **redis_operations_total** - Total number of Redis operations
   - Labels: operation, status

3. **redis_load_test_runs_total** - Total Redis load test runs
   - Labels: status

4. **db_load_test_runs_total** - Total database load test runs
   - Labels: status

### Gauges

#### Application Metrics
1. **app_memory_usage_bytes** - Current memory usage in bytes
   - Labels: type (alloc, total_alloc, sys, heap_alloc, heap_inuse)

2. **app_goroutines** - Number of active goroutines

3. **app_gc_runs_total** - Total number of garbage collection runs

4. **app_loaded_keys_count** - Number of keys stored in application memory

5. **app_loaded_values_count** - Number of values stored in application memory

#### Connection Status
6. **redis_connection_status** - Redis connection status (1=connected, 0=disconnected)

7. **postgres_connection_status** - PostgreSQL connection status (1=connected, 0=disconnected)

#### Redis Load Test Metrics
8. **redis_load_test_duration_seconds** - Duration of last Redis load test

9. **redis_load_test_throughput_keys_per_sec** - Throughput in keys/second

10. **redis_load_test_successful_keys** - Number of successfully loaded keys

11. **redis_load_test_failed_keys** - Number of failed keys

12. **redis_load_test_total_bytes** - Total bytes written

#### Database Load Test Metrics
13. **db_load_test_successful_connections** - Successful database connections

14. **db_load_test_failed_connections** - Failed database connections

15. **db_load_test_duration_seconds** - Duration of database load test

16. **db_load_test_average_latency_seconds** - Average connection latency

17. **db_active_connections_count** - Number of active connections being kept alive

#### Latency Metrics
18. **http_request_duration_seconds** - HTTP request duration
    - Labels: endpoint

19. **redis_operation_latency_seconds** - Redis operation latency
    - Labels: operation (set, get)

20. **postgres_operation_latency_seconds** - PostgreSQL operation latency
    - Labels: operation (create_table)

## UI Features

### Set Key-Value Form
- Input fields for key and value
- Submit button to set values in Redis
- Real-time success/error feedback

### Redis Load Test Button
- "Load on Redis" button to trigger load test
- Shows progress information
- Disables during execution to prevent multiple runs
- Creates 5,000 keys (4,096 chars) and values (10,000 chars)
- Stores all data in application memory permanently

### Database Load Test Button
- "Load on DataBase" button to trigger database load test
- Opens 200 concurrent PostgreSQL connections
- Keeps all connections alive permanently
- Increases database latency significantly

### Information Panel
- Links to all API endpoints
- Direct link to metrics endpoint
- All API calls made from client browser (not from UI container)

## Load Test Details

### Redis Load Test (`/api/load`)

1. **Generates Random Data:**
   - Keys: 4,096 characters (alphanumeric)
   - Values: 10,000 characters (alphanumeric)

2. **Batch Processing:**
   - Processes 100 keys per batch
   - Logs progress every 100 keys

3. **Memory Storage:**
   - All keys stored in global array
   - All values stored in global array
   - Arrays kept alive by keeper goroutine (runs every 10 seconds)
   - Prevents garbage collection

4. **Metrics Tracked:**
   - Total keys processed
   - Successful operations
   - Failed operations
   - Total bytes written
   - Duration in seconds
   - Throughput (keys/second)
   - Data rate (MB/second)

### Database Load Test (`/api/load-db`)

1. **Connection Creation:**
   - Opens 200 concurrent TCP connections
   - Performs PostgreSQL wire protocol handshake
   - Authenticates with database

2. **Connection Management:**
   - All connections stored in global array
   - Keeper goroutine runs every 15 seconds
   - Logs connection details (local/remote addresses)
   - Prevents connection closure

3. **Metrics Tracked:**
   - Total connections attempted
   - Successful connections
   - Failed connections
   - Average connection latency
   - Duration in seconds
   - Connection rate (conn/sec)

4. **Impact:**
   - Increases database resource usage
   - Increases connection latency
   - Holds database connections permanently

## Background Processes

### 1. Memory Monitor
- Runs every 5 seconds
- Tracks: Alloc, TotalAlloc, Sys, HeapAlloc, HeapInuse
- Tracks: NumGC, NumGoroutine
- Updates metrics registry

### 2. Array Keeper
- Runs every 10 seconds
- Accesses keys and values arrays
- Logs array sizes and samples
- Uses `runtime.KeepAlive()` to prevent GC
- Ensures ~68 MB stays in memory

### 3. Database Connection Keeper
- Runs every 15 seconds
- Logs all active connections
- Displays local and remote addresses
- Prevents connection closure

## Example Metrics Output

```
# HELP api_requests_total Total number of API requests by method, endpoint and status
# TYPE api_requests_total counter
api_requests_total{endpoint="/api/set",method="POST",status="200"} 15
api_requests_total{endpoint="/api/load",method="GET",status="202"} 1
api_requests_total{endpoint="/api/load-db",method="GET",status="202"} 1

# HELP redis_operations_total Total number of Redis operations by operation type and status
# TYPE redis_operations_total counter
redis_operations_total{operation="set",status="success"} 5015

# HELP app_memory_usage_bytes Current application memory usage in bytes by type
# TYPE app_memory_usage_bytes gauge
app_memory_usage_bytes{type="alloc"} 71303168
app_memory_usage_bytes{type="total_alloc"} 71303168
app_memory_usage_bytes{type="sys"} 75497472
app_memory_usage_bytes{type="heap_alloc"} 71303168
app_memory_usage_bytes{type="heap_inuse"} 72351744

# HELP app_goroutines Number of goroutines
# TYPE app_goroutines gauge
app_goroutines 7

# HELP app_gc_runs_total Total number of garbage collection runs
# TYPE app_gc_runs_total gauge
app_gc_runs_total 3

# HELP redis_connection_status Redis connection status (1=connected, 0=disconnected)
# TYPE redis_connection_status gauge
redis_connection_status 1

# HELP postgres_connection_status PostgreSQL connection status (1=connected, 0=disconnected)
# TYPE postgres_connection_status gauge
postgres_connection_status 1

# HELP redis_load_test_duration_seconds Duration of the last Redis load test in seconds
# TYPE redis_load_test_duration_seconds gauge
redis_load_test_duration_seconds 45.67

# HELP redis_load_test_throughput_keys_per_sec Throughput of the last Redis load test in keys per second
# TYPE redis_load_test_throughput_keys_per_sec gauge
redis_load_test_throughput_keys_per_sec 109.5

# HELP redis_load_test_successful_keys Number of successfully loaded keys in the last Redis load test
# TYPE redis_load_test_successful_keys gauge
redis_load_test_successful_keys 5000

# HELP app_loaded_keys_count Number of keys currently stored in application memory from load test
# TYPE app_loaded_keys_count gauge
app_loaded_keys_count 5000

# HELP app_loaded_values_count Number of values currently stored in application memory from load test
# TYPE app_loaded_values_count gauge
app_loaded_values_count 5000

# HELP db_load_test_successful_connections Number of successful database connections in the last load test
# TYPE db_load_test_successful_connections gauge
db_load_test_successful_connections 200

# HELP db_load_test_duration_seconds Duration of the last database load test in seconds
# TYPE db_load_test_duration_seconds gauge
db_load_test_duration_seconds 2.34

# HELP db_load_test_average_latency_seconds Average connection latency in the last database load test
# TYPE db_load_test_average_latency_seconds gauge
db_load_test_average_latency_seconds 0.0117

# HELP db_active_connections_count Number of active database connections being kept alive
# TYPE db_active_connections_count gauge
db_active_connections_count 200

# HELP redis_operation_latency_seconds Redis operation latency in seconds by operation type
# TYPE redis_operation_latency_seconds gauge
redis_operation_latency_seconds{operation="set"} 0.0012

# HELP postgres_operation_latency_seconds PostgreSQL operation latency in seconds by operation type
# TYPE postgres_operation_latency_seconds gauge
postgres_operation_latency_seconds{operation="create_table"} 0.0056

# HELP http_request_duration_seconds HTTP request duration in seconds by endpoint
# TYPE http_request_duration_seconds gauge
http_request_duration_seconds{endpoint="/api/set"} 0.0023
```

## Usage

### Start the Application
```bash
cd broken-project
docker-compose up --build
```

### Access Points
- UI: http://localhost:3000
- API: http://localhost:8080
- Metrics: http://localhost:8080/metrics

### Run Redis Load Test from UI
1. Open http://localhost:3000 in your browser
2. Scroll to "Redis Load Test" section
3. Click "Load on Redis" button
4. Monitor server logs: `docker-compose logs -f api`
5. Memory will increase by ~68 MB and stay there

### Run Database Load Test from UI
1. Open http://localhost:3000 in your browser
2. Scroll to "Database Load Test" section
3. Click "Load on DataBase" button
4. Monitor server logs: `docker-compose logs -f api`
5. 200 connections will open and remain active

### Run Load Tests from Command Line
```bash
# Redis load test
curl -X GET http://localhost:8080/api/load

# Database load test
curl -X GET http://localhost:8080/api/load-db
```

### View Metrics
```bash
curl http://localhost:8080/metrics
```

## Performance Expectations

### Redis Load Test
- **Expected Duration:** 30-60 seconds
- **Expected Throughput:** 80-150 keys/second
- **Memory Usage:** ~68 MB increase (permanent)
- **Network Traffic:** ~70 MB written to Redis

### Database Load Test
- **Expected Duration:** 2-5 seconds
- **Expected Connections:** 200 successful
- **Average Latency:** 10-20ms per connection
- **Impact:** Significant increase in database resource usage

Monitor the logs for detailed progress information.
