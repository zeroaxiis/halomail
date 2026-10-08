package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aashishrajdev/halomail/services/shared/config"
	"github.com/aashishrajdev/halomail/services/shared/connectutil"
	"github.com/aashishrajdev/halomail/services/shared/health"
	"github.com/aashishrajdev/halomail/services/shared/log"
	"github.com/aashishrajdev/halomail/services/shared/observability"
	pg "github.com/aashishrajdev/halomail/services/shared/postgres"
	"github.com/aashishrajdev/halomail/services/shared/ratelimit"
	rds "github.com/aashishrajdev/halomail/services/shared/redis"
	"github.com/aashishrajdev/halomail/services/shared/server"

	"github.com/aashishrajdev/halomail/services/contact"
	"github.com/aashishrajdev/halomail/services/identity"
	"github.com/aashishrajdev/halomail/services/notification"
	"github.com/aashishrajdev/halomail/services/scheduling"
	"github.com/aashishrajdev/halomail/services/template"
)

const serviceName = "gateway"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.New(log.Options{Service: serviceName}).Error("load config", "error", err.Error())
		os.Exit(1)
	}

	logger := log.New(log.Options{
		Level:   cfg.App.LogLevel,
		Format:  cfg.App.LogFormat,
		Service: serviceName,
		Env:     cfg.App.Env,
	})
	log.SetDefault(logger)
	logger.Info("boot", "addr", cfg.HTTP.Addr(), "mode", "monolith")

	shutdown, err := observability.Setup(ctx, observability.Config{
		Enabled:     cfg.OTel.Enabled,
		Endpoint:    cfg.OTel.Endpoint,
		ServiceName: serviceName,
		SampleRatio: cfg.OTel.SampleRatio,
	})
	if err != nil {
		logger.Warn("otel setup failed; continuing without tracing", "error", err.Error())
	} else {
		defer func() { _ = shutdown(ctx) }()
	}

	pool, err := pg.New(ctx, cfg.Postgres.URL)
	if err != nil {
		logger.Error("connect postgres", "error", err.Error())
		os.Exit(1)
	}
	defer pool.Close()

	var redisClient *rds.Client
	if cfg.Redis.URL != "" {
		if c, rerr := rds.New(ctx, cfg.Redis.URL); rerr != nil {
			logger.Warn("redis unavailable; using in-memory rate limiter", "error", rerr.Error())
		} else {
			redisClient = c
			defer func() { _ = redisClient.Close() }()
		}
	}

	interceptors, err := connectutil.Default(logger)
	if err != nil {
		logger.Error("build interceptors", "error", err.Error())
		os.Exit(1)
	}

	mux := http.NewServeMux()

	hc := health.New()
	hc.Register("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })
	mux.Handle("/healthz", hc.Liveness())
	mux.Handle("/readyz", hc.Readiness())
	mux.HandleFunc("/", root)
	mux.HandleFunc("/health", healthHandler)

	// Host every service in-process.
	identity.Mount(mux, identity.Deps{
		OTPDeliverySecret: cfg.Auth.OTPDeliverySecret,
		Pool:              pool, JWTSecret: cfg.Auth.JWTSecret, SessionTTL: cfg.Auth.SessionTTL,
		APIKeyPrefix: cfg.Auth.APIKeyPrefix, Redis: redisClient, Interceptors: interceptors,
		Limits: cfg.Limits,
	})
	scheduling.Mount(mux, scheduling.Deps{
		Pool: pool, JWTSecret: cfg.Auth.JWTSecret,
		Google: cfg.Google, Microsoft: cfg.Microsoft, Interceptors: interceptors,
		Limits: cfg.Limits, EncryptionKey: cfg.Auth.CalendarEncryptionKey, WebURL: cfg.App.PublicWebURL, Context: ctx, Logger: logger,
	})
	contact.Mount(mux, contact.Deps{
		Pool: pool, JWTSecret: cfg.Auth.JWTSecret, Redis: redisClient,
		Rate: cfg.Rate, Logger: logger, Interceptors: interceptors,
		IdentityURL: "http://127.0.0.1:" + fmt.Sprint(cfg.HTTP.Port), Limits: cfg.Limits,
	})
	template.Mount(mux, template.Deps{
		Pool: pool, JWTSecret: cfg.Auth.JWTSecret, Interceptors: interceptors,
	})
	notification.Mount(ctx, mux, notification.Deps{
		Pool: pool, JWTSecret: cfg.Auth.JWTSecret, Email: cfg.Email,
		Logger: logger, Interceptors: interceptors,
	})

	// Initialize Gateway global rate limiter
	limiter := ratelimit.New(redisClient, ratelimit.Config{
		RPS:   float64(cfg.Rate.PublicRPS) * 2, // Slightly higher for global
		Burst: cfg.Rate.PublicBurst * 2,
	})

	if err := server.Run(ctx, cfg.HTTP.Addr(), withRateLimit(withCORS(mux, cfg.App.PublicWebURL), limiter), logger); err != nil {
		logger.Error("server stopped with error", "error", err.Error())
		os.Exit(1)
	}
}

func root(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"service": "halomail-gateway",
		"mode":    "monolith",
		"status":  "ok",
	})
}

// withCORS allows the public surface (booking page, contact widget, SDK) to be
// called cross-origin, restricted to the official frontend domain.
func withCORS(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		
		// User requested using the API environment variable
		strictDomain := os.Getenv("API")
		if strictDomain != "" {
			if !strings.HasPrefix(strictDomain, "http") {
				strictDomain = "https://" + strictDomain
			}
			h.Set("Access-Control-Allow-Origin", strictDomain)
		} else if allowedOrigin != "" {
			h.Set("Access-Control-Allow-Origin", allowedOrigin)
		} else {
			h.Set("Access-Control-Allow-Origin", "http://localhost:3000") // Fallback
		}
		h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Connect-Protocol-Version, Connect-Timeout-Ms, X-Requested-With, X-HaloMail-Key")
		h.Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// withRateLimit globally protects all gateway endpoints by IP
func withRateLimit(next http.Handler, limiter ratelimit.Limiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}
		key := "gateway:ip:" + ip
		if ok, _ := limiter.Allow(r.Context(), key); !ok {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
