// Package microcache is the core package
package microcache

import (
	"github.com/lpicanco/micro-cache/configuration"
	"github.com/lpicanco/micro-cache/lru"
)

// Cache interface
type Cache interface {
	// Put a item to the cache
	Put(key interface{}, value interface{})
	// Get a item from the cache
	Get(key interface{}) (value interface{}, found bool)
	// Remove a item from the cache.
	Invalidate(key interface{}) (found bool)
	// Len of the cache
	Len() int
	// Close the cache
	Close()
}

// New returns a new cache
func New(config configuration.Configuration) Cache {
	return lru.New(config)
}
