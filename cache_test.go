package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMemoryCache_GetSet(t *testing.T) {
	cache := NewMemoryCache()

	// Test set and get
	cache.Set("key1", "value1", time.Hour)
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	// Test non-existent key
	_, ok = cache.Get("nonexistent")
	if ok {
		t.Error("expected nonexistent key to not exist")
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache()

	// Set with very short TTL
	cache.Set("expiring", "value", 50*time.Millisecond)

	// Should exist immediately
	val, ok := cache.Get("expiring")
	if !ok {
		t.Error("expected key to exist before expiration")
	}
	if val != "value" {
		t.Errorf("expected value, got %v", val)
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should be gone now
	_, ok = cache.Get("expiring")
	if ok {
		t.Error("expected key to be expired")
	}
}

func TestMemoryCache_NoExpiry(t *testing.T) {
	cache := NewMemoryCache()

	// Set with zero TTL (no expiry)
	cache.Set("permanent", "value", 0)

	val, ok := cache.Get("permanent")
	if !ok {
		t.Error("expected permanent key to exist")
	}
	if val != "value" {
		t.Errorf("expected value, got %v", val)
	}

	// Set with negative TTL (also no expiry)
	cache.Set("permanent2", "value2", -1*time.Second)

	val, ok = cache.Get("permanent2")
	if !ok {
		t.Error("expected permanent2 key to exist")
	}
	if val != "value2" {
		t.Errorf("expected value2, got %v", val)
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("key", "value", time.Hour)
	cache.Delete("key")

	_, ok := cache.Get("key")
	if ok {
		t.Error("expected deleted key to not exist")
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("key1", "value1", time.Hour)
	cache.Set("key2", "value2", time.Hour)
	cache.Set("key3", "value3", time.Hour)

	if cache.Size() != 3 {
		t.Errorf("expected size 3, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", cache.Size())
	}

	_, ok := cache.Get("key1")
	if ok {
		t.Error("expected cleared cache to be empty")
	}
}

func TestMemoryCache_Cleanup(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("fresh", "value", time.Hour)
	cache.Set("expiring1", "value1", 10*time.Millisecond)
	cache.Set("expiring2", "value2", 10*time.Millisecond)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	removed := cache.Cleanup()
	if removed != 2 {
		t.Errorf("expected 2 entries removed, got %d", removed)
	}

	if cache.Size() != 1 {
		t.Errorf("expected size 1 after cleanup, got %d", cache.Size())
	}

	// Fresh entry should still exist
	_, ok := cache.Get("fresh")
	if !ok {
		t.Error("expected fresh entry to still exist")
	}
}

func TestMemoryCache_Overwrite(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("key", "value1", time.Hour)
	cache.Set("key", "value2", time.Hour)

	val, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if val != "value2" {
		t.Errorf("expected value2, got %v", val)
	}
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache()
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			cache.Set("key", i, time.Hour)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			cache.Get("key")
		}
		done <- true
	}()

	// Deleter goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Delete("key")
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}

func TestDefaultCacheConfig(t *testing.T) {
	config := DefaultCacheConfig()

	if config.Cache == nil {
		t.Error("expected default cache to be set")
	}
	if config.CapabilityTTL != time.Hour {
		t.Errorf("expected 1 hour capability TTL, got %v", config.CapabilityTTL)
	}
	if config.DeviceProfileTTL != time.Hour {
		t.Errorf("expected 1 hour device profile TTL, got %v", config.DeviceProfileTTL)
	}
}

func TestCacheKey(t *testing.T) {
	tests := []struct {
		resourceType string
		ids          []string
		expected     string
	}{
		{"capability", []string{"switch", "1"}, "capability:switch:1"},
		{"deviceprofile", []string{"abc123"}, "deviceprofile:abc123"},
		{"type", nil, "type"},
		{"type", []string{}, "type"},
	}

	for _, tt := range tests {
		result := cacheKey(tt.resourceType, tt.ids...)
		if result != tt.expected {
			t.Errorf("cacheKey(%s, %v) = %s, want %s", tt.resourceType, tt.ids, result, tt.expected)
		}
	}
}

func TestWithCache(t *testing.T) {
	// Test with default config
	client, err := NewClient("test-token", WithCache(nil))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client.cacheConfig == nil {
		t.Error("expected cache config to be set with defaults")
	}

	// Test with custom config
	customCache := NewMemoryCache()
	customConfig := &CacheConfig{
		Cache:            customCache,
		CapabilityTTL:    30 * time.Minute,
		DeviceProfileTTL: 2 * time.Hour,
	}
	client2, err := NewClient("test-token", WithCache(customConfig))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client2.cacheConfig.CapabilityTTL != 30*time.Minute {
		t.Errorf("expected 30 minute capability TTL, got %v", client2.cacheConfig.CapabilityTTL)
	}

	// Test with config that has nil Cache but custom TTLs (should create default cache)
	configNilCache := &CacheConfig{
		Cache:            nil,
		CapabilityTTL:    45 * time.Minute,
		DeviceProfileTTL: 0, // Should get default 1 hour
	}
	client3, err := NewClient("test-token", WithCache(configNilCache))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client3.cacheConfig.Cache == nil {
		t.Error("expected cache to be created")
	}
	if client3.cacheConfig.CapabilityTTL != 45*time.Minute {
		t.Errorf("expected 45 minute capability TTL, got %v", client3.cacheConfig.CapabilityTTL)
	}
	if client3.cacheConfig.DeviceProfileTTL != time.Hour {
		t.Errorf("expected 1 hour device profile TTL (default), got %v", client3.cacheConfig.DeviceProfileTTL)
	}
}

func TestGetCapabilityWithCache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": "switch", "version": 1, "status": "live"}`))
	}))
	defer server.Close()

	client, _ := NewClient("test-token",
		WithBaseURL(server.URL),
		WithCache(DefaultCacheConfig()),
	)

	ctx := context.Background()

	// First call should hit the server
	cap1, err := client.GetCapability(ctx, "switch", 1)
	if err != nil {
		t.Fatalf("GetCapability failed: %v", err)
	}
	if cap1.ID != "switch" {
		t.Errorf("expected switch, got %s", cap1.ID)
	}
	if callCount != 1 {
		t.Errorf("expected 1 server call, got %d", callCount)
	}

	// Second call should use cache
	cap2, err := client.GetCapability(ctx, "switch", 1)
	if err != nil {
		t.Fatalf("GetCapability failed: %v", err)
	}
	if cap2.ID != "switch" {
		t.Errorf("expected switch, got %s", cap2.ID)
	}
	if callCount != 1 {
		t.Errorf("expected still 1 server call (cached), got %d", callCount)
	}
}

func TestGetDeviceProfileWithCache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": "profile1", "name": "Test Profile", "components": []}`))
	}))
	defer server.Close()

	client, _ := NewClient("test-token",
		WithBaseURL(server.URL),
		WithCache(DefaultCacheConfig()),
	)

	ctx := context.Background()

	// First call should hit the server
	profile1, err := client.GetDeviceProfile(ctx, "profile1")
	if err != nil {
		t.Fatalf("GetDeviceProfile failed: %v", err)
	}
	if profile1.ID != "profile1" {
		t.Errorf("expected profile1, got %s", profile1.ID)
	}
	if callCount != 1 {
		t.Errorf("expected 1 server call, got %d", callCount)
	}

	// Second call should use cache
	profile2, err := client.GetDeviceProfile(ctx, "profile1")
	if err != nil {
		t.Fatalf("GetDeviceProfile failed: %v", err)
	}
	if profile2.ID != "profile1" {
		t.Errorf("expected profile1, got %s", profile2.ID)
	}
	if callCount != 1 {
		t.Errorf("expected still 1 server call (cached), got %d", callCount)
	}
}

func TestInvalidateCache(t *testing.T) {
	client, _ := NewClient("test-token", WithCache(DefaultCacheConfig()))

	// Add something to cache
	client.cacheConfig.Cache.Set(cacheKey("capability", "switch", "1"), "cached", time.Hour)

	// Verify it's there
	_, ok := client.cacheConfig.Cache.Get(cacheKey("capability", "switch", "1"))
	if !ok {
		t.Error("expected cached value to exist")
	}

	// Invalidate
	client.InvalidateCache("capability", "switch", "1")

	// Verify it's gone
	_, ok = client.cacheConfig.Cache.Get(cacheKey("capability", "switch", "1"))
	if ok {
		t.Error("expected cached value to be invalidated")
	}
}

func TestGetCached_FetchError(t *testing.T) {
	// Test that getCached propagates fetch errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	client, _ := NewClient("test-token",
		WithBaseURL(server.URL),
		WithCache(DefaultCacheConfig()),
	)

	ctx := context.Background()

	// This should fail and the error should propagate through getCached
	_, err := client.GetCapability(ctx, "switch", 1)
	if err == nil {
		t.Error("expected error from server failure")
	}
}

func TestGetCapabilityWithoutCache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": "switch", "version": 1, "status": "live"}`))
	}))
	defer server.Close()

	// Client without cache
	client, _ := NewClient("test-token", WithBaseURL(server.URL))

	ctx := context.Background()

	// Both calls should hit the server
	client.GetCapability(ctx, "switch", 1)
	client.GetCapability(ctx, "switch", 1)

	if callCount != 2 {
		t.Errorf("expected 2 server calls without cache, got %d", callCount)
	}
}

func TestClient_InvalidateCapabilityCache(t *testing.T) {
	config := DefaultCacheConfig()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": "switch", "version": 1, "status": "live"}`))
	}))
	defer server.Close()

	client, _ := NewClient("test-token", WithBaseURL(server.URL), WithCache(config))

	ctx := context.Background()

	// First call - should hit server
	client.GetCapability(ctx, "switch", 1)
	if callCount != 1 {
		t.Errorf("expected 1 server call, got %d", callCount)
	}

	// Second call - should use cache
	client.GetCapability(ctx, "switch", 1)
	if callCount != 1 {
		t.Errorf("expected still 1 server call after cache hit, got %d", callCount)
	}

	// Invalidate capability cache
	client.InvalidateCapabilityCache()

	// Third call - should hit server again
	client.GetCapability(ctx, "switch", 1)
	if callCount != 2 {
		t.Errorf("expected 2 server calls after invalidation, got %d", callCount)
	}
}
