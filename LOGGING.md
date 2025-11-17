# Comprehensive Logging Documentation

This application includes **VERY VERBOSE** logging to stdout for all operations.

## Log Format

All logs include:
- Date and time with microseconds
- Source file and line number  
- Component prefix in brackets (e.g., `[REDIS]`, `[POSTGRES]`, `[HTTP]`)

## Key Log Categories

### 1. APPLICATION STARTUP (`[INIT]`)
- Environment variable configuration
- Service initialization
- Connection establishment

### 2. REDIS OPERATIONS (`[REDIS]`)
- TCP connection details
- RESP protocol commands
- Operation latency tracking
- SET/GET operations with byte counts

### 3. POSTGRESQL OPERATIONS (`[POSTGRES]`)
- TCP connection handshake
- Wire protocol messages
- Authentication flow
- Query execution with latency

### 4. HTTP REQUESTS (`[REQUEST:ID]`)
- Unique request ID per request
- Headers and payload logging
- Operation timing
- Success/error responses

### 5. REDIS LOAD TEST (`[LOAD-REDIS]`)
- 5,000 keys × 4,096 chars each
- 5,000 values × 10,000 chars each
- Progress every 100 keys
- Final statistics with throughput

### 6. DATABASE LOAD TEST (`[LOAD-DB]`)
- 200 concurrent connections
- Connection latency per connection
- Progress every 20 connections
- Final statistics with average latency

### 7. ARRAY KEEPER (`[KEEPER]`)
- Runs every 10 seconds
- Logs array sizes and samples
- Prevents GC collection
- Shows total memory held (~68 MB)

### 8. DATABASE CONNECTION KEEPER (`[LOAD-DB-KEEPER]`)
- Runs every 15 seconds
- Logs all 200 connection addresses
- Keeps connections alive permanently

### 9. MEMORY MONITORING (`[USAGE]`)
- Runs every 5 seconds
- Tracks: Alloc, TotalAlloc, Sys, HeapAlloc, HeapInuse
- Tracks: NumGC, NumGoroutine

### 10. METRICS EXPORT (`[METRICS]`)
- Request details
- Export size in bytes
- Counter and gauge counts

## Viewing Logs

```bash
# View all logs
docker-compose logs -f

# View only API logs
docker-compose logs -f api

# View last 100 lines
docker-compose logs --tail=100 api
```

## Log Volume

- **Startup**: 60-70 lines
- **Per HTTP Request**: 10-15 lines
- **Redis Load Test**: 5,000+ lines
- **Database Load Test**: 200+ lines
- **Background processes**: ~220 lines/minute

For detailed examples, see the full documentation in the repository.
