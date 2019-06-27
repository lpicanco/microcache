// Package microcache is the core package
package microcache

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

// NewCache returns a new cache
func NewCache(config Configuration) Cache {
	return newLRUCache(config)
}
