package microcache

import (
	"container/list"
	"sync"
	"time"
)

type lruCache struct {
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

func (c *lruCache) Put(key interface{}, value interface{}) {
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

func (c *lruCache) Get(key interface{}) (value interface{}, found bool) {
	item, found := c.getCacheItem(key)

	if !found {
		return
	}

	item.mu.RLock()
	value = item.data
	item.mu.RUnlock()
	c.promote(item)

	return
}

func (c *lruCache) Invalidate(key interface{}) (found bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if found {
		c.mu.Lock()
		delete(c.items, item.key)
		c.mu.Unlock()

		c.deletions <- item
	}

	return
}

func (c *lruCache) Close() {
	close(c.promotions)
	close(c.deletions)
}

func (c *lruCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func newLRUCache(config Configuration) *lruCache {
	lruCache := &lruCache{
		itemRank:     list.New(),
		items:        make(map[interface{}]*cacheItem),
		promotions:   make(chan *cacheItem, 1000),
		deletions:    make(chan *cacheItem, 1000),
		maxSize:      config.MaxSize,
		cleanupCount: config.CleanupCount,
	}

	go lruCache.doPromotions()
	return lruCache
}

func (c *lruCache) getCacheItem(key interface{}) (cacheItem *cacheItem, found bool) {
	c.mu.RLock()
	cacheItem, found = c.items[key]
	c.mu.RUnlock()
	return
}

func (c *lruCache) promote(cacheItem *cacheItem) {
	c.promotions <- cacheItem
}

func (c *lruCache) doPromotions() {
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

func (c *lruCache) cleanup() {
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

func getCurrentTimeStamp() int64 {
	return time.Now().UnixNano()
}
