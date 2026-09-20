// Package identity exposes Mount so the monolith (gateway) can host the
// identity service in-process. Standalone deployment uses cmd/server.
package identity

import (
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	identityv1connect "github.com/aashishrajdev/halomail/services/shared/gen/halomail/identity/v1/identityv1connect"

	"github.com/aashishrajdev/halomail/services/identity/internal/adapters/postgres"
	"github.com/aashishrajdev/halomail/services/identity/internal/adapters/rpc"
	"github.com/aashishrajdev/halomail/services/identity/internal/app"
	"github.com/aashishrajdev/halomail/services/shared/authn"
	"github.com/aashishrajdev/halomail/services/shared/httpx"
	"github.com/aashishrajdev/halomail/services/shared/usage"
)

type Deps struct {
	Pool         *pgxpool.Pool
	JWTSecret    string
	SessionTTL   time.Duration
	APIKeyPrefix string
	Interceptors []connect.Interceptor
	Limits       usage.Policy
}

// Mount registers the identity ConnectRPC handlers on mux.
func Mount(mux *http.ServeMux, d Deps) {
	svc := app.New(app.Repos{
		Users:    postgres.NewUsers(d.Pool),
		Sessions: postgres.NewSessions(d.Pool),
		APIKeys:  postgres.NewAPIKeys(d.Pool),
		Audit:    postgres.NewAudit(d.Pool),
	}, app.Config{
		JWTSecret:    d.JWTSecret,
		RefreshTTL:   d.SessionTTL,
		APIKeyPrefix: d.APIKeyPrefix,
	})
	h := rpc.NewHandlers(svc)
	opts := connect.WithInterceptors(d.Interceptors...)
	mux.Handle(identityv1connect.NewAuthServiceHandler(h, opts))
	mux.Handle(identityv1connect.NewUserServiceHandler(h, opts))
	mux.Handle(identityv1connect.NewApiKeyServiceHandler(h, opts))
	mux.Handle(identityv1connect.NewAuditServiceHandler(h, opts))
	mux.HandleFunc("POST /v1/usage", func(writer http.ResponseWriter, request *http.Request) {
		owner, err := httpx.Owner(request, authn.NewVerifier(d.JWTSecret))
		if err != nil {
			httpx.Error(writer, err)
			return
		}
		forms, err := usage.Read(request.Context(), d.Pool, d.Limits, owner, "forms")
		if err != nil {
			httpx.Error(writer, err)
			return
		}
		meetings, err := usage.Read(request.Context(), d.Pool, d.Limits, owner, "meetings")
		if err != nil {
			httpx.Error(writer, err)
			return
		}
		httpx.JSON(writer, 200, map[string]any{"forms": forms, "meetings": meetings, "plan": "free"})
	})
}
