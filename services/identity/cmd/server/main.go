// Command server boots the identity service: config → logging → tracing →
// Postgres → ConnectRPC handlers (auth, users, API keys, audit) → h2c server.
package main

import (
	"context"
	"github.com/aashishrajdev/halomail/services/identity"
	"net/http"
	"os"

	"github.com/aashishrajdev/halomail/services/shared/config"
	"github.com/aashishrajdev/halomail/services/shared/connectutil"
	"github.com/aashishrajdev/halomail/services/shared/health"
	"github.com/aashishrajdev/halomail/services/shared/log"
	"github.com/aashishrajdev/halomail/services/shared/observability"
	pg "github.com/aashishrajdev/halomail/services/shared/postgres"
	"github.com/aashishrajdev/halomail/services/shared/server"
)

const serviceName = "identity"

func main() {
	ctx := context.Background()

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
	logger.Info("boot", "addr", cfg.HTTP.Addr())

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

	interceptors, err := connectutil.Default(logger)
	if err != nil {
		logger.Error("build interceptors", "error", err.Error())
		os.Exit(1)
	}

	hc := health.New()
	hc.Register("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })

	mux := http.NewServeMux()
	mux.Handle("/healthz", hc.Liveness())
	mux.Handle("/readyz", hc.Readiness())

	identity.Mount(mux, identity.Deps{Pool: pool, JWTSecret: cfg.Auth.JWTSecret, SessionTTL: cfg.Auth.SessionTTL, APIKeyPrefix: cfg.Auth.APIKeyPrefix, Interceptors: interceptors, Limits: cfg.Limits})

	if err := server.Run(ctx, cfg.HTTP.Addr(), mux, logger); err != nil {
		logger.Error("server stopped with error", "error", err.Error())
		os.Exit(1)
	}
}
