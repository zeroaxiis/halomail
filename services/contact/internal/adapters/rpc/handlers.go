// Package rpc adapts ConnectRPC requests to the contact application service.
// One Handlers type satisfies FormService and MessageService.
package rpc

import (
	"context"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	commonv1 "github.com/aashishrajdev/halomail/services/shared/gen/halomail/common/v1"
	contactv1 "github.com/aashishrajdev/halomail/services/shared/gen/halomail/contact/v1"
	identityv1 "github.com/aashishrajdev/halomail/services/shared/gen/halomail/identity/v1"
	identityv1connect "github.com/aashishrajdev/halomail/services/shared/gen/halomail/identity/v1/identityv1connect"

	"github.com/aashishrajdev/halomail/services/contact/internal/app"
	"github.com/aashishrajdev/halomail/services/shared/authn"
	"github.com/aashishrajdev/halomail/services/shared/connectutil"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/aashishrajdev/halomail/services/shared/httpx"
)

type Handlers struct {
	app      *app.Service
	verifier authn.Verifier
	apiKeys  identityv1connect.ApiKeyServiceClient
}

func NewHandlers(a *app.Service, v authn.Verifier, apiKeys identityv1connect.ApiKeyServiceClient) *Handlers {
	return &Handlers{app: a, verifier: v, apiKeys: apiKeys}
}

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

// ---- FormService ---------------------------------------------------------

func (h *Handlers) CreateForm(ctx context.Context, req *connect.Request[contactv1.CreateFormRequest]) (*connect.Response[contactv1.CreateFormResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	f, err := h.app.CreateForm(ctx, ownerID, app.FormInput{
		Name:           req.Msg.GetName(),
		Slug:           req.Msg.GetSlug(),
		TargetEmail:    req.Msg.GetTargetEmail(),
		SpamProtection: spamProtFromProto(req.Msg.GetSpamProtection()),
		RedirectURL:    req.Msg.GetRedirectUrl(),
		Fields:         fieldsFromProto(req.Msg.GetFields()),
	})
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&contactv1.CreateFormResponse{Form: toProtoForm(f)}), nil
}

func (h *Handlers) GetForm(ctx context.Context, req *connect.Request[contactv1.GetFormRequest]) (*connect.Response[contactv1.GetFormResponse], error) {
	ownerID, _, authErr := h.principal(req)
	if authErr != nil {
		return nil, connectutil.ToConnect(authErr)
	}
	f, err := h.app.GetForm(ctx, req.Msg.GetId(), req.Msg.GetSlug())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	if f.OwnerID != ownerID {
		return nil, connectutil.ToConnect(errs.NotFound("form not found"))
	}
	return connect.NewResponse(&contactv1.GetFormResponse{Form: toProtoForm(f)}), nil
}

func (h *Handlers) ListForms(ctx context.Context, req *connect.Request[contactv1.ListFormsRequest]) (*connect.Response[contactv1.ListFormsResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	forms, err := h.app.ListForms(ctx, ownerID)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	out := make([]*contactv1.Form, 0, len(forms))
	for i := range forms {
		out = append(out, toProtoForm(&forms[i]))
	}
	return connect.NewResponse(&contactv1.ListFormsResponse{Forms: out}), nil
}

func (h *Handlers) UpdateForm(ctx context.Context, req *connect.Request[contactv1.UpdateFormRequest]) (*connect.Response[contactv1.UpdateFormResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	f, err := h.app.UpdateForm(ctx, ownerID, req.Msg.GetId(), app.FormInput{
		Name:           req.Msg.GetName(),
		TargetEmail:    req.Msg.GetTargetEmail(),
		SpamProtection: spamProtFromProto(req.Msg.GetSpamProtection()),
		RedirectURL:    req.Msg.GetRedirectUrl(),
		Fields:         fieldsFromProto(req.Msg.GetFields()),
	}, req.Msg.GetActive())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&contactv1.UpdateFormResponse{Form: toProtoForm(f)}), nil
}

func (h *Handlers) DeleteForm(ctx context.Context, req *connect.Request[contactv1.DeleteFormRequest]) (*connect.Response[contactv1.DeleteFormResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	if err := h.app.DeleteForm(ctx, ownerID, req.Msg.GetId()); err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&contactv1.DeleteFormResponse{}), nil
}

// ---- MessageService ------------------------------------------------------

func (h *Handlers) SubmitMessage(ctx context.Context, req *connect.Request[contactv1.SubmitMessageRequest]) (*connect.Response[contactv1.SubmitMessageResponse], error) {
	verifyRes, err := h.apiKeys.VerifyApiKey(ctx, connect.NewRequest(&identityv1.VerifyApiKeyRequest{
		Secret: req.Msg.GetAccessKey(),
	}))
	if err != nil {
		return nil, connectutil.ToConnect(errs.Unauthorized("invalid access key"))
	}
	if !verifyRes.Msg.GetValid() || len(verifyRes.Msg.GetScopes()) != 1 || verifyRes.Msg.GetScopes()[0] != "forms:submit" {
		return nil, connectutil.ToConnect(errs.Unauthorized("invalid access key"))
	}
	ownerID := verifyRes.Msg.GetUserId()

	res, err := h.app.SubmitMessage(ctx, ownerID, app.SubmitInput{
		SenderName:  req.Msg.GetSenderName(),
		SenderEmail: req.Msg.GetSenderEmail(),
		Data:        req.Msg.GetData(),
		Honeypot:    req.Msg.GetHoneypot(),
		IP:          clientIP(req),
		UserAgent:   req.Header().Get("User-Agent"),
	})
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&contactv1.SubmitMessageResponse{
		Id:          res.Message.ID,
		Accepted:    true,
		RedirectUrl: res.RedirectURL,
	}), nil
}

func (h *Handlers) ListMessages(ctx context.Context, req *connect.Request[contactv1.ListMessagesRequest]) (*connect.Response[contactv1.ListMessagesResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	limit := int(req.Msg.GetPage().GetPageSize())
	if limit <= 0 {
		limit = 50
	}
	offset := atoiSafe(req.Msg.GetPage().GetPageToken())

	msgs, err := h.app.ListMessages(ctx, ownerID, req.Msg.GetFormId(), req.Msg.GetUnreadOnly(), limit, offset)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	out := make([]*contactv1.Message, 0, len(msgs))
	for i := range msgs {
		out = append(out, toProtoMessage(&msgs[i]))
	}
	next := ""
	if len(msgs) == limit {
		next = strconv.Itoa(offset + limit)
	}
	return connect.NewResponse(&contactv1.ListMessagesResponse{
		Messages: out,
		Page:     &commonv1.PageResponse{NextPageToken: next, TotalSize: -1},
	}), nil
}

func (h *Handlers) GetMessage(ctx context.Context, req *connect.Request[contactv1.GetMessageRequest]) (*connect.Response[contactv1.GetMessageResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	m, err := h.app.GetMessage(ctx, ownerID, req.Msg.GetId())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&contactv1.GetMessageResponse{Message: toProtoMessage(m)}), nil
}

func (h *Handlers) MarkRead(ctx context.Context, req *connect.Request[contactv1.MarkReadRequest]) (*connect.Response[contactv1.MarkReadResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	if err := h.app.MarkRead(ctx, ownerID, req.Msg.GetId(), req.Msg.GetRead()); err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&contactv1.MarkReadResponse{}), nil
}

func (h *Handlers) DeleteMessage(ctx context.Context, req *connect.Request[contactv1.DeleteMessageRequest]) (*connect.Response[contactv1.DeleteMessageResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	if err := h.app.DeleteMessage(ctx, ownerID, req.Msg.GetId()); err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&contactv1.DeleteMessageResponse{}), nil
}

func (h *Handlers) GetUsageStats(ctx context.Context, req *connect.Request[contactv1.GetUsageStatsRequest]) (*connect.Response[contactv1.GetUsageStatsResponse], error) {
	ownerID, _, err := h.principal(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	stats, err := h.app.GetUsageStats(ctx, ownerID)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&contactv1.GetUsageStatsResponse{
		TotalMessages:  int32(stats.TotalMessages),
		UnreadMessages: int32(stats.UnreadMessages),
		SpamPrevented:  int32(stats.SpamPrevented),
	}), nil
}

// ---- helpers -------------------------------------------------------------

func clientIP(req connect.AnyRequest) string {
	return httpx.Peer(req.Peer().Addr)
}

func atoiSafe(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
