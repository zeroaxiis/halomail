// Command server boots the contact service: forms, public submissions with
// spam protection + rate limiting, message storage, and the embeddable widget.
package main

import (
	"context"
	"net/http"
	"os"

	"connectrpc.com/connect"

	"github.com/zeroaxiis/halomail/services/shared/authn"
	"github.com/zeroaxiis/halomail/services/shared/config"
	"github.com/zeroaxiis/halomail/services/shared/connectutil"
	contactv1connect "github.com/zeroaxiis/halomail/services/shared/gen/halomail/contact/v1/contactv1connect"
	identityv1connect "github.com/zeroaxiis/halomail/services/shared/gen/halomail/identity/v1/identityv1connect"
	"github.com/zeroaxiis/halomail/services/shared/health"
	"github.com/zeroaxiis/halomail/services/shared/log"
	"github.com/zeroaxiis/halomail/services/shared/observability"
	pg "github.com/zeroaxiis/halomail/services/shared/postgres"
	"github.com/zeroaxiis/halomail/services/shared/ratelimit"
	rds "github.com/zeroaxiis/halomail/services/shared/redis"
	"github.com/zeroaxiis/halomail/services/shared/server"

	cpg "github.com/zeroaxiis/halomail/services/contact/internal/adapters/postgres"
	"github.com/zeroaxiis/halomail/services/contact/internal/adapters/rpc"
	"github.com/zeroaxiis/halomail/services/contact/internal/app"
	"github.com/zeroaxiis/halomail/services/contact/internal/web"
)

const serviceName = "contact"

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

	// Redis is optional — fall back to an in-memory rate limiter.
	var redisClient *rds.Client
	if cfg.Redis.URL != "" {
		if c, rerr := rds.New(ctx, cfg.Redis.URL); rerr != nil {
			logger.Warn("redis unavailable; using in-memory rate limiter", "error", rerr.Error())
		} else {
			redisClient = c
			defer func() { _ = redisClient.Close() }()
		}
	}
	limiter := ratelimit.New(redisClient, ratelimit.Config{
		RPS:   float64(cfg.Rate.PublicRPS),
		Burst: cfg.Rate.PublicBurst,
	})

	svc := app.New(app.Repos{
		Forms:    cpg.NewForms(pool),
		Messages: cpg.NewMessages(pool, cfg.Limits),
	}, limiter, nil, redisClient)

	identityURL := cfg.App.PublicAPIURL
	identClient := identityv1connect.NewApiKeyServiceClient(http.DefaultClient, identityURL)
	handlers := rpc.NewHandlers(svc, authn.NewVerifier(cfg.Auth.JWTSecret), identClient)

	interceptors, err := connectutil.Default(logger)
	if err != nil {
		logger.Error("build interceptors", "error", err.Error())
		os.Exit(1)
	}
	opts := connect.WithInterceptors(interceptors...)

	hc := health.New()
	hc.Register("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })

	mux := http.NewServeMux()
	mux.Handle("/healthz", hc.Liveness())
	mux.Handle("/readyz", hc.Readiness())
	mux.Handle("/widget.js", web.WidgetHandler())
	mux.HandleFunc("POST /submit", web.SubmitHandler(svc, pool))
	mux.Handle(contactv1connect.NewFormServiceHandler(handlers, opts))
	mux.Handle(contactv1connect.NewMessageServiceHandler(handlers, opts))

	if err := server.Run(ctx, cfg.HTTP.Addr(), mux, logger); err != nil {
		logger.Error("server stopped with error", "error", err.Error())
		os.Exit(1)
	}
}
