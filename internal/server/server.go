// File: internal/server/server.go
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server is the main struct for our web server.
type Server struct {
	port   string
	logger *slog.Logger
}

// New creates a new instance of our server.
func New(port string, logger *slog.Logger) *Server {
	return &Server{
		port:   port,
		logger: logger.With("component", "http_server"),
	}
}

// Start runs the HTTP server and includes the graceful shutdown logic.
func (s *Server) Start(handler http.Handler) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.port),
		Handler: handler,
	}

	// Create a context that listens for the interrupt signal from the OS.
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the server in a separate goroutine.
	go func() {
		s.logger.Info("server is starting", "port", s.port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("listen and serve failed", "error", err)
		}
	}()

	// Wait here until the shutdown signal is received.
	<-shutdownCtx.Done()

	s.logger.Info("shutdown signal received, starting graceful shutdown")

	// Create a context with a timeout to allow existing requests to finish.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server.
	if err := srv.Shutdown(timeoutCtx); err != nil {
		s.logger.Error("graceful shutdown failed", "error", err)
		return err
	}

	s.logger.Info("server has shut down gracefully")
	return nil
}