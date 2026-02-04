package repo

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	value      interface{}
	expiration time.Time
}

// Expired checks if the cache item has expired
func (item *CacheItem) Expired() bool {
	return !item.expiration.IsZero() && time.Now().After(item.expiration)
}

// Cache represents an in-memory cache
type Cache struct {
	items      map[string]*CacheItem
	mu         sync.RWMutex
	defaultTTL time.Duration
}

// NewCache creates a new cache instance
func NewCache(defaultTTL time.Duration) *Cache {
	return &Cache{
		items:      make(map[string]*CacheItem),
		defaultTTL: defaultTTL,
	}
}

// Set adds an item to the cache
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL adds an item to the cache with custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	c.items[key] = &CacheItem{
		value:      value,
		expiration: expiration,
	}
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		return nil, false
	}

	// Check if expired
	if item.Expired() {
		c.Delete(key)
		return nil, false
	}

	return item.value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheItem)
}

// Exists checks if a key exists in the cache
func (c *Cache) Exists(key string) bool {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if found && !item.Expired() {
		return true
	}

	if found && item.Expired() {
		c.Delete(key)
	}

	return false
}

// Keys returns all cache keys
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key, item := range c.items {
		if !item.Expired() {
			keys = append(keys, key)
		}
	}
	return keys
}

// Size returns the number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, item := range c.items {
		if !item.Expired() {
			count++
		}
	}
	return count
}

// Cleanup removes expired items from the cache
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if item.Expired() {
			delete(c.items, key)
		}
	}
}

// GetOrSet retrieves a value from cache or sets it if not found
func (c *Cache) GetOrSet(key string, fn func() (interface{}, error), ttl ...time.Duration) (interface{}, error) {
	// Try to get from cache first
	if val, found := c.Get(key); found {
		return val, nil
	}

	// Execute function to get value
	val, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if len(ttl) > 0 && ttl[0] > 0 {
		c.SetWithTTL(key, val, ttl[0])
	} else {
		c.Set(key, val)
	}

	return val, nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	Items    int `json:"items"`
	Capacity int `json:"capacity"`
	Hits     int `json:"hits"`
	Misses   int `json:"misses"`
}

// WithBackgroundCleanup creates a cache with background cleanup
func NewCacheWithCleanup(defaultTTL time.Duration, cleanupInterval time.Duration) *Cache {
	cache := NewCache(defaultTTL)

	// Start background cleanup goroutine
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			cache.Cleanup()
		}
	}()

	return cache
}
