package microcache

import (
	"sync"
	"testing"
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

func BenchmarkPut(b *testing.B) {
	b.StopTimer()
	cache := NewCache()

	b.StartTimer()
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
	b.StopTimer()
	cache := NewCache()

	for i := 0; i < b.N; i++ {
		cache.Put(string(i), 42)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := cache.Get(string(i)); !ok {
			b.Fatal()
		}
	}
}

func BenchmarkPutGetConcurrent(b *testing.B) {
	cache := NewCache()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Put(string(i), 42)
			cache.Get(string(i))
			i++
		}
	})
}