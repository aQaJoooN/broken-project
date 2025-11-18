# Broken Application

A simple Go application with intentional resource usage for testing and monitoring.

## Features

- Redis key-value operations
- PostgreSQL database integration
- Prometheus metrics
- Memory-intensive load tests
- Database connection pooling tests

## Quick Start

```bash
cd broken-project
docker-compose up --build
```

## Access Points

- **UI**: http://localhost:8090
- **API**: http://localhost:8080
- **Metrics**: http://localhost:8080/metrics

## API Endpoints

- `POST /api/set` - Set key-value in Redis
- `GET /api/load` - Load 5,000 keys into Redis (creates ~68 MB in memory)
- `GET /api/load-db` - Open 200 database connections
- `GET /metrics` - Prometheus metrics

## Configuration

To change the API URL in the UI, edit `ui/config.js`:

```javascript
const API_CONFIG = {
    baseUrl: 'http://localhost:8080',  // Change this
    // ...
};
```

## Documentation

- [FEATURES.md](FEATURES.md) - Detailed feature documentation
- [LOGGING.md](LOGGING.md) - Logging documentation
- [ui/README.md](ui/README.md) - UI configuration guide

## Stopping

```bash
docker-compose down

# Remove volumes
docker-compose down -v
```
