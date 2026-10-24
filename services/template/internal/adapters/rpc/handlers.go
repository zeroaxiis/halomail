// Package rpc adapts ConnectRPC requests to the template application service.
package rpc

import (
	"context"
	"strings"

	"connectrpc.com/connect"

	templatev1 "github.com/zeroaxiis/halomail/services/shared/gen/halomail/template/v1"

	"github.com/zeroaxiis/halomail/services/template/internal/app"
	"github.com/zeroaxiis/halomail/services/shared/authn"
	"github.com/zeroaxiis/halomail/services/shared/connectutil"
	"github.com/zeroaxiis/halomail/services/shared/errs"
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

// ListThemes is public — the built-in gallery with previews.
func (h *Handlers) ListThemes(_ context.Context, _ *connect.Request[templatev1.ListThemesRequest]) (*connect.Response[templatev1.ListThemesResponse], error) {
	infos := h.app.ListThemes()
	out := make([]*templatev1.Theme, 0, len(infos))
	for _, ti := range infos {
		out = append(out, toProtoTheme(ti))
	}
	return connect.NewResponse(&templatev1.ListThemesResponse{Themes: out}), nil
}

// RenderPreview is public — render a theme or custom HTML with variables.
func (h *Handlers) RenderPreview(_ context.Context, req *connect.Request[templatev1.RenderPreviewRequest]) (*connect.Response[templatev1.RenderPreviewResponse], error) {
	subject, htmlOut := h.app.RenderPreview(
		themeFromProto(req.Msg.GetTheme()),
		req.Msg.GetSubject(),
		req.Msg.GetCustomHtml(),
		req.Msg.GetVariables(),
	)
	return connect.NewResponse(&templatev1.RenderPreviewResponse{Subject: subject, Html: htmlOut}), nil
}

func (h *Handlers) ListTemplates(ctx context.Context, req *connect.Request[templatev1.ListTemplatesRequest]) (*connect.Response[templatev1.ListTemplatesResponse], error) {
	owner, err := h.ownerID(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	tpls, err := h.app.ListTemplates(ctx, owner)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	out := make([]*templatev1.Template, 0, len(tpls))
	for i := range tpls {
		out = append(out, toProtoTemplate(&tpls[i]))
	}
	return connect.NewResponse(&templatev1.ListTemplatesResponse{Templates: out}), nil
}

func (h *Handlers) GetTemplate(ctx context.Context, req *connect.Request[templatev1.GetTemplateRequest]) (*connect.Response[templatev1.GetTemplateResponse], error) {
	owner, err := h.ownerID(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	t, err := h.app.GetTemplate(ctx, owner, req.Msg.GetId())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&templatev1.GetTemplateResponse{Template: toProtoTemplate(t)}), nil
}

func (h *Handlers) CreateTemplate(ctx context.Context, req *connect.Request[templatev1.CreateTemplateRequest]) (*connect.Response[templatev1.CreateTemplateResponse], error) {
	owner, err := h.ownerID(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	t, err := h.app.CreateTemplate(ctx, owner,
		req.Msg.GetName(), themeFromProto(req.Msg.GetTheme()), req.Msg.GetSubject(), req.Msg.GetCustomHtml())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&templatev1.CreateTemplateResponse{Template: toProtoTemplate(t)}), nil
}

func (h *Handlers) UpdateTemplate(ctx context.Context, req *connect.Request[templatev1.UpdateTemplateRequest]) (*connect.Response[templatev1.UpdateTemplateResponse], error) {
	owner, err := h.ownerID(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	t, err := h.app.UpdateTemplate(ctx, owner, req.Msg.GetId(),
		req.Msg.GetName(), themeFromProto(req.Msg.GetTheme()), req.Msg.GetSubject(), req.Msg.GetCustomHtml())
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&templatev1.UpdateTemplateResponse{Template: toProtoTemplate(t)}), nil
}

func (h *Handlers) DeleteTemplate(ctx context.Context, req *connect.Request[templatev1.DeleteTemplateRequest]) (*connect.Response[templatev1.DeleteTemplateResponse], error) {
	owner, err := h.ownerID(req)
	if err != nil {
		return nil, connectutil.ToConnect(err)
	}
	if err := h.app.DeleteTemplate(ctx, owner, req.Msg.GetId()); err != nil {
		return nil, connectutil.ToConnect(err)
	}
	return connect.NewResponse(&templatev1.DeleteTemplateResponse{}), nil
}
