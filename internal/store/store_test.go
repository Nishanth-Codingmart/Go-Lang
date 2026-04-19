package store

import (
	"sync"
	"testing"
	"time"
)

func TestShardedStore_SetGet(t *testing.T) {
	s := NewShardedStore(4)

	s.Set("foo", "bar", 0)
	val, _, exists := s.Get("foo")
	if !exists || val != "bar" {
		t.Errorf("Expected bar, got %s", val)
	}

	s.Delete("foo")
	_, _, exists = s.Get("foo")
	if exists {
		t.Error("Expected foo to be deleted")
	}
}

func TestShardedStore_TTL(t *testing.T) {
	s := NewShardedStore(4)

	s.Set("expired", "val", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	_, _, exists := s.Get("expired")
	if exists {
		t.Error("Expected key to be expired and deleted")
	}
}

func TestShardedStore_Concurrency(t *testing.T) {
	s := NewShardedStore(16)
	wg := sync.WaitGroup{}
	numGoroutines := 100
	opsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := "key-" + string(rune(id)) + "-" + string(rune(j))
				s.Set(key, "val", 0)
				s.Get(key)
			}
		}(i)
	}

	wg.Wait()
	stats := s.Stats()
	if stats.TotalKeys != numGoroutines*opsPerGoroutine {
		t.Errorf("Expected %d keys, got %d", numGoroutines*opsPerGoroutine, stats.TotalKeys)
	}
}

func BenchmarkShardedStore_Set(b *testing.B) {
	s := NewShardedStore(16)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			s.Set("key-"+string(rune(i)), "val", 0)
			i++
		}
	})
}
