// Package rpc adapts ConnectRPC requests to the identity application service.
// A single Handlers type satisfies every generated identity service interface
// (AuthService, UserService, ApiKeyService, AuditService) — method names don't
// collide across them.
package rpc

import (
	"context"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	commonv1 "github.com/aashishrajdev/halomail/services/shared/gen/halomail/common/v1"
	identityv1 "github.com/aashishrajdev/halomail/services/shared/gen/halomail/identity/v1"

	"github.com/aashishrajdev/halomail/services/identity/internal/app"
	"github.com/aashishrajdev/halomail/services/shared/connectutil"
	"github.com/aashishrajdev/halomail/services/shared/errs"
)

type Handlers struct {
	app *app.Service
}

func NewHandlers(a *app.Service) *Handlers { return &Handlers{app: a} }

// principal resolves the caller from the Authorization header (Bearer JWT or
// ApiKey secret) for authenticated RPCs.
func (h *Handlers) principal(ctx context.Context, req connect.AnyRequest) (userID, orgID string, err error) {
	authz := req.Header().Get("Authorization")
	switch {
	case strings.HasPrefix(authz, "Bearer "):
		return h.app.VerifyToken(ctx, strings.TrimSpace(authz[len("Bearer "):]))
	default:
		return "", "", errs.Unauthorized("missing or malformed Authorization header")
	}
}

// ---- AuthService ---------------------------------------------------------

func (h *Handlers) Register(ctx context.Context, req *connect.Request[identityv1.RegisterRequest]) (*connect.Response[identityv1.RegisterResponse], error) {
	res, err := h.app.Register(ctx, req.Msg.GetEmail(), req.Msg.GetPassword(), req.Msg.GetName())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.RegisterResponse{
		User:    toProtoUser(res.User),
		Session: toProtoSession(res.Session),
	}), nil
}

func (h *Handlers) Login(ctx context.Context, req *connect.Request[identityv1.LoginRequest]) (*connect.Response[identityv1.LoginResponse], error) {
	res, err := h.app.Login(ctx, req.Msg.GetEmail(), req.Msg.GetPassword())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.LoginResponse{
		User:    toProtoUser(res.User),
		Session: toProtoSession(res.Session),
	}), nil
}

func (h *Handlers) Logout(ctx context.Context, req *connect.Request[identityv1.LogoutRequest]) (*connect.Response[identityv1.LogoutResponse], error) {
	if err := h.app.Logout(ctx, req.Msg.GetRefreshToken()); err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.LogoutResponse{}), nil
}

func (h *Handlers) RefreshSession(ctx context.Context, req *connect.Request[identityv1.RefreshSessionRequest]) (*connect.Response[identityv1.RefreshSessionResponse], error) {
	sess, err := h.app.Refresh(ctx, req.Msg.GetRefreshToken())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.RefreshSessionResponse{Session: toProtoSession(sess)}), nil
}

func (h *Handlers) GetCurrentUser(ctx context.Context, req *connect.Request[identityv1.GetCurrentUserRequest]) (*connect.Response[identityv1.GetCurrentUserResponse], error) {
	userID, _, err := h.principal(ctx, req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	user, err := h.app.GetUser(ctx, userID)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.GetCurrentUserResponse{User: toProtoUser(user)}), nil
}

// VerifyToken never errors on an invalid token; it reports valid=false so the
// gateway can branch.
func (h *Handlers) VerifyToken(ctx context.Context, req *connect.Request[identityv1.VerifyTokenRequest]) (*connect.Response[identityv1.VerifyTokenResponse], error) {
	userID, orgID, err := h.app.VerifyToken(ctx, req.Msg.GetAccessToken())
	return connect.NewResponse(&identityv1.VerifyTokenResponse{
		Valid:  err == nil,
		UserId: userID,
		OrgId:  orgID,
	}), nil
}

// ---- UserService ---------------------------------------------------------

func (h *Handlers) GetUser(ctx context.Context, req *connect.Request[identityv1.GetUserRequest]) (*connect.Response[identityv1.GetUserResponse], error) {
	ownerID, _, authErr := h.principal(ctx, req)
	if authErr != nil {
		return nil, connectutil.ToConnect(authErr)
	}
	if ownerID != req.Msg.GetId() {
		return nil, connectutil.ToConnect(errs.Forbidden("not your profile"))
	}
	user, err := h.app.GetUser(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.GetUserResponse{User: toProtoUser(user)}), nil
}

func (h *Handlers) GetUserByHandle(ctx context.Context, req *connect.Request[identityv1.GetUserByHandleRequest]) (*connect.Response[identityv1.GetUserByHandleResponse], error) {
	user, err := h.app.GetUserByHandle(ctx, req.Msg.GetHandle())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.GetUserByHandleResponse{User: &identityv1.User{Id: user.ID, Name: user.Name, Handle: user.Handle, Timezone: user.Timezone}}), nil
}

func (h *Handlers) UpdateUser(ctx context.Context, req *connect.Request[identityv1.UpdateUserRequest]) (*connect.Response[identityv1.UpdateUserResponse], error) {
	userID, _, err := h.principal(ctx, req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	user, err := h.app.UpdateUser(ctx, userID,
		req.Msg.GetName(), req.Msg.GetHandle(), req.Msg.GetAvatarUrl(), req.Msg.GetTimezone())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.UpdateUserResponse{User: toProtoUser(user)}), nil
}

// ---- ApiKeyService -------------------------------------------------------

func (h *Handlers) CreateApiKey(ctx context.Context, req *connect.Request[identityv1.CreateApiKeyRequest]) (*connect.Response[identityv1.CreateApiKeyResponse], error) {
	userID, orgID, err := h.principal(ctx, req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	key, secret, err := h.app.CreateAPIKey(ctx, userID, orgID, req.Msg.GetName(), req.Msg.GetScopes())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.CreateApiKeyResponse{
		Key:    toProtoAPIKey(*key),
		Secret: secret,
	}), nil
}

func (h *Handlers) ListApiKeys(ctx context.Context, req *connect.Request[identityv1.ListApiKeysRequest]) (*connect.Response[identityv1.ListApiKeysResponse], error) {
	userID, _, err := h.principal(ctx, req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	keys, err := h.app.ListAPIKeys(ctx, userID)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	out := make([]*identityv1.ApiKey, 0, len(keys))
	for _, k := range keys {
		out = append(out, toProtoAPIKey(k))
	}
	return connect.NewResponse(&identityv1.ListApiKeysResponse{Keys: out}), nil
}

func (h *Handlers) RevokeApiKey(ctx context.Context, req *connect.Request[identityv1.RevokeApiKeyRequest]) (*connect.Response[identityv1.RevokeApiKeyResponse], error) {
	userID, orgID, err := h.principal(ctx, req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	if err := h.app.RevokeAPIKey(ctx, req.Msg.GetId(), userID, orgID); err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.RevokeApiKeyResponse{}), nil
}

func (h *Handlers) VerifyApiKey(ctx context.Context, req *connect.Request[identityv1.VerifyApiKeyRequest]) (*connect.Response[identityv1.VerifyApiKeyResponse], error) {
	userID, orgID, scopes, err := h.app.VerifyAPIKey(ctx, req.Msg.GetSecret())
	return connect.NewResponse(&identityv1.VerifyApiKeyResponse{
		Valid:  err == nil,
		UserId: userID,
		OrgId:  orgID,
		Scopes: scopes,
	}), nil
}

// ---- AuditService --------------------------------------------------------

func (h *Handlers) ListAuditLogs(ctx context.Context, req *connect.Request[identityv1.ListAuditLogsRequest]) (*connect.Response[identityv1.ListAuditLogsResponse], error) {
	_, orgID, err := h.principal(ctx, req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	limit := int(req.Msg.GetPage().GetPageSize())
	if limit <= 0 {
		limit = 50
	}
	offset := atoiSafe(req.Msg.GetPage().GetPageToken())

	logs, err := h.app.ListAudit(ctx, orgID, limit, offset)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	out := make([]*identityv1.AuditLog, 0, len(logs))
	for _, l := range logs {
		out = append(out, toProtoAudit(l))
	}
	next := ""
	if len(logs) == limit {
		next = strconv.Itoa(offset + limit)
	}
	return connect.NewResponse(&identityv1.ListAuditLogsResponse{
		Logs: out,
		Page: &commonv1.PageResponse{NextPageToken: next, TotalSize: -1},
	}), nil
}

func (h *Handlers) RecordAuditLog(ctx context.Context, req *connect.Request[identityv1.RecordAuditLogRequest]) (*connect.Response[identityv1.RecordAuditLogResponse], error) {
	ownerID, orgID, authErr := h.principal(ctx, req)
	if authErr != nil {
		return nil, connectutil.ToConnect(authErr)
	}
	id, err := h.app.RecordAudit(ctx, app.RecordAuditInput{
		OrgID:      orgID,
		ActorID:    ownerID,
		Action:     req.Msg.GetAction(),
		TargetType: req.Msg.GetTargetType(),
		TargetID:   req.Msg.GetTargetId(),
		Metadata:   req.Msg.GetMetadata(),
		IP:         req.Msg.GetIp(),
	})
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&identityv1.RecordAuditLogResponse{Id: id}), nil
}

func atoiSafe(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
