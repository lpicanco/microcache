package microcache

import "sync"

type cacheItem struct {
	Data interface{}
}

// Cache struct
type Cache struct {
	items map[string]cacheItem
	m     sync.RWMutex
}

// Return a new cache
func NewCache() Cache {
	return Cache{items: make(map[string]cacheItem)}
}

// Put an item to cache
func (c Cache) Put(key string, value interface{}) {
	c.m.Lock()

	c.items[key] = cacheItem{Data: value}

	c.m.Unlock()
}

// Get an item from cache
func (c Cache) Get(key string) (value interface{}, found bool) {
	c.m.RLock()
	item, found := c.items[key]

	if found {
		value = item.Data
	}

	c.m.RUnlock()
	return
}
