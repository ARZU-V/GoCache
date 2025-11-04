// File: cmd/server/main.go
package main

import (
	"flag"
	"go-caching-proxy/internal/admin"
	"go-caching-proxy/internal/cache"
	"go-caching-proxy/internal/config"
	"go-caching-proxy/internal/middleware"
	"go-caching-proxy/internal/metrics"
	"go-caching-proxy/internal/proxy"
	"go-caching-proxy/internal/server"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// initCache is a helper function to initialize the cache based on config.
// It returns the Storer interface, so the rest of the app doesn't
// care about the concrete implementation.
func initCache(cfg *config.Config, logger *slog.Logger) (cache.Storer, error) {
	switch cfg.Cache.CacheType {
	case "redis":
		logger.Info("initializing Redis cache")
		return cache.NewRedisCache(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	
	case "lru":
		logger.Info("initializing LRU in-memory cache")
		return cache.NewLRUCache(cfg.Cache.LRU.Size), nil
	
	default:
		logger.Info("no cache_type specified, defaulting to LRU")
		return cache.NewLRUCache(cfg.Cache.LRU.Size), nil
	}
}

func main() {
	// --- 1. Initialization ---
	configPath := flag.String("config", "configs/config.yaml", "Path to the configuration file")
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load configuration", "path", *configPath, "error", err)
		os.Exit(1)
	}
	logger.Info("configuration loaded successfully")

	// --- 2. Dependency Injection (Wiring) ---

	// Create metrics collectors
	mets := metrics.New()

	// Initialize the cache using our new helper function
	appCache, err := initCache(cfg, logger)
	if err != nil {
		logger.Error("failed to initialize cache", "error", err)
		os.Exit(1)
	}

	// Create the core proxy handler, injecting the cache
	proxyHandler, err := proxy.NewHandler(cfg.Proxy.Target, appCache, cfg.GetDefaultTTL(), logger, mets)
	if err != nil {
		logger.Error("failed to create proxy handler", "error", err)
		os.Exit(1)
	}

	// ... (rest of main.go is unchanged)
	mainMux := http.NewServeMux()
	finalHandler := middleware.Metrics(mets, middleware.Logging(logger, proxyHandler))
	mainMux.Handle("/", finalHandler)
	mainMux.HandleFunc("/healthz", admin.HealthzHandler)

	go func() {
		adminMux := http.NewServeMux()
		adminMux.Handle("/metrics", promhttp.Handler())
		adminPort := "9090"
		logger.Info("starting admin server", "port", adminPort)
		if err := http.ListenAndServe(":"+adminPort, adminMux); err != nil {
			logger.Error("admin server failed", "error", err)
		}
	}()

	srv := server.New(cfg.Server.Port, logger)
	if err := srv.Start(mainMux); err != nil {
		logger.Error("main server failed to start", "error", err)
		os.Exit(1)
	}
}