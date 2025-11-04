// File: internal/proxy/handler.go
package proxy

import (
	"bytes"
	"context"
	"go-caching-proxy/internal/cache"
	"go-caching-proxy/internal/metrics"
	"go-caching-proxy/internal/key"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// Use a custom type for our context key to avoid collisions.
type contextKey string

const cacheKeyContextKey = contextKey("cacheKey")

type Handler struct {
	target     *url.URL
	proxy      *httputil.ReverseProxy
	cache      cache.Storer
	defaultTTL time.Duration
	logger     *slog.Logger
	metrics *metrics.Metrics 
}

func NewHandler(target string, cache cache.Storer, defaultTTL time.Duration, logger *slog.Logger, mets *metrics.Metrics) (*Handler, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		target:     targetURL,
		cache:      cache,
		defaultTTL: defaultTTL,
		logger:     logger.With("component", "proxy_handler"),
		metrics: mets,
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ModifyResponse = h.modifyResponse
	h.proxy = proxy

	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.proxy.ServeHTTP(w, r)
		return
	}

	// === THE FIX - PART 1 ===
	// Generate the key once here.
	cacheKey := key.Generate(r)
	log := h.logger.With("cache_key", cacheKey, "path", r.URL.Path)

	if entry, found := h.cache.Get(cacheKey); found {
		log.Info("cache hit")
		h.metrics.CacheHits.Inc()
		h.writeCachedResponse(w, entry)
		return
	}

	log.Info("cache miss")
	h.metrics.CacheMisses.Inc() 

	// === THE FIX - PART 2 ===
	// Store the consistent key in the request's context before forwarding it.
	ctx := context.WithValue(r.Context(), cacheKeyContextKey, cacheKey)
	h.proxy.ServeHTTP(w, r.WithContext(ctx))
}

func (h *Handler) modifyResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Cache-Control") == "no-store" {
		return nil
	}

	// === THE FIX - PART 3 ===
	// Retrieve the consistent key from the context.
	cacheKey, ok := resp.Request.Context().Value(cacheKeyContextKey).(string)
	if !ok {
		// If the key is not in the context, something is wrong. Don't cache.
		return nil
	}
	log := h.logger.With("cache_key", cacheKey, "status", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("failed to read response body", "error", err)
		return err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	entry := cache.CacheEntry{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       body,
		ExpiresAt:  time.Now().Add(h.defaultTTL),
	}

	h.cache.Set(cacheKey, entry)
	log.Info("response cached successfully")
	return nil
}

func (h *Handler) writeCachedResponse(w http.ResponseWriter, entry *cache.CacheEntry) {
	for key, values := range entry.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(entry.StatusCode)
	w.Write(entry.Body)
}
