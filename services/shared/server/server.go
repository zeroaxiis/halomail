// Package server runs an h2c HTTP server (HTTP/2 cleartext — required for gRPC
// over plain HTTP in dev) with graceful shutdown on SIGINT/SIGTERM.
package server

import (
	"context"
	"errors"
	"github.com/aashishrajdev/halomail/services/shared/httpx"
	"github.com/aashishrajdev/halomail/services/shared/ratelimit"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Run serves handler on addr and blocks until a termination signal, then
// drains in-flight requests within a 15s deadline.
func Run(ctx context.Context, addr string, handler http.Handler, logger *slog.Logger) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(Protect(handler), &http2.Server{}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		return err
	case <-sigCtx.Done():
		logger.Info("shutdown signal received, draining")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func Protect(next http.Handler) http.Handler {
	public := ratelimit.NewMemory(ratelimit.Config{RPS: 10, Burst: 30})
	login := ratelimit.NewMemory(ratelimit.Config{RPS: 1, Burst: 10})
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("Referrer-Policy", "no-referrer")
		request.Body = http.MaxBytesReader(writer, request.Body, 64<<10)
		if request.Method == http.MethodPost {
			limiter := public
			if strings.HasSuffix(request.URL.Path, "AuthService/Login") || strings.HasSuffix(request.URL.Path, "AuthService/Register") {
				limiter = login
			}
			allowed, err := limiter.Allow(request.Context(), httpx.Peer(request.RemoteAddr))
			if err != nil || !allowed {
				writer.Header().Set("Retry-After", "2")
				httpx.JSON(writer, http.StatusTooManyRequests, map[string]string{"code": "resource_exhausted", "message": "Too many requests; try again shortly"})
				return
			}
		}
		next.ServeHTTP(writer, request)
	})
}
