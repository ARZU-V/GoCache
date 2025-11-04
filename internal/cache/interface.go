// File: internal/cache/interface.go
package cache

import (
	"net/http"
	"time"
)

// CacheEntry represents everything we need to store for a single cached HTTP response.
// By creating a dedicated struct, we ensure our cache stores data in a consistent,
// structured way.
type CacheEntry struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	ExpiresAt  time.Time
}

// Storer is the interface that defines the contract for all cache implementations.
// This is a powerful abstraction that makes our system pluggable.
type Storer interface {
	// Get retrieves a CacheEntry by its key. The boolean indicates if the entry was found.
	Get(key string) (entry *CacheEntry, found bool)

	// Set stores a CacheEntry with a given key.
	Set(key string, entry CacheEntry)
	
	// Delete removes an entry from the cache.
	Delete(key string)
}