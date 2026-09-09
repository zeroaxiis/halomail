// Package rpc adapts ConnectRPC requests to the scheduling application service.
// One Handlers type satisfies EventTypeService, AvailabilityService,
// BookingService, and CalendarService.
package rpc

import (
	"context"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	commonv1 "github.com/aashishrajdev/halomail/services/shared/gen/halomail/common/v1"
	schedulingv1 "github.com/aashishrajdev/halomail/services/shared/gen/halomail/scheduling/v1"

	"github.com/aashishrajdev/halomail/services/scheduling/internal/app"
	"github.com/aashishrajdev/halomail/services/shared/authn"
	"github.com/aashishrajdev/halomail/services/shared/connectutil"
	"github.com/aashishrajdev/halomail/services/shared/errs"
)

type Handlers struct {
	app      *app.Service
	verifier authn.Verifier
}

func NewHandlers(a *app.Service, v authn.Verifier) *Handlers {
	return &Handlers{app: a, verifier: v}
}

// principal authenticates an owner via the Bearer access token.
func (h *Handlers) principal(req connect.AnyRequest) (userID, orgID string, err error) {
	authz := req.Header().Get("Authorization")
	if !strings.HasPrefix(authz, "Bearer ") {
		return "", "", errs.Unauthorized("a bearer token is required")
	}
	uid, oid, verr := h.verifier.Verify(strings.TrimSpace(authz[len("Bearer "):]))
	if verr != nil {
		return "", "", errs.Unauthorized("invalid or expired token")
	}
	return uid, oid, nil
}

// ---- EventTypeService ----------------------------------------------------

func (h *Handlers) CreateEventType(ctx context.Context, req *connect.Request[schedulingv1.CreateEventTypeRequest]) (*connect.Response[schedulingv1.CreateEventTypeResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	et, err := h.app.CreateEventType(ctx, ownerID, app.EventTypeInput{
		Title:               req.Msg.GetTitle(),
		Slug:                req.Msg.GetSlug(),
		Description:         req.Msg.GetDescription(),
		DurationMinutes:     int(req.Msg.GetDurationMinutes()),
		LocationKind:        locKindFromProto(req.Msg.GetLocationKind()),
		LocationDetail:      req.Msg.GetLocationDetail(),
		BufferBeforeMinutes: int(req.Msg.GetBufferBeforeMinutes()),
		BufferAfterMinutes:  int(req.Msg.GetBufferAfterMinutes()),
		Color:               req.Msg.GetColor(),
	})
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.CreateEventTypeResponse{EventType: toProtoEventType(et)}), nil
}

func (h *Handlers) GetEventType(ctx context.Context, req *connect.Request[schedulingv1.GetEventTypeRequest]) (*connect.Response[schedulingv1.GetEventTypeResponse], error) {
	et, err := h.app.GetEventType(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.GetEventTypeResponse{EventType: toProtoEventType(et)}), nil
}

func (h *Handlers) ListEventTypes(ctx context.Context, req *connect.Request[schedulingv1.ListEventTypesRequest]) (*connect.Response[schedulingv1.ListEventTypesResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	ets, err := h.app.ListEventTypes(ctx, ownerID)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	out := make([]*schedulingv1.EventType, 0, len(ets))
	for i := range ets {
		out = append(out, toProtoEventType(&ets[i]))
	}
	return connect.NewResponse(&schedulingv1.ListEventTypesResponse{EventTypes: out}), nil
}

func (h *Handlers) UpdateEventType(ctx context.Context, req *connect.Request[schedulingv1.UpdateEventTypeRequest]) (*connect.Response[schedulingv1.UpdateEventTypeResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	et, err := h.app.UpdateEventType(ctx, ownerID, req.Msg.GetId(), app.EventTypeInput{
		Title:               req.Msg.GetTitle(),
		Description:         req.Msg.GetDescription(),
		DurationMinutes:     int(req.Msg.GetDurationMinutes()),
		LocationKind:        locKindFromProto(req.Msg.GetLocationKind()),
		LocationDetail:      req.Msg.GetLocationDetail(),
		BufferBeforeMinutes: int(req.Msg.GetBufferBeforeMinutes()),
		BufferAfterMinutes:  int(req.Msg.GetBufferAfterMinutes()),
		Color:               req.Msg.GetColor(),
	}, req.Msg.GetActive())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.UpdateEventTypeResponse{EventType: toProtoEventType(et)}), nil
}

func (h *Handlers) DeleteEventType(ctx context.Context, req *connect.Request[schedulingv1.DeleteEventTypeRequest]) (*connect.Response[schedulingv1.DeleteEventTypeResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	if err := h.app.DeleteEventType(ctx, ownerID, req.Msg.GetId()); err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.DeleteEventTypeResponse{}), nil
}

// ---- AvailabilityService -------------------------------------------------

func (h *Handlers) GetAvailability(ctx context.Context, req *connect.Request[schedulingv1.GetAvailabilityRequest]) (*connect.Response[schedulingv1.GetAvailabilityResponse], error) {
	ownerID := req.Msg.GetOwnerId()
	if ownerID == "" {
		var err error
		if ownerID, _, err = h.principal(req); err != nil {
			return nil, connectutil.ToConnect(err)
		}
	}
	a, err := h.app.GetAvailability(ctx, ownerID)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.GetAvailabilityResponse{Availability: toProtoAvailability(a)}), nil
}

func (h *Handlers) SetAvailability(ctx context.Context, req *connect.Request[schedulingv1.SetAvailabilityRequest]) (*connect.Response[schedulingv1.SetAvailabilityResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	a, err := h.app.SetAvailability(ctx, ownerID, req.Msg.GetTimezone(),
		rulesFromProto(req.Msg.GetRules()), overridesFromProto(req.Msg.GetOverrides()))
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.SetAvailabilityResponse{Availability: toProtoAvailability(a)}), nil
}

// ---- BookingService ------------------------------------------------------

func (h *Handlers) ListSlots(ctx context.Context, req *connect.Request[schedulingv1.ListSlotsRequest]) (*connect.Response[schedulingv1.ListSlotsResponse], error) {
	slotList, err := h.app.ListSlots(ctx, req.Msg.GetEventTypeId(), req.Msg.GetFromDate(), req.Msg.GetToDate(), req.Msg.GetInviteeTimezone())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	out := make([]*schedulingv1.Slot, len(slotList))
	for i, s := range slotList {
		out[i] = toProtoSlot(s)
	}
	return connect.NewResponse(&schedulingv1.ListSlotsResponse{Slots: out}), nil
}

func (h *Handlers) CreateBooking(ctx context.Context, req *connect.Request[schedulingv1.CreateBookingRequest]) (*connect.Response[schedulingv1.CreateBookingResponse], error) {
	b, err := h.app.CreateBooking(ctx,
		req.Msg.GetEventTypeId(),
		req.Msg.GetInviteeName(),
		req.Msg.GetInviteeEmail(),
		req.Msg.GetInviteeTimezone(),
		req.Msg.GetStart().AsTime(),
		req.Msg.GetNotes(),
	)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.CreateBookingResponse{Booking: toProtoBooking(b)}), nil
}

func (h *Handlers) GetBooking(ctx context.Context, req *connect.Request[schedulingv1.GetBookingRequest]) (*connect.Response[schedulingv1.GetBookingResponse], error) {
	b, err := h.app.GetBooking(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.GetBookingResponse{Booking: toProtoBooking(b)}), nil
}

func (h *Handlers) ListBookings(ctx context.Context, req *connect.Request[schedulingv1.ListBookingsRequest]) (*connect.Response[schedulingv1.ListBookingsResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	limit := int(req.Msg.GetPage().GetPageSize())
	if limit <= 0 {
		limit = 50
	}
	offset := atoiSafe(req.Msg.GetPage().GetPageToken())

	bookings, err := h.app.ListBookings(ctx, ownerID, statusFromProto(req.Msg.GetStatus()), limit, offset)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	out := make([]*schedulingv1.Booking, 0, len(bookings))
	for i := range bookings {
		out = append(out, toProtoBooking(&bookings[i]))
	}
	next := ""
	if len(bookings) == limit {
		next = strconv.Itoa(offset + limit)
	}
	return connect.NewResponse(&schedulingv1.ListBookingsResponse{
		Bookings: out,
		Page:     &commonv1.PageResponse{NextPageToken: next, TotalSize: -1},
	}), nil
}

func (h *Handlers) RescheduleBooking(ctx context.Context, req *connect.Request[schedulingv1.RescheduleBookingRequest]) (*connect.Response[schedulingv1.RescheduleBookingResponse], error) {
	b, err := h.app.RescheduleBooking(ctx, req.Msg.GetRescheduleToken(), req.Msg.GetNewStart().AsTime())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.RescheduleBookingResponse{Booking: toProtoBooking(b)}), nil
}

func (h *Handlers) CancelBooking(ctx context.Context, req *connect.Request[schedulingv1.CancelBookingRequest]) (*connect.Response[schedulingv1.CancelBookingResponse], error) {
	b, err := h.app.CancelBooking(ctx, req.Msg.GetCancelToken(), req.Msg.GetReason())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.CancelBookingResponse{Booking: toProtoBooking(b)}), nil
}

func (h *Handlers) GetUsageStats(ctx context.Context, req *connect.Request[schedulingv1.GetUsageStatsRequest]) (*connect.Response[schedulingv1.GetUsageStatsResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	stats, err := h.app.GetUsageStats(ctx, ownerID)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.GetUsageStatsResponse{
		TotalBookings:     int32(stats.TotalBookings),
		UpcomingBookings:  int32(stats.UpcomingBookings),
		CancelledBookings: int32(stats.CancelledBookings),
	}), nil
}

// ---- CalendarService -----------------------------------------------------

func (h *Handlers) StartConnect(ctx context.Context, req *connect.Request[schedulingv1.StartConnectRequest]) (*connect.Response[schedulingv1.StartConnectResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	url, err := h.app.StartConnect(ctx, ownerID, providerFromProto(req.Msg.GetProvider()))
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.StartConnectResponse{AuthorizationUrl: url}), nil
}

func (h *Handlers) ListConnections(ctx context.Context, req *connect.Request[schedulingv1.ListConnectionsRequest]) (*connect.Response[schedulingv1.ListConnectionsResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	conns, err := h.app.ListConnections(ctx, ownerID)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	out := make([]*schedulingv1.CalendarConnection, 0, len(conns))
	for _, c := range conns {
		out = append(out, toProtoConnection(c))
	}
	return connect.NewResponse(&schedulingv1.ListConnectionsResponse{Connections: out}), nil
}

func (h *Handlers) DisconnectCalendar(ctx context.Context, req *connect.Request[schedulingv1.DisconnectCalendarRequest]) (*connect.Response[schedulingv1.DisconnectCalendarResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	if err := h.app.DisconnectCalendar(ctx, ownerID, req.Msg.GetId()); err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&schedulingv1.DisconnectCalendarResponse{}), nil
}

func atoiSafe(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
