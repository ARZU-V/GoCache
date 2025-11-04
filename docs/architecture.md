
### 2. `docs/architecture.md` (The "Blueprint")

This document explains the components and how they fit together.

```markdown
# Architecture Overview

This system is designed as a set of decoupled, containerized services that work together to provide a single, cohesive application. The core application is the `proxy` service, which is supported by `redis` for caching and the `prometheus`/`grafana` stack for observability.



## Components

### 1. The Proxy Service
The Go application itself. It's composed of several internal modules:
* **Server (`internal/server`):** The main web server. It's responsible for handling TCP connections, routing, graceful shutdown, and chaining middleware.
* **Proxy Handler (`internal/proxy`):** The core logic. It receives requests, generates a cache key, and orchestrates the cache-or-fetch decision. It uses the standard library's `httputil.ReverseProxy` and hooks into its `ModifyResponse` function to save responses to the cache.
* **Cache (`internal/cache`):** A modular caching backend. It is defined by a single **`Storer` interface**, which provides `Get`, `Set`, and `Delete` methods.
    * **`LRUCache`:** An in-memory, thread-safe LRU cache implementation. Fast but local to each proxy instance.
    * **`RedisCache`:** A distributed cache implementation. It serializes `CacheEntry` structs to JSON before storing them in Redis, allowing multiple proxy instances to share a single cache.
* **Admin Server:** A separate, lightweight server started as a goroutine. It runs on a different port (`9090`) and exposes internal endpoints like `/healthz` and `/metrics` so that monitoring traffic doesn't interfere with user traffic.

### 2. Redis
A containerized Redis instance that serves as the distributed cache. The Go proxy connects to it using the `redis:6379` internal Docker network address.

### 3. Prometheus
A containerized Prometheus instance. It is configured to "scrape" (poll) the `/metrics` endpoint of the `proxy` service every 15 seconds, collecting all the exported metrics.

### 4. Grafana
A containerized Grafana instance. It is pre-configured to use Prometheus as its data source, allowing you to build dashboards by querying the metrics that Prometheus has collected.

## Request Data Flow

### Cache Miss
1.  A `GET` request hits the proxy on `localhost:8080`.
2.  The `Metrics` and `Logging` middleware execute.
3.  The `Proxy Handler` generates a unique cache key (e.g., `GET|localhost:8080|/uuid`).
4.  The handler calls `cache.Get(key)` on the `Storer` interface.
5.  The cache (LRU or Redis) reports a miss.
6.  The `CacheMisses` counter in Prometheus is incremented.
7.  The request is forwarded to the origin server (`httpbin.org`).
8.  The origin responds. The proxy's `ModifyResponse` hook intercepts this response.
9.  The response body is read, and a `CacheEntry` is created.
10. The handler calls `cache.Set(key, entry)`, saving the entry to Redis/LRU with a TTL.
11. The proxy streams the response back to the client.

### Cache Hit
1.  A `GET` request hits the proxy.
2.  Middleware executes.
3.  The `Proxy Handler` generates the *same* cache key.
4.  `cache.Get(key)` is called. The cache (LRU or Redis) finds the entry.
5.  The `CacheHits` counter in Prometheus is incremented.
6.  The cached response (headers and body) is reconstructed and sent *immediately* to the client.
7.  The origin server is **never contacted**.