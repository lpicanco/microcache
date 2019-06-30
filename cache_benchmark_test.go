package microcache

import (
	"sync"
	"testing"

	"github.com/golang/groupcache/lru"
	"github.com/lpicanco/microcache/configuration"
)

func BenchmarkMapPut(b *testing.B) {
	b.StopTimer()
	m := make(map[string]interface{})

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m[string(i)] = i
	}
}

func BenchmarkSyncMapPut(b *testing.B) {
	b.StopTimer()
	var m sync.Map

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Store(string(i), i)
	}
}

func BenchmarkGroupCachePut(b *testing.B) {
	lru := lru.New(100)
	var mu sync.RWMutex

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		lru.Add(i, i)
		mu.Unlock()
	}
}

func BenchmarkPut(b *testing.B) {
	cache := New(configuration.DefaultConfiguration(100))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Put(string(i), i)
	}
}

func BenchmarkMapGet(b *testing.B) {
	b.StopTimer()
	m := make(map[string]interface{})

	for i := 0; i < b.N; i++ {
		m[string(i)] = 42
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if m[string(i)] == nil {
			b.Fatal()
		}
	}
}

func BenchmarkSyncMapGet(b *testing.B) {
	var m sync.Map

	for i := 0; i < b.N; i++ {
		m.Store(string(i), 42)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := m.Load(string(i)); !ok {
			b.Fatal()
		}
	}
}

func BenchmarkGet(b *testing.B) {
	cache := New(configuration.DefaultConfiguration(100))

	for i := 0; i < 100; i++ {
		cache.Put(string(i), 42)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(string(i))
	}
}

func BenchmarkPutGetConcurrent(b *testing.B) {
	cache := New(configuration.DefaultConfiguration(100))

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Put(string(i), 42)
			cache.Get(string(i))
			i++
		}
	})
}

func BenchmarkGroupCacheConcurrent(b *testing.B) {
	cache := lru.New(10000)
	var mu sync.RWMutex

	var wg sync.WaitGroup
	wg.Add(b.N * 2)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		go func(i int) {
			mu.Lock()
			cache.Add(i, i)
			mu.Unlock()
			wg.Done()
		}(i)

		if i%10 == 3 {
			wg.Add(1)
			go func(i int) {
				mu.Lock()
				cache.Remove(i)
				mu.Unlock()
				wg.Done()
			}(i - 1)
		}

		go func(i int) {
			mu.Lock()
			cache.Get(i)
			mu.Unlock()
			wg.Done()
		}(i)
	}

	wg.Wait()
}
