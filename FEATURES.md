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

### 2. POST /api/load
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

### 3. GET /metrics
Prometheus-compatible metrics endpoint.

## Metrics Available

### Counters
1. **api_requests_total** - Total number of API requests
   - Labels: method, endpoint, status

2. **redis_operations_total** - Total number of Redis operations
   - Labels: operation, status

3. **redis_load_test_keys_total** - Total keys loaded in Redis load test
   - Labels: status

### Gauges
1. **app_memory_usage_bytes** - Current memory usage in bytes
   - Labels: type (alloc, total_alloc, sys)

2. **app_goroutines** - Number of active goroutines

3. **redis_connection_status** - Redis connection status (1=connected, 0=disconnected)

4. **postgres_connection_status** - PostgreSQL connection status (1=connected, 0=disconnected)

5. **redis_load_test_duration_seconds** - Duration of last Redis load test

6. **redis_load_test_throughput** - Throughput of last Redis load test (keys/second)

7. **http_request_duration_seconds** - HTTP request duration
   - Labels: endpoint

## UI Features

### Set Key-Value Form
- Input fields for key and value
- Submit button to set values in Redis
- Real-time success/error feedback

### Redis Load Test Button
- "Load on Redis" button to trigger load test
- Shows progress information
- Disables during execution to prevent multiple runs
- Displays test configuration:
  - 5,000 keys
  - 4,096 character keys
  - 10,000 character values
  - ~70 MB total data

### Information Panel
- Links to all API endpoints
- Direct link to metrics endpoint
- All API calls made from client browser (not from UI container)

## Load Test Details

The load test (`/api/load` endpoint):

1. **Generates Random Data:**
   - Keys: 4,096 characters (alphanumeric)
   - Values: 10,000 characters (alphanumeric)

2. **Batch Processing:**
   - Processes 100 keys per batch
   - Logs progress every 100 keys

3. **Metrics Tracked:**
   - Total keys processed
   - Successful operations
   - Failed operations
   - Total bytes written
   - Duration in seconds
   - Throughput (keys/second)
   - Data rate (MB/second)

4. **Verbose Logging:**
   - Each key set operation is logged
   - Progress updates every 100 keys
   - ETA calculation
   - Final statistics summary

## Example Metrics Output

```
# HELP api_requests_total Total number of requests
# TYPE api_requests_total counter
api_requests_total{endpoint="/api/set",method="POST",status="200"} 15

# HELP redis_operations_total Total number of Redis operations
# TYPE redis_operations_total counter
redis_operations_total{operation="set",status="success"} 5015

# HELP app_memory_usage_bytes Current memory usage in bytes
# TYPE app_memory_usage_bytes gauge
app_memory_usage_bytes{type="alloc"} 8388608
app_memory_usage_bytes{type="total_alloc"} 16777216
app_memory_usage_bytes{type="sys"} 25165824

# HELP app_goroutines Number of goroutines
# TYPE app_goroutines gauge
app_goroutines 5

# HELP redis_connection_status Redis connection status (1=connected, 0=disconnected)
# TYPE redis_connection_status gauge
redis_connection_status 1

# HELP postgres_connection_status PostgreSQL connection status (1=connected, 0=disconnected)
# TYPE postgres_connection_status gauge
postgres_connection_status 1

# HELP redis_load_test_duration_seconds Duration of last Redis load test in seconds
# TYPE redis_load_test_duration_seconds gauge
redis_load_test_duration_seconds 45.67

# HELP redis_load_test_throughput Throughput of last Redis load test in keys per second
# TYPE redis_load_test_throughput gauge
redis_load_test_throughput 109.5

# HELP http_request_duration_seconds HTTP request duration in seconds
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

### Run Load Test from UI
1. Open http://localhost:3000 in your browser
2. Scroll to "Redis Load Test" section
3. Click "Load on Redis" button
4. Monitor server logs: `docker-compose logs -f api`

### Run Load Test from Command Line
```bash
curl -X POST http://localhost:8080/api/load
```

### View Metrics
```bash
curl http://localhost:8080/metrics
```

## Performance Expectations

Load test performance depends on system resources:
- **Expected Duration:** 30-60 seconds
- **Expected Throughput:** 80-150 keys/second
- **Memory Usage:** ~100-200 MB during load test
- **Network Traffic:** ~70 MB written to Redis

Monitor the logs for detailed progress information.
