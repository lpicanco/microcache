package microcache

import (
	"container/list"
	"sync"
	"time"
)

type LRUCache struct {
	items             map[interface{}]*cacheItem
	itemRank          *list.List
	mu                sync.RWMutex
	promotions        chan *cacheItem
	maxSize           int
	size              int
	expireAfterWrite  time.Duration
	expireAfterAccess time.Duration
	cleanupCount      int
}

type cacheItem struct {
	key         interface{}
	data        interface{}
	mu          sync.RWMutex
	listElement *list.Element
	createdOn   int64
	accessedOn  int64
	deleted     bool
}

func (c *LRUCache) Put(key interface{}, value interface{}) {
	item := &cacheItem{key: key, data: value, createdOn: getCurrentTimeStamp(), accessedOn: getCurrentTimeStamp()}
	c.mu.Lock()
	c.items[key] = item
	c.mu.Unlock()
	c.promote(item)
}

// Get an item from cache
func (c *LRUCache) Get(key interface{}) (value interface{}, found bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		return
	}

	value = item.data
	c.promote(item)

	return
}

func (c *LRUCache) Invalidate(key interface{}) (found bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if found {
		c.mu.Lock()
		delete(c.items, item.key)
		item.deleted = true

		if item.listElement != nil {
			c.itemRank.Remove(item.listElement)
		}

		c.size--
		c.mu.Unlock()
	}

	return
}

func (c *LRUCache) Close() {
	close(c.promotions)
}

func (c *LRUCache) Len() int {
	return c.size
}

func NewLRUCache(config Configuration) *LRUCache {
	lruCache := &LRUCache{
		itemRank:     list.New(),
		items:        make(map[interface{}]*cacheItem),
		promotions:   make(chan *cacheItem, 1000),
		maxSize:      config.MaxSize,
		cleanupCount: config.CleanupCount,
	}

	go lruCache.doPromotions()
	return lruCache
}

func (c *LRUCache) promote(cacheItem *cacheItem) {
	c.promotions <- cacheItem
}

func (c *LRUCache) doPromotions() {
	for item := range c.promotions {
		if item.deleted {
			continue
		}

		if item.listElement == nil {
			c.size++
			item.listElement = c.itemRank.PushFront(item)

			if c.size > c.maxSize {
				c.cleanup()
			}

			continue
		}

		item.touch()
		c.itemRank.MoveToFront(item.listElement)
	}
}

func (c *LRUCache) cleanup() {
	for i := 0; i < c.cleanupCount; i++ {
		lastItem := c.itemRank.Back()
		if lastItem == nil {
			return
		}
		c.itemRank.Remove(lastItem)
		c.mu.Lock()
		if lastItem.Value != nil {
			delete(c.items, lastItem.Value.(*cacheItem).key)
		}
		c.mu.Unlock()
		c.size--
	}
}

func (ci *cacheItem) touch() {
	// ci.mu.Lock()
	ci.accessedOn = getCurrentTimeStamp()
	// ci.mu.Unlock()
}

func getCurrentTimeStamp() int64 {
	return time.Now().UnixNano()
}
