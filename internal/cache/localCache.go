package cache

import (
	"expo-open-ota/internal/version"
	"fmt"
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

	item, exists := c.items[withPrefix(key)]
	if !exists {
		return ""
	}

	if item.Expiration != nil && time.Now().After(*item.Expiration) {
		delete(c.items, withPrefix(key))
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

	c.items[withPrefix(key)] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
	return nil
}

func (c *LocalCache) Delete(key string) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, withPrefix(key))
}

func (c *LocalCache) Clear() error {
	if version.Version != "development" {
		fmt.Println("Cache can only be cleared in development mode.")
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]CacheItem)
	return nil
}

func (c *LocalCache) TryLock(key string, ttl int) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[withPrefix(key)]; exists {
		return false, nil
	}

	exp := time.Now().Add(time.Duration(ttl) * time.Second)
	c.items[withPrefix(key)] = CacheItem{
		Value:      "locked",
		Expiration: &exp,
	}

	go func() {
		time.Sleep(time.Duration(ttl) * time.Second)
		c.mu.Lock()
		delete(c.items, withPrefix(key))
		c.mu.Unlock()
	}()

	return true, nil
}
