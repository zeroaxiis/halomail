// Package outbox queues transactional mail in the same database transaction
// as the change that triggers it. The notification mail worker drains the
// mail_outbox table and delivers each row.
package outbox

import (
	"context"
	"html"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Queue inserts one mail into the outbox. Reusing an id is a no-op, so callers
// can retry safely.
func Queue(ctx context.Context, tx pgx.Tx, id, recipient, replyTo, subject, body string) error {
	_, err := tx.Exec(ctx, `INSERT INTO mail_outbox(id,recipient,reply_to,subject,html) VALUES($1,$2,$3,$4,$5) ON CONFLICT(id) DO NOTHING`, id, recipient, replyTo, subject, body)
	return err
}

// Fields renders heading and the submitted fields into a minimal, clean
// email theme matching the team invite and password reset templates.
func Fields(heading string, fields map[string]string) string {
	var body strings.Builder
	title := html.EscapeString(heading)
	if title == "" {
		title = "New form submission"
	}

	body.WriteString(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>` + title + `</title></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f9fafb; padding: 40px 20px; color: #111827; margin: 0;">
  <div style="max-width: 520px; margin: 0 auto; background: #ffffff; border: 1px solid #e5e7eb; border-radius: 8px; padding: 32px;">
    <h2 style="font-size: 20px; font-weight: 600; margin-top: 0; margin-bottom: 8px; color: #111827;">` + title + `</h2>
    <p style="font-size: 14px; line-height: 1.5; color: #6b7280; margin: 0 0 24px;">You received a new submission from your website form:</p>
    <table role="presentation" style="width: 100%; border-collapse: collapse; margin-bottom: 24px; font-size: 14px;">
      <tbody>`)

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := fields[key]
		body.WriteString(`
        <tr style="border-bottom: 1px solid #f3f4f6;">
          <td style="padding: 12px 0; color: #6b7280; width: 34%; vertical-align: top; font-weight: 500;">` + html.EscapeString(key) + `</td>
          <td style="padding: 12px 0; color: #111827; width: 66%; vertical-align: top; white-space: pre-wrap; word-break: break-word;">` + html.EscapeString(val) + `</td>
        </tr>`)
	}

	body.WriteString(`
      </tbody>
    </table>
    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0 20px;" />
    <div style="text-align: center;">
      <table role="presentation" cellpadding="0" cellspacing="0" border="0" style="margin: 0 auto; text-align: center;">
        <tr>
          <td align="center" style="padding-bottom: 6px;">
            <img src="https://halomail.zeroaxiis.tech/brand/zeroaxiis.png" alt="ZeroAxiis" width="24" height="24" style="display: block; width: 24px; height: 24px; border: 0; margin: 0 auto;" />
          </td>
        </tr>
        <tr>
          <td align="center" style="font-size: 12px; color: #6b7280; line-height: 1.5;">
            A product of <a href="https://zeroaxiis.tech" style="color: #111827; font-weight: 600; text-decoration: none;">ZeroAxiis</a>
          </td>
        </tr>
        <tr>
          <td align="center" style="font-size: 11px; color: #9ca3af; font-style: italic; padding-top: 2px;">
            Everything starts from zero.
          </td>
        </tr>
      </table>
    </div>
  </div>
</body>
</html>`)

	return body.String()
}
