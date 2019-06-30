package main

import (
	"fmt"
	"time"

	"github.com/lpicanco/microcache"
	"github.com/lpicanco/microcache/configuration"
)

func main() {
	cache := microcache.New(configuration.Configuration{
		MaxSize:           10000,
		ExpireAfterWrite:  1 * time.Hour,
		ExpireAfterAccess: 10 * time.Minute,
		CleanupCount:      5,
	})
	cache.Put(42, "answer")

	value, found := cache.Get(42)
	if found {
		fmt.Printf("Value: %v\n", value)
	}

	fmt.Printf("Cache len: %v\n", cache.Len())

	cache.Invalidate(42)
}
