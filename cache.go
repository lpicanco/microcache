package microcache

type Cache interface {
	Put(key interface{}, value interface{})
	Get(key interface{}) (value interface{}, found bool)
	Invalidate(key interface{}) (found bool)
	Len() int
	Close()
}

// Return a new cache
func NewCache() Cache {
	return NewLRUCache(DefaultConfiguration(100))
}
