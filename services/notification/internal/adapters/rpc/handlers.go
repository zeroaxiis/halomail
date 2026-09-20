// Package rpc adapts ConnectRPC requests to the notification application
// service. One Handlers type satisfies EmailService and WebhookService.
//
// SendEmail and Dispatch are internal (called by other services); the gateway
// does not expose them publicly. Webhook management RPCs require a bearer token.
package rpc

import (
	"context"
	"strings"

	"connectrpc.com/connect"

	notificationv1 "github.com/aashishrajdev/halomail/services/shared/gen/halomail/notification/v1"

	"github.com/aashishrajdev/halomail/services/notification/internal/app"
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

func (h *Handlers) ownerID(req connect.AnyRequest) (string, error) {
	authz := req.Header().Get("Authorization")
	if !strings.HasPrefix(authz, "Bearer ") {
		return "", errs.Unauthorized("a bearer token is required")
	}
	uid, _, err := h.verifier.Verify(strings.TrimSpace(authz[len("Bearer "):]))
	if err != nil {
		return "", errs.Unauthorized("invalid or expired token")
	}
	return uid, nil
}

// ---- EmailService (internal) ---------------------------------------------

func (h *Handlers) SendEmail(ctx context.Context, req *connect.Request[notificationv1.SendEmailRequest]) (*connect.Response[notificationv1.SendEmailResponse], error) {
	return nil, connectutil.ToConnect(errs.Forbidden("direct email delivery is unavailable"))
}

// ---- WebhookService ------------------------------------------------------

func (h *Handlers) CreateWebhook(ctx context.Context, req *connect.Request[notificationv1.CreateWebhookRequest]) (*connect.Response[notificationv1.CreateWebhookResponse], error) {
	owner, err := h.ownerID(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	w, secret, err := h.app.CreateWebhook(ctx, owner, req.Msg.GetUrl(), eventsFromProto(req.Msg.GetEvents()))
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&notificationv1.CreateWebhookResponse{Webhook: toProtoWebhook(w), Secret: secret}), nil
}

func (h *Handlers) ListWebhooks(ctx context.Context, req *connect.Request[notificationv1.ListWebhooksRequest]) (*connect.Response[notificationv1.ListWebhooksResponse], error) {
	owner, err := h.ownerID(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	ws, err := h.app.ListWebhooks(ctx, owner)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	out := make([]*notificationv1.Webhook, 0, len(ws))
	for i := range ws {
		out = append(out, toProtoWebhook(&ws[i]))
	}
	return connect.NewResponse(&notificationv1.ListWebhooksResponse{Webhooks: out}), nil
}

func (h *Handlers) DeleteWebhook(ctx context.Context, req *connect.Request[notificationv1.DeleteWebhookRequest]) (*connect.Response[notificationv1.DeleteWebhookResponse], error) {
	owner, err := h.ownerID(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	if err := h.app.DeleteWebhook(ctx, owner, req.Msg.GetId()); err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&notificationv1.DeleteWebhookResponse{}), nil
}

func (h *Handlers) RotateSecret(ctx context.Context, req *connect.Request[notificationv1.RotateSecretRequest]) (*connect.Response[notificationv1.RotateSecretResponse], error) {
	owner, err := h.ownerID(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	secret, err := h.app.RotateSecret(ctx, owner, req.Msg.GetId())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&notificationv1.RotateSecretResponse{Secret: secret}), nil
}

// Dispatch is internal: other services publish events to fan out.
func (h *Handlers) Dispatch(ctx context.Context, req *connect.Request[notificationv1.DispatchRequest]) (*connect.Response[notificationv1.DispatchResponse], error) {
	return nil, connectutil.ToConnect(errs.Forbidden("direct event dispatch is unavailable"))
}
