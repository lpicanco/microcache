package lru

import (
	"container/list"
	"sync"
	"time"

	"github.com/lpicanco/microcache/configuration"
)

// Cache structure
type Cache struct {
	items             map[interface{}]*cacheItem
	itemRank          *list.List
	mu                sync.RWMutex
	promotions        chan *cacheItem
	deletions         chan *cacheItem
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
}

// Put a item to the cache
func (c *Cache) Put(key interface{}, value interface{}) {
	item, found := c.getCacheItem(key)

	if found {
		item.mu.Lock()
		item.data = value
		item.mu.Unlock()
		c.promote(item)
		return
	}

	item = &cacheItem{key: key, data: value, createdOn: getCurrentTimeStamp(), accessedOn: getCurrentTimeStamp()}
	c.mu.Lock()
	c.items[key] = item
	c.mu.Unlock()
	c.promote(item)
}

// Get a item from the cache
func (c *Cache) Get(key interface{}) (value interface{}, found bool) {
	item, found := c.getCacheItem(key)

	if !found {
		return
	}

	if item.expired(c.expireAfterWrite, c.expireAfterAccess) {
		c.invalidate(item)
		return nil, false
	}

	item.mu.RLock()
	value = item.data
	item.mu.RUnlock()
	c.promote(item)

	return
}

// Invalidate remove a item from the cache.
func (c *Cache) Invalidate(key interface{}) (found bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if found {
		c.invalidate(item)
	}

	return
}

// Close the cache
func (c *Cache) Close() {
	close(c.promotions)
	close(c.deletions)
}

// Len of the cache
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// New returns a new LRU Cache
func New(config configuration.Configuration) *Cache {
	Cache := &Cache{
		itemRank:          list.New(),
		items:             make(map[interface{}]*cacheItem),
		promotions:        make(chan *cacheItem, 1000),
		deletions:         make(chan *cacheItem, 1000),
		maxSize:           config.MaxSize,
		cleanupCount:      config.CleanupCount,
		expireAfterWrite:  config.ExpireAfterWrite,
		expireAfterAccess: config.ExpireAfterAccess,
	}

	go Cache.doPromotions()
	return Cache
}

func (c *Cache) getCacheItem(key interface{}) (cacheItem *cacheItem, found bool) {
	c.mu.RLock()
	cacheItem, found = c.items[key]
	c.mu.RUnlock()
	return
}

func (c *Cache) invalidate(item *cacheItem) {
	c.mu.Lock()
	delete(c.items, item.key)
	c.mu.Unlock()

	c.deletions <- item
}

func (c *Cache) promote(cacheItem *cacheItem) {
	c.promotions <- cacheItem
}

func (c *Cache) doPromotions() {
	for {
		select {
		case item, ok := <-c.promotions:
			if !ok {
				return
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

		case item, ok := <-c.deletions:
			if !ok {
				return
			}

			if item.listElement == nil {
				continue
			}

			c.itemRank.Remove(item.listElement)
			c.size--
		}
	}
}

func (c *Cache) cleanup() {
	for i := 0; i < c.cleanupCount; i++ {
		lastItem := c.itemRank.Back()
		if lastItem == nil {
			return
		}
		c.itemRank.Remove(lastItem)
		c.mu.Lock()
		delete(c.items, lastItem.Value.(*cacheItem).key)
		c.size--
		c.mu.Unlock()
	}
}

func (ci *cacheItem) touch() {
	ci.mu.Lock()
	ci.accessedOn = getCurrentTimeStamp()
	ci.mu.Unlock()
}

func (ci *cacheItem) expired(expireAfterWrite time.Duration, expireAfterAccess time.Duration) bool {
	if expireAfterWrite > 0 && getCurrentTimeStamp()-ci.createdOn >= expireAfterWrite.Nanoseconds() {
		return true
	}

	if expireAfterAccess > 0 {
		ci.mu.RLock()
		accessedOn := ci.accessedOn
		ci.mu.RUnlock()

		if getCurrentTimeStamp()-accessedOn >= expireAfterAccess.Nanoseconds() {
			return true
		}
	}

	return false
}

func getCurrentTimeStamp() int64 {
	return time.Now().UnixNano()
}
