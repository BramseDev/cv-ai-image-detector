package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type AnalysisCache struct {
	mu    sync.RWMutex
	items map[string]CacheItem
}

type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewAnalysisCache() *AnalysisCache {
	cache := &AnalysisCache{
		items: make(map[string]CacheItem),
	}

	// Temporarily disable cleanup for testing
	// go cache.cleanup()

	return cache
}

func (c *AnalysisCache) GetFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (c *AnalysisCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	item, exists := c.items[key]

	// DEBUG Logging - Enhanced
	fmt.Printf("CACHE DEBUG: Get key=%s, exists=%t, total_items=%d",
		key[:16], exists, len(c.items))
	if exists {
		fmt.Printf(", expired=%t, now=%s, expires_at=%s\n",
			now.After(item.ExpiresAt),
			now.Format("15:04:05"),
			item.ExpiresAt.Format("15:04:05"))
	} else {
		fmt.Printf("\n")
		// List all keys for debugging
		if len(c.items) > 0 {
			fmt.Printf("CACHE DEBUG: Available keys: ")
			for k := range c.items {
				fmt.Printf("%s ", k[:16])
			}
			fmt.Printf("\n")
		}
	}

	if !exists || now.After(item.ExpiresAt) {
		return nil, false
	}

	return item.Data, true
}

func (c *AnalysisCache) Set(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiresAt := time.Now().Add(ttl)

	c.items[key] = CacheItem{
		Data:      data,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	// Enhanced DEBUG logging
	fmt.Printf("CACHE DEBUG: Set key=%s, expires_at=%s, ttl=%s, total_items=%d\n",
		key[:16], expiresAt.Format("15:04:05"), ttl, len(c.items))

	// List all keys after set
	fmt.Printf("CACHE DEBUG: All keys after set: ")
	for k := range c.items {
		fmt.Printf("%s ", k[:16])
	}
	fmt.Printf("\n")
}

func (c *AnalysisCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Clean every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			itemsRemoved := 0

			for key, item := range c.items {
				if now.After(item.ExpiresAt) {
					delete(c.items, key)
					itemsRemoved++
				}
			}

			// Debug logging
			fmt.Printf("CACHE CLEANUP: Removed %d expired items, %d items remaining\n",
				itemsRemoved, len(c.items))

			c.mu.Unlock()
		}
	}
}

func (c *AnalysisCache) GetWithMetrics(key string, metrics interface{}) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists || time.Now().After(item.ExpiresAt) {
		// Cache Miss - Metrics interface würde hier genutzt
		return nil, false
	}

	// Cache Hit - Metrics interface würde hier genutzt
	return item.Data, true
}
