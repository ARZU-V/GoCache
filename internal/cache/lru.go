// File: internal/cache/lru.go
package cache

import (
	"container/list"
	"sync"
	"time"
)

// LRUCache is a thread-safe, memory-bounded LRU cache implementation.
// It fulfills the Storer interface.
type LRUCache struct {
	maxSize int
	ll      *list.List // Doubly-linked list to track usage order (front=most recent, back=least recent)
	items   map[string]*list.Element // Map for fast key-based lookups
	mu      sync.Mutex
}

// lruEntry is the internal wrapper stored in the linked list.
// It holds the value and the key, so we can delete from the map during eviction.
type lruEntry struct {
	key   string
	value CacheEntry
}

// NewLRUCache creates a new LRUCache with a given size.
func NewLRUCache(size int) *LRUCache {
	if size <= 0 {
		size = 1 // Ensure the cache is usable
	}
	return &LRUCache{
		maxSize: size,
		ll:      list.New(),
		items:   make(map[string]*list.Element),
	}
}

// Set adds or updates a key-value pair.
func (c *LRUCache) Set(key string, value CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If the item already exists, update its value and move it to the front.
	if elem, ok := c.items[key]; ok {
		c.ll.MoveToFront(elem)
		elem.Value.(*lruEntry).value = value
		return
	}

	// If the cache is full, evict the least recently used item (from the back of the list).
	if c.ll.Len() >= c.maxSize {
		c.evict()
	}

	// Add the new item to the front of the list and to the map.
	newElem := c.ll.PushFront(&lruEntry{key: key, value: value})
	c.items[key] = newElem
}

// Get retrieves a value by its key.
func (c *LRUCache) Get(key string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, false
	}

	entry := elem.Value.(*lruEntry).value

	// Check for TTL expiration. This is "lazy eviction".
	if time.Now().After(entry.ExpiresAt) {
		// Item expired, remove it and report a miss.
		c.ll.Remove(elem)
		delete(c.items, key)
		return nil, false
	}

	// This item was just accessed, so move it to the front to mark it as most recently used.
	c.ll.MoveToFront(elem)
	return &entry, true
}

// Delete removes an item from the cache.
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.ll.Remove(elem)
		delete(c.items, key)
	}
}

// evict removes the least recently used item. Must be called with the lock held.
func (c *LRUCache) evict() {
	elem := c.ll.Back()
	if elem != nil {
		c.ll.Remove(elem)
		entry := elem.Value.(*lruEntry)
		delete(c.items, entry.key)
	}
}