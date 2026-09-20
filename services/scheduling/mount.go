// Package scheduling exposes Mount so the monolith can host it in-process.
package scheduling

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aashishrajdev/halomail/services/scheduling/internal/calendar"
	"github.com/aashishrajdev/halomail/services/shared/accesskey"
	"github.com/aashishrajdev/halomail/services/shared/authn"
	"github.com/aashishrajdev/halomail/services/shared/config"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	schedulingv1connect "github.com/aashishrajdev/halomail/services/shared/gen/halomail/scheduling/v1/schedulingv1connect"
	"github.com/aashishrajdev/halomail/services/shared/httpx"
	"github.com/aashishrajdev/halomail/services/shared/usage"

	"github.com/aashishrajdev/halomail/services/scheduling/internal/adapters/postgres"
	"github.com/aashishrajdev/halomail/services/scheduling/internal/adapters/rpc"
	"github.com/aashishrajdev/halomail/services/scheduling/internal/app"
)

type Deps struct {
	Pool          *pgxpool.Pool
	JWTSecret     string
	Google        config.OAuth
	Microsoft     config.Microsoft
	Interceptors  []connect.Interceptor
	Limits        usage.Policy
	EncryptionKey string
	WebURL        string
	Context       context.Context
	Logger        *slog.Logger
}

func Mount(mux *http.ServeMux, d Deps) {
	calendarClient := calendar.New(d.Pool, d.Google, d.EncryptionKey, d.WebURL)
	svc := app.New(app.Repos{
		EventTypes:   postgres.NewEventTypes(d.Pool),
		Availability: postgres.NewAvailability(d.Pool),
		Bookings:     postgres.NewBookings(d.Pool, d.Limits),
		Calendars:    postgres.NewCalendars(d.Pool),
	}, app.Config{Google: d.Google, Microsoft: d.Microsoft, Calendar: calendarClient})
	h := rpc.NewHandlers(svc, authn.NewVerifier(d.JWTSecret), d.Pool)
	if d.Context != nil && d.Logger != nil {
		go calendarClient.Run(d.Context, d.Logger)
	}
	opts := connect.WithInterceptors(d.Interceptors...)
	mux.Handle(schedulingv1connect.NewEventTypeServiceHandler(h, opts))
	mux.Handle(schedulingv1connect.NewAvailabilityServiceHandler(h, opts))
	mux.Handle(schedulingv1connect.NewBookingServiceHandler(h, opts))
	mux.Handle(schedulingv1connect.NewCalendarServiceHandler(h, opts))
	mux.HandleFunc("POST /v1/meetings/info", func(writer http.ResponseWriter, request *http.Request) {
		owner, err := accesskey.Verify(request.Context(), d.Pool, request.Header.Get("X-HaloMail-Key"), "meetings")
		if err != nil {
			httpx.Error(writer, err)
			return
		}
		events, err := svc.ListEventTypes(request.Context(), owner.ID)
		if err != nil {
			httpx.Error(writer, err)
			return
		}
		public := []map[string]any{}
		for _, event := range events {
			if event.Active {
				public = append(public, map[string]any{"id": event.ID, "title": event.Title, "description": event.Description, "durationMinutes": event.DurationMinutes})
			}
		}
		httpx.JSON(writer, 200, map[string]any{"owner": map[string]string{"name": owner.Name, "handle": owner.Handle}, "events": public})
	})
	mux.HandleFunc("POST /v1/calendar/callback", func(writer http.ResponseWriter, request *http.Request) {
		owner, err := httpx.Owner(request, authn.NewVerifier(d.JWTSecret))
		if err != nil {
			httpx.Error(writer, err)
			return
		}
		var input struct {
			State string
			Code  string
		}
		if err = json.NewDecoder(request.Body).Decode(&input); err != nil {
			httpx.Error(writer, errs.Invalid("invalid callback"))
			return
		}
		if err = calendarClient.Callback(request.Context(), owner, input.State, input.Code); err != nil {
			httpx.Error(writer, err)
			return
		}
		httpx.JSON(writer, 200, map[string]bool{"connected": true})
	})
}
