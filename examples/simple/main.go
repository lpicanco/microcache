package main

import (
	"fmt"

	"github.com/lpicanco/microcache"
	"github.com/lpicanco/microcache/configuration"
)

func main() {
	cache := microcache.New(configuration.DefaultConfiguration(100))
	cache.Put(42, "answer")

	value, found := cache.Get(42)
	if found {
		fmt.Printf("Value: %v\n", value)
	}

	fmt.Printf("Cache len: %v\n", cache.Len())

	cache.Invalidate(42)
}
