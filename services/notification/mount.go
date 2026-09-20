// Package notification exposes Mount so the monolith can host it in-process,
// including the background webhook delivery worker.
package notification

import (
	"context"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aashishrajdev/halomail/services/shared/authn"
	"github.com/aashishrajdev/halomail/services/shared/config"
	notificationv1connect "github.com/aashishrajdev/halomail/services/shared/gen/halomail/notification/v1/notificationv1connect"

	"github.com/aashishrajdev/halomail/services/notification/internal/adapters/postgres"
	"github.com/aashishrajdev/halomail/services/notification/internal/adapters/rpc"
	"github.com/aashishrajdev/halomail/services/notification/internal/app"
	"github.com/aashishrajdev/halomail/services/notification/internal/email"
)

type Deps struct {
	Pool         *pgxpool.Pool
	JWTSecret    string
	Email        config.Email
	Logger       *slog.Logger
	Interceptors []connect.Interceptor
}

// Mount registers handlers and starts the delivery worker, which stops when ctx
// is cancelled.
func Mount(ctx context.Context, mux *http.ServeMux, d Deps) {
	var sender email.Sender
	if d.Email.UseResend() {
		sender = email.NewResend(d.Email.ResendAPIKey, d.Email.From)
	} else {
		sender = email.NewSMTP(d.Email.SMTPHost, d.Email.SMTPPort, d.Email.From)
	}

	repos := app.Repos{
		Webhooks:   postgres.NewWebhooks(d.Pool),
		Deliveries: postgres.NewDeliveries(d.Pool),
	}
	svc := app.New(repos, sender)
	h := rpc.NewHandlers(svc, authn.NewVerifier(d.JWTSecret))

	go app.NewWorker(repos, d.Logger).Run(ctx)
	go app.RunMailWorker(ctx, d.Pool, sender, d.Logger)

	opts := connect.WithInterceptors(d.Interceptors...)
	mux.Handle(notificationv1connect.NewWebhookServiceHandler(h, opts))
}
