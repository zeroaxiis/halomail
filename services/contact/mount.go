// Package contact exposes Mount so the monolith can host it in-process,
// including the embeddable widget and an in-memory/Redis rate limiter.
package contact

import (
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aashishrajdev/halomail/services/shared/authn"
	"github.com/aashishrajdev/halomail/services/shared/config"
	contactv1connect "github.com/aashishrajdev/halomail/services/shared/gen/halomail/contact/v1/contactv1connect"
	identityv1connect "github.com/aashishrajdev/halomail/services/shared/gen/halomail/identity/v1/identityv1connect"
	"github.com/aashishrajdev/halomail/services/shared/ratelimit"
	rds "github.com/aashishrajdev/halomail/services/shared/redis"
	"github.com/aashishrajdev/halomail/services/shared/usage"

	"github.com/aashishrajdev/halomail/services/contact/internal/adapters/postgres"
	"github.com/aashishrajdev/halomail/services/contact/internal/adapters/rpc"
	"github.com/aashishrajdev/halomail/services/contact/internal/app"
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
	Limits       usage.Policy
}

func Mount(mux *http.ServeMux, d Deps) {
	limiter := ratelimit.New(d.Redis, ratelimit.Config{
		RPS:   float64(d.Rate.PublicRPS),
		Burst: d.Rate.PublicBurst,
	})
	svc := app.New(app.Repos{
		Forms:    postgres.NewForms(d.Pool),
		Messages: postgres.NewMessages(d.Pool, d.Limits),
	}, limiter, nil)

	identClient := identityv1connect.NewApiKeyServiceClient(http.DefaultClient, d.IdentityURL)
	h := rpc.NewHandlers(svc, authn.NewVerifier(d.JWTSecret), identClient)
	opts := connect.WithInterceptors(d.Interceptors...)

	mux.Handle("/widget.js", web.WidgetHandler())
	mux.HandleFunc("POST /submit", web.SubmitHandler(svc, d.Pool))
	mux.Handle(contactv1connect.NewFormServiceHandler(h, opts))
	mux.Handle(contactv1connect.NewMessageServiceHandler(h, opts))
}
