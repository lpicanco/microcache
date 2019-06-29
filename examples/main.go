package main

import (
	"fmt"

	microcache "github.com/lpicanco/micro-cache"
	"github.com/lpicanco/micro-cache/configuration"
)

func main() {
	cache := microcache.New(configuration.DefaultConfiguration(100))
	cache.Put(42, "answer")

	value, found := cache.Get(42)
	if found {
		fmt.Printf("Value: %v\n", value)
	}
}
