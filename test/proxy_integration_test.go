// File: test/proxy_integration_test.go
package test

import (
	"go-caching-proxy/internal/cache"
	"go-caching-proxy/internal/metrics" // <-- 1. IMPORT METRICS
	"go-caching-proxy/internal/proxy"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

// TestProxyCacheHitAndMiss is an integration test that verifies
// the cache miss and cache hit logic of our proxy.
func TestProxyCacheHitAndMiss(t *testing.T) {
	// 1. Create a "mock" origin server
	var originHitCount int32
	mockOrigin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&originHitCount, 1)
		w.Write([]byte("hello from origin"))
	}))
	defer mockOrigin.Close()

	// 2. Create our proxy's dependencies
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	appCache := cache.NewLRUCache(10)
	defaultTTL := 1 * time.Minute
	mets := metrics.New() // <-- 2. CREATE A VALID METRICS OBJECT

	// 3. Create the real proxy handler, configured to use our mock server
	//    Pass the 'mets' object instead of 'nil'
	proxyHandler, err := proxy.NewHandler(mockOrigin.URL, appCache, defaultTTL, logger, mets) // <-- 3. PASS METS
	if err != nil {
		t.Fatalf("failed to create proxy handler: %v", err)
	}

	// 4. Start a test server running our proxy
	proxyServer := httptest.NewServer(proxyHandler)
	defer proxyServer.Close()

	// 5. Run Test 1: The Cache Miss
	t.Run("first request is a cache miss", func(t *testing.T) {
		resp, err := http.Get(proxyServer.URL)
		if err != nil {
			t.Fatalf("request to proxy failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if string(body) != "hello from origin" {
			t.Errorf("expected body 'hello from origin', got '%s'", string(body))
		}
		if atomic.LoadInt32(&originHitCount) != 1 {
			t.Errorf("expected origin server to be hit 1 time, got %d", originHitCount)
		}
	})

	// 6. Run Test 2: The Cache Hit
	t.Run("second request is a cache hit", func(t *testing.T) {
		resp, err := http.Get(proxyServer.URL)
		if err != nil {
			t.Fatalf("request to proxy failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if string(body) != "hello from origin" {
			t.Errorf("expected body 'hello from origin', got '%s'", string(body))
		}
		if atomic.LoadInt32(&originHitCount) != 1 {
			t.Errorf("expected origin server to still be hit only 1 time, got %d", originHitCount)
		}
	})
}
