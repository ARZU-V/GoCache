# Usage and Configuration Guide

This guide provides detailed instructions on how to configure, run, and interact with the Go Caching Proxy.

## Configuration

All configuration is managed via the `configs/config.yaml` file.

```yaml
# Settings for the HTTP server itself
server:
  port: "8080" # Port for main user-facing traffic

# Settings for the reverse proxy behavior
proxy:
  # The backend server to forward requests to
  target: "[https://httpbin.org](https://httpbin.org)"

# Settings for the caching layer
cache:
  # Type can be "lru" or "redis"
  cache_type: "redis"
  
  lru:
    # Max number of items for the in-memory cache
    size: 100
  
  # Default cache duration in seconds
  default_ttl_seconds: 60

# Settings for Redis
redis:
  # The address for the redis server.
  # 'redis:6379' works in Docker Compose.
  # Use 'localhost:6379' for local dev.
  address: "redis:6379" 
  password: "" # No password for our local dev instance
  db: 0