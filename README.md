# Sharded In-Memory KV Store with TTL

A production-quality, concurrent, sharded in-memory key-value store built with Go.

## Features

- **Sharded Architecture**: Reduces lock contention using multiple shards (configurable).
- **TTL Support**: Hybrid strategy with lazy deletion and background cleanup worker.
- **RESTful HTTP API**: Easy integration via standard HTTP methods.
- **Metrics**: Prometheus-compatible endpoints.
- **Graceful Shutdown**: Handles SIGINT/SIGTERM properly.
- **Structured Logging**: JSON logs using `slog`.

## How to Run

### Local (requires Go 1.21+)

```bash
go run cmd/server/main.go
```

### Docker

```bash
docker-compose up --build
```

## Configuration

Environment variables:
- `PORT`: Server port (default: `8080`)
- `CLEANUP_INTERVAL`: Interval for background cleanup (default: `1m`)
- `SHARD_COUNT`: Number of shards for the store (default: `16`)

## API Examples

### Set a key
```bash
curl -X PUT http://localhost:8080/kv/foo \
  -H "Content-Type: application/json" \
  -d '{"value": "bar", "ttl_seconds": 60}'
```

### Get a key
```bash
curl http://localhost:8080/kv/foo
```

### Delete a key
```bash
curl -X DELETE http://localhost:8080/kv/foo
```

### Stats
```bash
curl http://localhost:8080/stats
```

### Metrics (Prometheus)
```bash
curl http://localhost:8080/metrics
```

## Design Decisions

- **Sharding**: Using `sync.RWMutex` per shard instead of a single global lock significantly improves throughput under heavy concurrent load.
- **TTL Strategy**:
    - **Lazy Deletion**: Keys are checked for expiry during `Get`. This ensures expired keys are never returned.
    - **Background Worker**: Periodically scans all shards to remove expired keys that haven't been accessed. This prevents memory leaks from "dark" keys.
- **Chi Router**: Chosen for its lightweight nature, standard library compatibility, and great middleware support.
- **Prometheus**: Industry standard for metrics collection and monitoring.

## Testing

Run all tests including race detector:
```bash
go test -race ./...
```

Run benchmarks:
```bash
go test -bench=. ./internal/store/...
```
