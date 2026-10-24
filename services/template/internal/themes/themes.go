// Package themes renders transactional emails in built-in visual styles. It is
// pure: one email-safe layout driven by per-theme style tokens, with simple
// {{variable}} substitution (values HTML-escaped).
package themes

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/zeroaxiis/halomail/services/template/internal/domain"
)

// Data is the content rendered into a theme.
type Data struct {
	Heading    string
	Body       string
	ButtonText string
	ButtonURL  string
	Footer     string
	BrandName  string
}

type style struct {
	PageBG, CardBG, Font, HeadingFont string
	Text, Muted, Accent, AccentText, Border, Radius string
}

var styles = map[string]style{
	domain.ThemeMinimal: {
		PageBG: "#ffffff", CardBG: "#ffffff", Font: "-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif", HeadingFont: "inherit",
		Text: "#111111", Muted: "#666666", Accent: "#111111", AccentText: "#ffffff", Border: "#eaeaea", Radius: "6px",
	},
	domain.ThemeApple: {
		PageBG: "#f5f5f7", CardBG: "#ffffff", Font: "-apple-system,'SF Pro Text',BlinkMacSystemFont,sans-serif", HeadingFont: "inherit",
		Text: "#1d1d1f", Muted: "#86868b", Accent: "#0071e3", AccentText: "#ffffff", Border: "#d2d2d7", Radius: "18px",
	},
	domain.ThemeNotion: {
		PageBG: "#ffffff", CardBG: "#ffffff", Font: "Georgia,'Iowan Old Style',serif", HeadingFont: "inherit",
		Text: "#37352f", Muted: "#9b9a97", Accent: "#37352f", AccentText: "#ffffff", Border: "#e9e9e7", Radius: "3px",
	},
	domain.ThemeGlass: {
		PageBG: "linear-gradient(135deg,#7c83ff,#b69bff 60%,#ff9bd6)", CardBG: "rgba(255,255,255,0.16)", Font: "-apple-system,'Segoe UI',sans-serif", HeadingFont: "inherit",
		Text: "#ffffff", Muted: "rgba(255,255,255,0.82)", Accent: "#ffffff", AccentText: "#5b61e0", Border: "rgba(255,255,255,0.3)", Radius: "16px",
	},
	domain.ThemeTerminal: {
		PageBG: "#0a0a0a", CardBG: "#0f0f10", Font: "'Geist Mono',ui-monospace,'SF Mono',Menlo,monospace", HeadingFont: "inherit",
		Text: "#34d399", Muted: "#5a7d6f", Accent: "#34d399", AccentText: "#0a0a0a", Border: "#1f2a25", Radius: "4px",
	},
}

var varRe = regexp.MustCompile(`{{\s*([a-zA-Z0-9_]+)\s*}}`)

func replaceVars(s string, vars map[string]string) string {
	return varRe.ReplaceAllStringFunc(s, func(m string) string {
		key := strings.Trim(m, "{} ")
		if v, ok := vars[key]; ok {
			return html.EscapeString(v)
		}
		return ""
	})
}

// Render renders a built-in theme. Unknown kinds fall back to minimal.
func Render(kind, subject string, vars map[string]string) (renderedSubject, renderedHTML string) {
	subj := firstNonEmpty(replaceVars(subject, vars), vars["subject"], "Hello from HaloMail")
	s, ok := styles[kind]
	if !ok {
		s = styles[domain.ThemeMinimal]
	}
	return subj, renderLayout(s, dataFrom(vars))
}

// RenderCustom renders author-supplied HTML with variable substitution.
func RenderCustom(customHTML, subject string, vars map[string]string) (renderedSubject, renderedHTML string) {
	return firstNonEmpty(replaceVars(subject, vars), "Hello from HaloMail"), replaceVars(customHTML, vars)
}

func renderLayout(s style, d Data) string {
	var paras strings.Builder
	for _, p := range splitParas(d.Body) {
		fmt.Fprintf(&paras, `<p style="margin:0 0 16px;color:%s;font-size:15px;line-height:1.6">%s</p>`, s.Text, html.EscapeString(p))
	}

	button := ""
	if strings.TrimSpace(d.ButtonText) != "" {
		button = fmt.Sprintf(`<a href="%s" style="display:inline-block;margin-top:8px;padding:11px 22px;background:%s;color:%s;text-decoration:none;border-radius:%s;font-size:14px;font-weight:600">%s</a>`,
			html.EscapeString(d.ButtonURL), s.Accent, s.AccentText, s.Radius, html.EscapeString(d.ButtonText))
	}

	return fmt.Sprintf(`<!doctype html><html><body style="margin:0;padding:32px 16px;background:%s;font-family:%s">
  <div style="max-width:520px;margin:0 auto;background:%s;border:1px solid %s;border-radius:%s;padding:32px">
    <div style="font-size:13px;font-weight:600;letter-spacing:-0.01em;color:%s;margin-bottom:24px">%s</div>
    <h1 style="margin:0 0 16px;font-size:24px;line-height:1.25;color:%s;font-family:%s">%s</h1>
    %s%s
    <div style="margin-top:28px;padding-top:16px;border-top:1px solid %s;color:%s;font-size:12px">%s</div>
  </div>
</body></html>`,
		s.PageBG, s.Font,
		s.CardBG, s.Border, s.Radius,
		s.Accent, html.EscapeString(d.BrandName),
		s.Text, s.HeadingFont, html.EscapeString(d.Heading),
		paras.String(), button,
		s.Border, s.Muted, html.EscapeString(d.Footer),
	)
}

// Themes returns the built-in gallery with rendered preview HTML.
func Themes() []domain.ThemeInfo {
	meta := map[string][2]string{
		domain.ThemeMinimal:  {"Minimal", "Clean, system font, lots of whitespace."},
		domain.ThemeApple:    {"Apple", "SF-style, soft shadows, rounded card."},
		domain.ThemeNotion:   {"Notion", "Document-like, serif, subtle dividers."},
		domain.ThemeGlass:    {"Glass", "Translucent panel over a gradient."},
		domain.ThemeTerminal: {"Terminal", "Monospace, dark, developer aesthetic."},
	}
	sample := map[string]string{
		"heading":     "Your meeting is confirmed",
		"body":        "Hi Grace,\nYour 30-minute intro call is booked for Monday at 09:00.",
		"button_text": "Add to calendar",
		"button_url":  "https://halomail.dev",
		"footer":      "Reschedule or cancel anytime from the link in your inbox.",
	}
	out := make([]domain.ThemeInfo, 0, len(domain.BuiltinThemes))
	for _, k := range domain.BuiltinThemes {
		m := meta[k]
		_, preview := Render(k, "Your meeting is confirmed", sample)
		out = append(out, domain.ThemeInfo{Kind: k, Name: m[0], Description: m[1], PreviewHTML: preview})
	}
	return out
}

// ---- helpers -------------------------------------------------------------

func dataFrom(vars map[string]string) Data {
	return Data{
		Heading:    firstNonEmpty(vars["heading"], "Welcome to HaloMail"),
		Body:       firstNonEmpty(vars["body"], "This is a live preview of your email theme. Edit the heading, body, and button through template variables."),
		ButtonText: vars["button_text"],
		ButtonURL:  firstNonEmpty(vars["button_url"], "#"),
		Footer:     firstNonEmpty(vars["footer"], "You're receiving this because you use HaloMail."),
		BrandName:  firstNonEmpty(vars["brand"], "HaloMail"),
	}
}

func splitParas(body string) []string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	var out []string
	for _, p := range strings.Split(body, "\n") {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		out = []string{""}
	}
	return out
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}
