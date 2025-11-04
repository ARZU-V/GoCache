# Design Decisions & Trade-offs

This document outlines the key engineering decisions and trade-offs made during the project's development.

### 1. Architecture: Interface-Based Caching
* **Decision:** To define a generic `Storer` interface (`Get`, `Set`, `Delete`) for the cache, rather than coding the proxy to talk directly to the LRU cache.
* **Rationale:** This decouples the proxy logic from the cache implementation. It allows the system to be flexible and "pluggable."
* **Trade-off:** This adds a small layer of abstraction, but the benefit is immense. It allowed us to add a `RedisCache` implementation later without changing a single line of code in the core `proxy.Handler`.

### 2. Concurrency: `sync.Mutex` vs. `sync.RWMutex` in LRU Cache
* **Decision:** To use a standard `sync.Mutex` for the `LRUCache`, even though `Get` operations seem like "reads."
* **Rationale:** A `Get` in an LRU cache is not a pure read. It **mutates** the internal linked list by moving the accessed item to the front. Since both `Get` and `Set` are write operations on the list, a simple `sync.Mutex` is the correct, safer, and less complex choice than a Read-Write mutex.

### 3. Cache Implementation: LRU vs. Redis
* **Decision:** To implement both a local in-memory (LRU) cache and a distributed (Redis) cache, selectable via configuration.
* **Rationale:** This demonstrates a deep understanding of scaling and performance trade-offs.
* **Trade-off:**
    * **LRU:** Blisteringly fast (tens of thousands of req/s), as it's just a local memory access. However, it cannot be shared, leading to a low cache hit-rate when the proxy is scaled out to multiple instances.
    * **Redis:** Measurably slower (due to network I/O and JSON serialization), but it provides a **central, shared cache**. This allows 1,000 proxy instances to share one cache "brain," dramatically increasing the overall cache hit-rate and system scalability.

### 4. Observability: Separate Admin Server
* **Decision:** To run the `/metrics` and `/healthz` endpoints on a separate port (`9090`) in a separate goroutine.
* **Rationale:** This is a production-ready pattern. It separates internal "admin" traffic from public "user" traffic. It ensures that a flood of user requests cannot starve the `/metrics` endpoint, which is critical for monitoring tools.
* **Trade-off:** A tiny increase in code complexity for a massive gain in reliability.

### 5. Production Readiness: Graceful Shutdown
* **Decision:** To implement graceful shutdown using `os.Signal` and `context.WithTimeout`.
* **Rationale:** In a real deployment, simply killing a server (`Ctrl+C`) would terminate active user connections and cause errors. Graceful shutdown tells the server to stop accepting *new* requests but allows existing, in-flight requests to finish, enabling zero-downtime deployments.