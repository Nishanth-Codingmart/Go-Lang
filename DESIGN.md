# Design Document: Sharded KV Store

## Architecture Overview

```ascii
      +------------------------------------------+
      |               HTTP API                   |
      |   (Chi Router, Handlers, Middlewares)    |
      +-------------------+----------------------+
                          |
      +-------------------v----------------------+
      |              Store Interface             |
      +-------------------+----------------------+
                          |
      +-------------------v----------------------+
      |            Sharded Store                 |
      |  +---------+  +---------+     +---------+|
      |  | Shard 0 |  | Shard 1 | ... | Shard N ||
      |  | (Mutex) |  | (Mutex) |     | (Mutex) ||
      |  +---------+  +---------+     +---------+|
      +---------^------------------^-------------+
                |                  |
      +---------+---------+  +-----+--------------+
      | Background Cleanup|  | Prometheus Metrics |
      |      Worker       |  |      Registry      |
      +-------------------+  +--------------------+
```

## Data Structures

### Item
Represents a single entry in the store.
```go
type Item struct {
    Value     string
    ExpiresAt *time.Time
}
```

### ShardedStore
The top-level store containing multiple shards.
```go
type ShardedStore struct {
    shards []*shard
    count  int
}

type shard struct {
    items map[string]Item
    mu    sync.RWMutex
}
```

## Concurrency Model

- **Sharding**: The key space is partitioned using FNV-1a hashing. This allows multiple goroutines to access different shards concurrently without blocking each other.
- **RWMutex**: Each shard uses a `sync.RWMutex`. 
    - `Get` operations use `RLock`, allowing multiple concurrent readers.
    - `Set`, `Delete`, and `Cleanup` operations use `Lock`, ensuring exclusive write access.

## Expiry Strategy

A hybrid approach is used:
1. **Lazy Deletion**: Every `Get` call checks the `ExpiresAt` field. If the current time is past the expiry, the key is deleted and not returned.
2. **Periodic Cleanup**: A background goroutine iterates through all shards at a configurable interval (`CLEANUP_INTERVAL`) and deletes expired items.

## Scaling Approach

- **Vertical**: Increasing the number of shards (`SHARD_COUNT`) reduces contention on high-core-count machines.
- **Horizontal**: The current implementation is in-memory and non-persistent. To scale horizontally, a consistent hashing proxy or a distributed consensus mechanism (like Raft) would be needed.

## Failure Handling

- **Graceful Shutdown**: The server listens for termination signals and allows in-flight requests to complete within a 10-second timeout.
- **Panic Recovery**: Middleware is used to recover from panics in HTTP handlers, preventing the whole server from crashing.
- **Resource Management**: Background workers are started with a `context.Context` to ensure they stop cleanly when the application shuts down.
