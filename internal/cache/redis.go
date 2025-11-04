// File: internal/cache/redis.go
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a cache implementation that uses Redis as the backend.
// It satisfies the Storer interface.
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new connection to Redis and returns a RedisCache.
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Ping the server to ensure a connection is established.
	ctx := context.Background()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// Get retrieves an entry from Redis.
func (c *RedisCache) Get(key string) (*CacheEntry, bool) {
	// Fetch the value (which is a JSON string) from Redis.
	val, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return nil, false // Cache miss
	} else if err != nil {
		return nil, false // Some other error
	}

	// Value found, deserialize the JSON string back into a CacheEntry struct.
	var entry CacheEntry
	if err := json.Unmarshal([]byte(val), &entry); err != nil {
		return nil, false // Failed to parse the data
	}

	// We don't need to check TTL here, as Redis's `Set` command handles expiration for us.
	return &entry, true
}

// Set stores an entry in Redis.
func (c *RedisCache) Set(key string, entry CacheEntry) {
	// Serialize the Go struct into a JSON string.
	data, err := json.Marshal(entry)
	if err != nil {
		return // Don't cache if serialization fails
	}

	// Calculate the cache duration from the entry's expiry time.
	ttl := time.Until(entry.ExpiresAt)
	if ttl <= 0 {
		return // Already expired, don't cache.
	}

	// Set the value in Redis with the calculated TTL.
	c.client.Set(c.ctx, key, data, ttl)
}

// Delete removes an entry from Redis.
func (c *RedisCache) Delete(key string) {
	c.client.Del(c.ctx, key)
}