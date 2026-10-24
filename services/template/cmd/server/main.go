// Command server boots the template service: built-in email themes, custom
// templates, and live preview rendering over ConnectRPC.
package main

import (
	"context"
	"net/http"
	"os"

	"connectrpc.com/connect"

	"github.com/zeroaxiis/halomail/services/shared/authn"
	"github.com/zeroaxiis/halomail/services/shared/config"
	"github.com/zeroaxiis/halomail/services/shared/connectutil"
	templatev1connect "github.com/zeroaxiis/halomail/services/shared/gen/halomail/template/v1/templatev1connect"
	"github.com/zeroaxiis/halomail/services/shared/health"
	"github.com/zeroaxiis/halomail/services/shared/log"
	"github.com/zeroaxiis/halomail/services/shared/observability"
	pg "github.com/zeroaxiis/halomail/services/shared/postgres"
	"github.com/zeroaxiis/halomail/services/shared/server"

	tpg "github.com/zeroaxiis/halomail/services/template/internal/adapters/postgres"
	"github.com/zeroaxiis/halomail/services/template/internal/adapters/rpc"
	"github.com/zeroaxiis/halomail/services/template/internal/app"
)

const serviceName = "template"

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

	svc := app.New(app.Repos{Templates: tpg.NewTemplates(pool)})
	handlers := rpc.NewHandlers(svc, authn.NewVerifier(cfg.Auth.JWTSecret))

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
	mux.Handle(templatev1connect.NewTemplateServiceHandler(handlers, opts))

	if err := server.Run(ctx, cfg.HTTP.Addr(), mux, logger); err != nil {
		logger.Error("server stopped with error", "error", err.Error())
		os.Exit(1)
	}
}
