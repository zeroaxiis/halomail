// Package contact exposes Mount so the monolith can host it in-process,
// including the embeddable widget and an in-memory/Redis rate limiter.
package contact

import (
	"context"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	contactv1connect "github.com/aashishrajdev/halomail/services/shared/gen/halomail/contact/v1/contactv1connect"
	identityv1connect "github.com/aashishrajdev/halomail/services/shared/gen/halomail/identity/v1/identityv1connect"
	"github.com/aashishrajdev/halomail/services/shared/authn"
	"github.com/aashishrajdev/halomail/services/shared/config"
	"github.com/aashishrajdev/halomail/services/shared/ratelimit"
	rds "github.com/aashishrajdev/halomail/services/shared/redis"

	"github.com/aashishrajdev/halomail/services/contact/internal/adapters/postgres"
	"github.com/aashishrajdev/halomail/services/contact/internal/adapters/rpc"
	"github.com/aashishrajdev/halomail/services/contact/internal/app"
	"github.com/aashishrajdev/halomail/services/contact/internal/domain"
	"github.com/aashishrajdev/halomail/services/contact/internal/web"
)

type Deps struct {
	Pool         *pgxpool.Pool
	JWTSecret    string
	Redis        *rds.Client // nil → in-memory rate limiter
	Rate         config.Rate
	Logger       *slog.Logger
	Interceptors []connect.Interceptor
	IdentityURL  string
}

func Mount(mux *http.ServeMux, d Deps) {
	limiter := ratelimit.New(d.Redis, ratelimit.Config{
		RPS:   float64(d.Rate.PublicRPS),
		Burst: d.Rate.PublicBurst,
	})
	svc := app.New(app.Repos{
		Forms:    postgres.NewForms(d.Pool),
		Messages: postgres.NewMessages(d.Pool),
	}, limiter, logForwarder{logger: d.Logger})

	identClient := identityv1connect.NewApiKeyServiceClient(http.DefaultClient, d.IdentityURL)
	h := rpc.NewHandlers(svc, authn.NewVerifier(d.JWTSecret), identClient)
	opts := connect.WithInterceptors(d.Interceptors...)

	mux.Handle("/widget.js", web.WidgetHandler())
	mux.Handle(contactv1connect.NewFormServiceHandler(h, opts))
	mux.Handle(contactv1connect.NewMessageServiceHandler(h, opts))
}

// logForwarder is the default Forwarder used in monolith mode.
type logForwarder struct{ logger *slog.Logger }

func (f logForwarder) Forward(ctx context.Context, form *domain.Form, msg *domain.Message) error {
	f.logger.InfoContext(ctx, "forwarding contact message",
		"form", form.Slug, "to", form.TargetEmail, "message_id", msg.ID, "spam", msg.IsSpam)
	return nil
}
