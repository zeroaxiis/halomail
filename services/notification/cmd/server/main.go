// Command server boots the notification service: transactional email
// (Resend/SMTP) and webhook dispatch with a background delivery worker.
package main

import (
	"context"
	"net/http"
	"os"

	"connectrpc.com/connect"

	"github.com/aashishrajdev/halomail/services/shared/authn"
	"github.com/aashishrajdev/halomail/services/shared/config"
	"github.com/aashishrajdev/halomail/services/shared/connectutil"
	notificationv1connect "github.com/aashishrajdev/halomail/services/shared/gen/halomail/notification/v1/notificationv1connect"
	"github.com/aashishrajdev/halomail/services/shared/health"
	"github.com/aashishrajdev/halomail/services/shared/log"
	"github.com/aashishrajdev/halomail/services/shared/observability"
	pg "github.com/aashishrajdev/halomail/services/shared/postgres"
	"github.com/aashishrajdev/halomail/services/shared/server"

	npg "github.com/aashishrajdev/halomail/services/notification/internal/adapters/postgres"
	"github.com/aashishrajdev/halomail/services/notification/internal/adapters/rpc"
	"github.com/aashishrajdev/halomail/services/notification/internal/app"
	"github.com/aashishrajdev/halomail/services/notification/internal/email"
)

const serviceName = "notification"

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

	// Resend when configured, otherwise SMTP (Mailpit locally).
	var sender email.Sender
	if cfg.Email.UseResend() {
		sender = email.NewResend(cfg.Email.ResendAPIKey, cfg.Email.From)
		logger.Info("boot", "addr", cfg.HTTP.Addr(), "email", "resend")
	} else {
		sender = email.NewSMTP(cfg.Email.SMTPHost, cfg.Email.SMTPPort, cfg.Email.From)
		logger.Info("boot", "addr", cfg.HTTP.Addr(), "email", "smtp")
	}

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

	repos := app.Repos{
		Webhooks:   npg.NewWebhooks(pool),
		Deliveries: npg.NewDeliveries(pool),
	}
	svc := app.New(repos, sender)
	handlers := rpc.NewHandlers(svc, authn.NewVerifier(cfg.Auth.JWTSecret))

	// Background webhook delivery worker.
	go app.NewWorker(repos, logger).Run(ctx)
	go app.RunMailWorker(ctx, pool, sender, logger)

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
	mux.Handle(notificationv1connect.NewWebhookServiceHandler(handlers, opts))

	if err := server.Run(ctx, cfg.HTTP.Addr(), mux, logger); err != nil {
		logger.Error("server stopped with error", "error", err.Error())
		os.Exit(1)
	}
}
