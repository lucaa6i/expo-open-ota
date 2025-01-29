package cache

import "sync"

type Cache interface {
	Get(key string) string
	Set(key string, value string, ttl *int) error
	Delete(key string)
	Clear() error
}

type CacheType string

const (
	LocalCacheType CacheType = "local"
)

func ResolveCacheType() CacheType {
	return LocalCacheType
}

var (
	cacheInstance Cache
	once          sync.Once
)

func GetCache() Cache {
	once.Do(func() {
		cacheType := ResolveCacheType()
		switch cacheType {
		case LocalCacheType:
			cacheInstance = NewLocalCache()
		default:
			cacheInstance = nil
		}
	})
	return cacheInstance
}
