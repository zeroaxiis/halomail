// Package app holds template use cases. ListThemes and RenderPreview are pure
// (no storage); the rest are owner-scoped CRUD over a TemplateRepo.
package app

import (
	"context"
	"strings"
	"time"

	"github.com/zeroaxiis/halomail/services/template/internal/domain"
	"github.com/zeroaxiis/halomail/services/template/internal/themes"
	"github.com/zeroaxiis/halomail/services/shared/errs"
	"github.com/zeroaxiis/halomail/services/shared/idgen"
)

type TemplateRepo interface {
	Create(ctx context.Context, t *domain.Template) error
	GetByID(ctx context.Context, id string) (*domain.Template, error)
	ListByOwner(ctx context.Context, ownerID string) ([]domain.Template, error)
	Update(ctx context.Context, t *domain.Template) error
	Delete(ctx context.Context, id, ownerID string) error
}

type Repos struct{ Templates TemplateRepo }

type Service struct {
	templates TemplateRepo
	now       func() time.Time
}

func New(r Repos) *Service { return &Service{templates: r.Templates, now: time.Now} }

// ListThemes returns the built-in theme gallery with previews.
func (s *Service) ListThemes() []domain.ThemeInfo { return themes.Themes() }

// RenderPreview renders a theme (or custom HTML) with variables. Pure.
func (s *Service) RenderPreview(theme, subject, customHTML string, vars map[string]string) (string, string) {
	if theme == domain.ThemeCustom || (theme == "" && strings.TrimSpace(customHTML) != "") {
		return themes.RenderCustom(customHTML, subject, vars)
	}
	if theme == "" {
		theme = domain.ThemeMinimal
	}
	return themes.Render(theme, subject, vars)
}

func (s *Service) CreateTemplate(ctx context.Context, ownerID, name, theme, subject, customHTML string) (*domain.Template, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errs.Invalid("template name is required")
	}
	if theme == "" {
		theme = domain.ThemeMinimal
	}
	now := s.now()
	t := &domain.Template{
		ID:         idgen.Prefixed("tpl_"),
		OwnerID:    ownerID,
		Name:       name,
		Theme:      theme,
		Subject:    subject,
		CustomHTML: customHTML,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.templates.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) GetTemplate(ctx context.Context, ownerID, id string) (*domain.Template, error) {
	t, err := s.templates.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.OwnerID != ownerID {
		return nil, errs.NotFound("template not found")
	}
	return t, nil
}

func (s *Service) ListTemplates(ctx context.Context, ownerID string) ([]domain.Template, error) {
	return s.templates.ListByOwner(ctx, ownerID)
}

func (s *Service) UpdateTemplate(ctx context.Context, ownerID, id, name, theme, subject, customHTML string) (*domain.Template, error) {
	t, err := s.templates.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.OwnerID != ownerID {
		return nil, errs.Forbidden("not your template")
	}
	if name != "" {
		t.Name = name
	}
	if theme != "" {
		t.Theme = theme
	}
	t.Subject = subject
	t.CustomHTML = customHTML
	t.UpdatedAt = s.now()
	if err := s.templates.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) DeleteTemplate(ctx context.Context, ownerID, id string) error {
	return s.templates.Delete(ctx, id, ownerID)
}
