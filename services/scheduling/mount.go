// Package scheduling exposes Mount so the monolith can host it in-process.
package scheduling

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zeroaxiis/halomail/services/scheduling/internal/calendar"
	"github.com/zeroaxiis/halomail/services/shared/accesskey"
	"github.com/zeroaxiis/halomail/services/shared/authn"
	"github.com/zeroaxiis/halomail/services/shared/config"
	"github.com/zeroaxiis/halomail/services/shared/errs"
	schedulingv1connect "github.com/zeroaxiis/halomail/services/shared/gen/halomail/scheduling/v1/schedulingv1connect"
	"github.com/zeroaxiis/halomail/services/shared/httpx"
	"github.com/zeroaxiis/halomail/services/shared/usage"

	"github.com/zeroaxiis/halomail/services/scheduling/internal/adapters/postgres"
	"github.com/zeroaxiis/halomail/services/scheduling/internal/adapters/rpc"
	"github.com/zeroaxiis/halomail/services/scheduling/internal/app"
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
		key := request.Header.Get("X-HaloMail-Key")
		var ownerID, ownerName, ownerHandle, avatarURL, tier string
		if key != "" {
			owner, err := accesskey.Verify(request.Context(), d.Pool, key, "meetings")
			if err != nil {
				httpx.Error(writer, err)
				return
			}
			ownerID = owner.ID
			ownerName = owner.Name
			ownerHandle = owner.Handle
			_ = d.Pool.QueryRow(request.Context(), `SELECT u.avatar_url, COALESCE(o.tier, 'free') FROM users u JOIN orgs o ON o.id = u.org_id WHERE u.id = $1`, ownerID).Scan(&avatarURL, &tier)
		} else {
			var body struct {
				Handle string `json:"handle"`
			}
			_ = json.NewDecoder(request.Body).Decode(&body)
			handle := strings.TrimSpace(body.Handle)
			if handle == "" {
				handle = strings.TrimSpace(request.URL.Query().Get("handle"))
			}
			if handle == "" {
				httpx.Error(writer, errs.Invalid("handle or meeting access key required"))
				return
			}
			err := d.Pool.QueryRow(request.Context(), `SELECT u.id, u.name, u.handle, u.avatar_url, COALESCE(o.tier, 'free') FROM users u JOIN orgs o ON o.id = u.org_id WHERE LOWER(u.handle) = LOWER($1)`, handle).Scan(&ownerID, &ownerName, &ownerHandle, &avatarURL, &tier)
			if err != nil {
				httpx.Error(writer, errs.NotFound("host not found"))
				return
			}
		}

		events, err := svc.ListEventTypes(request.Context(), ownerID)
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
		watermark := (tier == "free" || tier == "")
		httpx.JSON(writer, 200, map[string]any{
			"owner": map[string]string{
				"id":     ownerID,
				"name":   ownerName,
				"handle": ownerHandle,
				"avatar": avatarURL,
				"tier":   tier,
			},
			"watermark": watermark,
			"events":    public,
		})
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
