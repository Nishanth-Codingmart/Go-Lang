package store

import (
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/user/kvstore/internal/metrics"
)

type Store interface {
	Set(key, value string, ttl time.Duration)
	Get(key string) (string, *time.Time, bool)
	Delete(key string)
	Stats() Stats
	Cleanup() int
}

type Stats struct {
	TotalKeys      int   `json:"total_keys"`
	ExpiredKeys    int64 `json:"expired_keys"`
	GetRequests    int64 `json:"get_requests"`
	PutRequests    int64 `json:"put_requests"`
	DeleteRequests int64 `json:"delete_requests"`
}

type shard struct {
	items map[string]Item
	mu    sync.RWMutex
}

type ShardedStore struct {
	shards         []*shard
	count          int
	expiredCount   int64
	getRequests    int64
	putRequests    int64
	deleteRequests int64
}

func NewShardedStore(count int) *ShardedStore {
	if count <= 0 {
		count = 16
	}
	shards := make([]*shard, count)
	for i := 0; i < count; i++ {
		shards[i] = &shard{
			items: make(map[string]Item),
		}
	}
	return &ShardedStore{
		shards: shards,
		count:  count,
	}
}

func (s *ShardedStore) getShard(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return s.shards[int(h.Sum32())%s.count]
}

func (s *ShardedStore) Set(key, value string, ttl time.Duration) {
	atomic.AddInt64(&s.putRequests, 1)
	metrics.PutRequests.Inc()
	var expiresAt *time.Time
	if ttl > 0 {
		t := time.Now().Add(ttl)
		expiresAt = &t
	}

	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, exists := shard.items[key]; !exists {
		metrics.ActiveKeys.Inc()
	}
	shard.items[key] = Item{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

func (s *ShardedStore) Get(key string) (string, *time.Time, bool) {
	atomic.AddInt64(&s.getRequests, 1)
	metrics.GetRequests.Inc()
	shard := s.getShard(key)
	shard.mu.RLock()
	item, exists := shard.items[key]
	shard.mu.RUnlock()

	if !exists {
		return "", nil, false
	}

	if item.IsExpired() {
		// Lazy deletion
		s.Delete(key)
		atomic.AddInt64(&s.expiredCount, 1)
		metrics.ExpiredKeys.Inc()
		return "", nil, false
	}

	return item.Value, item.ExpiresAt, true
}

func (s *ShardedStore) Delete(key string) {
	atomic.AddInt64(&s.deleteRequests, 1)
	metrics.DeleteRequests.Inc()
	shard := s.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, exists := shard.items[key]; exists {
		delete(shard.items, key)
		metrics.ActiveKeys.Dec()
	}
}

func (s *ShardedStore) Stats() Stats {
	totalKeys := 0
	for _, shard := range s.shards {
		shard.mu.RLock()
		totalKeys += len(shard.items)
		shard.mu.RUnlock()
	}

	return Stats{
		TotalKeys:      totalKeys,
		ExpiredKeys:    atomic.LoadInt64(&s.expiredCount),
		GetRequests:    atomic.LoadInt64(&s.getRequests),
		PutRequests:    atomic.LoadInt64(&s.putRequests),
		DeleteRequests: atomic.LoadInt64(&s.deleteRequests),
	}
}

func (s *ShardedStore) Cleanup() int {
	expiredCount := 0
	for _, shard := range s.shards {
		shard.mu.Lock()
		for k, v := range shard.items {
			if v.IsExpired() {
				delete(shard.items, k)
				expiredCount++
				metrics.ActiveKeys.Dec()
				metrics.ExpiredKeys.Inc()
			}
		}
		shard.mu.Unlock()
	}
	atomic.AddInt64(&s.expiredCount, int64(expiredCount))
	return expiredCount
}
