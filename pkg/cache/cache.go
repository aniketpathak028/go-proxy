package cache

import (
	"net/http"
	"sync"
	"time"
)

// http response cache
type CacheEntry struct {
	Response     *http.Response
	Body         []byte
	CachedAt     time.Time
	ExpiresAt    time.Time
	ETag         string
	LastModified string
}

// interface for http cache
type Cache interface {
	Get(key string) (*CacheEntry, bool)
	Set(key string, entry *CacheEntry)
	Delete(key string)
}

// in-mem cache implements Cache
type MemoryCache struct {
	entries map[string]*CacheEntry
	mutex   sync.RWMutex
}

// in-mem constructor
func NewMemCache() *MemoryCache {
	return &MemoryCache{
		entries: make(map[string]*CacheEntry),
	}
}

/*
   in-mem cache methods
*/

// retrieve an entry from cache
func (c *MemoryCache) Get(key string) (*CacheEntry, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// not found
	entry, found := c.entries[key]
	if !found {
		return nil, false
	}

	// check expiry
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		return entry, true
	}

	return entry, true
}

// adds an entry to cache
func (c *MemoryCache) Set(key string, entry *CacheEntry) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[key] = entry
}

// deletes an entry from cache
func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.entries, key)
}
