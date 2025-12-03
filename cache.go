package smartthings

import (
	"sync"
	"time"
)

// Cache defines an interface for caching API responses.
// Implementations must be safe for concurrent access.
type Cache interface {
	// Get retrieves a value from the cache.
	// Returns the value and true if found and not expired, or nil and false otherwise.
	Get(key string) (any, bool)

	// Set stores a value in the cache with the given TTL.
	// If TTL is 0 or negative, the entry never expires.
	Set(key string, value any, ttl time.Duration)

	// Delete removes a value from the cache.
	Delete(key string)

	// Clear removes all values from the cache.
	Clear()
}

// cacheEntry holds a cached value with its expiration time.
type cacheEntry struct {
	value     any
	expiresAt time.Time
	noExpiry  bool
}

// MemoryCache is a thread-safe in-memory cache implementation.
type MemoryCache struct {
	entries map[string]*cacheEntry
	mu      sync.RWMutex
}

// NewMemoryCache creates a new in-memory cache.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		entries: make(map[string]*cacheEntry),
	}
}

// Get retrieves a value from the cache.
func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Check expiration
	if !entry.noExpiry && time.Now().After(entry.expiresAt) {
		// Entry expired, remove it
		c.Delete(key)
		return nil, false
	}

	return entry.value, true
}

// Set stores a value in the cache with the given TTL.
func (c *MemoryCache) Set(key string, value any, ttl time.Duration) {
	entry := &cacheEntry{
		value: value,
	}

	if ttl <= 0 {
		entry.noExpiry = true
	} else {
		entry.expiresAt = time.Now().Add(ttl)
	}

	c.mu.Lock()
	c.entries[key] = entry
	c.mu.Unlock()
}

// Delete removes a value from the cache.
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Clear removes all values from the cache.
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]*cacheEntry)
	c.mu.Unlock()
}

// Size returns the number of entries in the cache (including expired ones).
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Cleanup removes all expired entries from the cache.
// Call this periodically to prevent memory leaks from expired entries.
func (c *MemoryCache) Cleanup() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0

	for key, entry := range c.entries {
		if !entry.noExpiry && now.After(entry.expiresAt) {
			delete(c.entries, key)
			removed++
		}
	}

	return removed
}

// CacheConfig configures the caching behavior for a Client.
type CacheConfig struct {
	// Cache is the cache implementation to use.
	Cache Cache

	// CapabilityTTL is how long to cache capability definitions.
	// Defaults to 1 hour if zero.
	CapabilityTTL time.Duration

	// DeviceProfileTTL is how long to cache device profiles.
	// Defaults to 1 hour if zero.
	DeviceProfileTTL time.Duration
}

// DefaultCacheConfig returns a CacheConfig with sensible defaults.
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Cache:            NewMemoryCache(),
		CapabilityTTL:    1 * time.Hour,
		DeviceProfileTTL: 1 * time.Hour,
	}
}

// cacheKey generates a cache key for the given resource type and identifiers.
func cacheKey(resourceType string, ids ...string) string {
	key := resourceType
	for _, id := range ids {
		key += ":" + id
	}
	return key
}

// WithCache enables response caching for the client.
// Cached resources include capability definitions and device profiles.
//
// Example:
//
//	client, _ := smartthings.NewClient(token,
//	    smartthings.WithCache(smartthings.DefaultCacheConfig()),
//	)
func WithCache(config *CacheConfig) Option {
	return func(c *Client) {
		if config == nil {
			config = DefaultCacheConfig()
		}
		if config.Cache == nil {
			config.Cache = NewMemoryCache()
		}
		if config.CapabilityTTL == 0 {
			config.CapabilityTTL = 1 * time.Hour
		}
		if config.DeviceProfileTTL == 0 {
			config.DeviceProfileTTL = 1 * time.Hour
		}
		c.cacheConfig = config
	}
}

// getCached retrieves a value from cache or executes the fetch function and caches the result.
func (c *Client) getCached(key string, ttl time.Duration, fetch func() (any, error)) (any, error) {
	if c.cacheConfig == nil || c.cacheConfig.Cache == nil {
		return fetch()
	}

	// Check cache first
	if cached, ok := c.cacheConfig.Cache.Get(key); ok {
		return cached, nil
	}

	// Fetch and cache
	result, err := fetch()
	if err != nil {
		return nil, err
	}

	c.cacheConfig.Cache.Set(key, result, ttl)
	return result, nil
}

// InvalidateCapabilityCache removes all cached capability entries.
func (c *Client) InvalidateCapabilityCache() {
	if c.cacheConfig != nil && c.cacheConfig.Cache != nil {
		// Clear all entries (we don't have a prefix-based delete)
		c.cacheConfig.Cache.Clear()
	}
}

// InvalidateCache removes a specific entry from the cache.
func (c *Client) InvalidateCache(resourceType string, ids ...string) {
	if c.cacheConfig != nil && c.cacheConfig.Cache != nil {
		c.cacheConfig.Cache.Delete(cacheKey(resourceType, ids...))
	}
}
