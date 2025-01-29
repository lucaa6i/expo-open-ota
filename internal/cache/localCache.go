package cache

import (
	"sync"
	"time"
)

type LocalCache struct {
	items map[string]CacheItem
	mu    sync.RWMutex // RWMutex for safe concurrent access
}

type CacheItem struct {
	Value      string
	Expiration *time.Time // nil if no TTL
}

func NewLocalCache() *LocalCache {
	return &LocalCache{
		items: make(map[string]CacheItem),
	}
}

func (c *LocalCache) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return ""
	}

	if item.Expiration != nil && time.Now().After(*item.Expiration) {
		delete(c.items, key)
		return ""
	}

	return item.Value
}

func (c *LocalCache) Set(key string, value string, ttl *int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration *time.Time
	if ttl != nil {
		exp := time.Now().Add(time.Duration(*ttl) * time.Second)
		expiration = &exp
	}

	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
	return nil
}

func (c *LocalCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; !exists {
		return
	}

	delete(c.items, key)
	return
}

func (c *LocalCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]CacheItem)
	return nil
}
