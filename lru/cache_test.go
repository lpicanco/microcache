package lru

import (
	"sync"
	"testing"
	"time"

	"github.com/lpicanco/microcache/configuration"
)

func TestPutGet(t *testing.T) {
	cache := New(configuration.DefaultConfiguration(100))

	structValue := struct {
		key   int32
		value string
	}{
		42, "answer",
	}

	cases := []struct {
		in   string
		want interface{}
	}{
		{"Integer value", 432},
		{"String value", "string key"},
		{"String value", "string key 2"},
		{"Array value", [3]string{"Val01", "Val02", "Val03"}},
		{"Struct value", structValue},
		{"Struct reference value", &structValue},
	}
	for _, c := range cases {
		cache.Put(c.in, c.want)

		got, found := cache.Get(c.in)
		if !found {
			t.Errorf("Cache.Get(%q) not found", c.in)
		}

		if got != c.want {
			t.Errorf("Cache.Get(%q) == %v, want %v", c.in, got, c.want)
		}
	}
}

func TestPutTwice(t *testing.T) {
	cache := New(configuration.DefaultConfiguration(100))

	cache.Put(1, 1)
	cache.Put(1, 2)

	if item, _ := cache.Get(1); item != 2 {
		t.Errorf("Cache.Get(1) == %v, want %v", item, 2)
	}

	if cache.Len() != 1 {
		t.Errorf("Cache.Len() == %v, want %v", cache.Len(), 1)
	}
}

func TestNotFound(t *testing.T) {
	cache := New(configuration.DefaultConfiguration(100))
	got, found := cache.Get("key")

	if found {
		t.Error("Cache.Get(key) returned true")
	}

	if got != nil {
		t.Errorf("Cache.Get(key) == %v, want %v", got, nil)
	}

}

func TestExpiredAfterWrite(t *testing.T) {
	expireAfterWrite := time.Millisecond * 10
	cache := New(configuration.Configuration{MaxSize: 100, ExpireAfterWrite: expireAfterWrite})
	cache.Put(1, 1)

	<-time.After(expireAfterWrite - time.Millisecond)
	if _, found := cache.Get(1); !found {
		t.Error("Cache.Get(key) returned false")
	}

	<-time.After(time.Millisecond)

	got, found := cache.Get(1)

	if found {
		t.Error("Cache.Get(key) returned true")
	}

	if got != nil {
		t.Errorf("Cache.Get(key) == %v, want %v", got, nil)
	}
}

func TestExpiredAfterAccess(t *testing.T) {
	expireAfterAccess := time.Millisecond * 10
	cache := New(configuration.Configuration{MaxSize: 100, ExpireAfterAccess: expireAfterAccess, ExpireAfterWrite: 1 * time.Second})
	key := 1
	cache.Put(key, 1)

	<-time.After(expireAfterAccess - time.Millisecond)
	if _, found := cache.Get(key); !found {
		t.Error("Cache.Get(key) returned false")
	}

	<-time.After(expireAfterAccess - time.Millisecond)
	if _, found := cache.Get(key); !found {
		t.Error("Cache.Get(key) returned false")
	}

	<-time.After(expireAfterAccess)

	got, found := cache.Get(1)

	if found {
		t.Error("Cache.Get(key) returned true")
	}

	if got != nil {
		t.Errorf("Cache.Get(key) == %v, want %v", got, nil)
	}
}
func TestLRUSizeEviction(t *testing.T) {
	cases := []struct {
		in   int
		want bool
	}{
		{3, false},
		{2, false},
		{0, true},
		{4, true},
		{1, true},
	}

	maxSize := 5
	cache := New(configuration.Configuration{MaxSize: maxSize, CleanupCount: 1})

	for _, i := range cases {
		cache.Put(i.in, i)
	}

	for _, c := range cases {
		cache.Get(c.in)
	}

	cache.Put(5, 5)
	cache.Put(6, 6)

	<-time.After(time.Millisecond * 10)

	for _, c := range cases {
		if _, found := cache.Get(c.in); found != c.want {
			t.Errorf("Cache.Get(%v) == %v", c.in, found)
		}
	}
}

func TestLRUCleanup(t *testing.T) {
	maxSize := 100
	cleanUpCount := 25
	want := maxSize - cleanUpCount + 1
	cache := New(configuration.Configuration{MaxSize: maxSize, CleanupCount: cleanUpCount})

	for i := 0; i <= maxSize; i++ {
		cache.Put(i, i)
	}

	<-time.After(time.Millisecond * 10)

	if cache.Len() != want {
		t.Errorf("Cache.Len() == %v. want %v", cache.Len(), want)
	}
}

func TestLRUInvalidate(t *testing.T) {
	maxSize := 100
	cache := New(configuration.Configuration{MaxSize: maxSize})

	cache.Put(1, 1)
	cache.Put(2, 2)
	cache.Put(3, 3)

	if found := cache.Invalidate(2); !found {
		t.Errorf("Cache.Invalidate(2) == false")
	}

	if _, found := cache.Get(2); found {
		t.Errorf("Cache.Get(2) == found")
	}

	if found := cache.Invalidate(20); found {
		t.Errorf("Cache.Invalidate(20) == true")
	}
}

func BenchmarkLRUConcurrent(b *testing.B) {
	cache := New(configuration.DefaultConfiguration(10000))
	defer cache.Close()

	var wg sync.WaitGroup
	wg.Add(b.N * 2)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		go func(i int) {
			cache.Put(i, i)
			wg.Done()
		}(i)

		if i%10 == 3 {
			wg.Add(1)
			go func(i int) {
				cache.Invalidate(i)
				wg.Done()
			}(i - 1)
		}

		go func(i int) {
			cache.Get(i)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func BenchmarkLRUGetSequencial(b *testing.B) {
	maxSize := 10000
	cache := New(configuration.DefaultConfiguration(maxSize))
	defer cache.Close()

	var wg sync.WaitGroup
	wg.Add(maxSize)

	for i := 0; i < maxSize; i++ {
		go func(i int) {
			cache.Put(i, i)
			wg.Done()
		}(i)
	}
	wg.Wait()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Get(i % maxSize)
	}
}
