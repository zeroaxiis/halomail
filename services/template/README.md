# template — email themes, render, preview

The email designer: built-in themes plus custom HTML, with live preview. Used by
`scheduling` (confirmations) and `notification` (delivery).

- **Dev port:** `8084`
- **Proto:** [`proto/halomail/template/v1`](../../proto/halomail/template/v1/template.proto)
- **Module:** `github.com/zeroaxiis/halomail/services/template`

## Services / RPCs

| Service           | RPCs                                                        |
| ----------------- | ---------------------------------------------------------- |
| `TemplateService` | ListThemes, ListTemplates, GetTemplate, Create/Update/Delete, RenderPreview |

## Built-in themes

| Theme      | Look                                                |
| ---------- | --------------------------------------------------- |
| `MINIMAL`  | Clean, system-font, lots of whitespace              |
| `APPLE`    | SF-style, soft shadows, rounded                     |
| `NOTION`   | Document-like, neutral, subtle dividers             |
| `GLASS`    | Translucent panels, gradient accents                |
| `TERMINAL` | Monospace, dark, developer aesthetic                |
| `CUSTOM`   | Author-supplied HTML                                |

Themes are **code**, not rows: each is a Go `html/template` under
`internal/themes/`. Rendering substitutes `{{variables}}` and inlines CSS so the
output works across email clients.

## Render flow

```
RenderPreview(theme | custom_html, subject, variables)
  ─ select theme template (or parse custom HTML)
  ─ execute with variables (auto-escaped)
  ─ inline CSS, wrap in email-safe boilerplate
  ─ return { subject, html }
```

## Data model

| Table       | Notes                                                  |
| ----------- | ------------------------------------------------------ |
| `templates` | owner_id, name, theme, subject, custom_html, timestamps |

## Run

```bash
cd services/template && HTTP_PORT=8084 go run ./cmd/server
```
