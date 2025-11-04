# Go Caching Proxy

A high-performance, production-ready HTTP reverse proxy with a pluggable caching layer, built in Go. This project is designed to demonstrate best practices in system design, concurrency, testing, and observability.

[Image of a Grafana dashboard showing cache metrics]

## ✨ Features

* **High Performance:** Built in Go, using goroutines and a highly concurrent design.
* **Pluggable Cache:** Switch between a local **in-memory LRU cache** or a shared **Redis cache** via a simple config change.
* **Full Observability:** Pre-configured stack with **Prometheus** for metrics collection and **Grafana** for real-time dashboards.
* **Production Ready:** Features structured JSON logging, graceful shutdown, and a complete containerized environment via Docker Compose.
* **Robust Design:** A modular, interface-based architecture that is clean, testable, and easy to extend.

## 🚀 Quick Start (Recommended)

This project is best run using Docker Compose, which starts the proxy, Redis, Prometheus, and Grafana all at once.

**Prerequisites:**
* Docker & Docker Compose

**Run the Stack:**
1.  Navigate to the `deployments/` directory:
    ```bash
    cd deployments
    ```
2.  Bring the entire stack online:
    ```bash
    docker-compose up --build -d
    ```

**Your environment is now running:**
* **Go Proxy:** `http://localhost:8080`
* **Grafana Dashboard:** `http://localhost:3000` (Login: `admin` / `admin`)
* **Prometheus:** `http://localhost:9091`
* **Proxy Metrics:** `http://localhost:9090/metrics`

## 🏃‍♂️ Run Locally (For Go Development)

**Prerequisites:**
* Go 1.25+
* A running Redis server (if using Redis mode)

1.  **Install dependencies:**
    ```bash
    go mod tidy
    ```
2.  **Edit `configs/config.yaml`:**
    * Set `cache_type` to `"lru"` or `"redis"`.
    * If using Redis, ensure the `redis.address` is correct (e.g., `localhost:6379`).
3.  **Run the server:**
    ```bash
    make run
    ```

## 🧪 Running Tests

A full integration test suite is included to validate caching logic.

```bash
# From the project root
go test ./...