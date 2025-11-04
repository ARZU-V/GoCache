// File: internal/metrics/prometheus.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all the Prometheus metrics for the application.
type Metrics struct {
	CacheHits   prometheus.Counter
	CacheMisses prometheus.Counter
	CacheSize   prometheus.Gauge
	Latency     prometheus.Histogram
}

// New creates and registers the Prometheus metrics.
func New() *Metrics {
	return &Metrics{
		CacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "proxy_cache_hits_total",
			Help: "The total number of cache hits",
		}),
		CacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "proxy_cache_misses_total",
			Help: "The total number of cache misses",
		}),
		CacheSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "proxy_cache_size_items",
			Help: "The current number of items in the cache",
		}),
		Latency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "proxy_request_duration_seconds",
			Help:    "A histogram of the request latency.",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10), // 10 buckets, 0.1s width
		}),
	}
}
