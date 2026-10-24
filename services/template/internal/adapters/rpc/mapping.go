package rpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	templatev1 "github.com/zeroaxiis/halomail/services/shared/gen/halomail/template/v1"

	"github.com/zeroaxiis/halomail/services/template/internal/domain"
)

func ts(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func themeToProto(s string) templatev1.ThemeKind {
	switch s {
	case domain.ThemeMinimal:
		return templatev1.ThemeKind_THEME_KIND_MINIMAL
	case domain.ThemeApple:
		return templatev1.ThemeKind_THEME_KIND_APPLE
	case domain.ThemeNotion:
		return templatev1.ThemeKind_THEME_KIND_NOTION
	case domain.ThemeGlass:
		return templatev1.ThemeKind_THEME_KIND_GLASS
	case domain.ThemeTerminal:
		return templatev1.ThemeKind_THEME_KIND_TERMINAL
	case domain.ThemeCustom:
		return templatev1.ThemeKind_THEME_KIND_CUSTOM
	default:
		return templatev1.ThemeKind_THEME_KIND_UNSPECIFIED
	}
}

// themeFromProto maps the enum to a domain string. UNSPECIFIED maps to "" so
// the app can apply its own default (or detect custom HTML).
func themeFromProto(k templatev1.ThemeKind) string {
	switch k {
	case templatev1.ThemeKind_THEME_KIND_MINIMAL:
		return domain.ThemeMinimal
	case templatev1.ThemeKind_THEME_KIND_APPLE:
		return domain.ThemeApple
	case templatev1.ThemeKind_THEME_KIND_NOTION:
		return domain.ThemeNotion
	case templatev1.ThemeKind_THEME_KIND_GLASS:
		return domain.ThemeGlass
	case templatev1.ThemeKind_THEME_KIND_TERMINAL:
		return domain.ThemeTerminal
	case templatev1.ThemeKind_THEME_KIND_CUSTOM:
		return domain.ThemeCustom
	default:
		return ""
	}
}

func toProtoTemplate(t *domain.Template) *templatev1.Template {
	if t == nil {
		return nil
	}
	return &templatev1.Template{
		Id:         t.ID,
		OwnerId:    t.OwnerID,
		Name:       t.Name,
		Theme:      themeToProto(t.Theme),
		Subject:    t.Subject,
		CustomHtml: t.CustomHTML,
		CreatedAt:  ts(t.CreatedAt),
		UpdatedAt:  ts(t.UpdatedAt),
	}
}

func toProtoTheme(ti domain.ThemeInfo) *templatev1.Theme {
	return &templatev1.Theme{
		Kind:        themeToProto(ti.Kind),
		Name:        ti.Name,
		Description: ti.Description,
		PreviewHtml: ti.PreviewHTML,
	}
}
