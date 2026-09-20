package outbox

import (
	"context"
	"github.com/jackc/pgx/v5"
	"html"
	"sort"
	"strings"
)

func Queue(ctx context.Context, tx pgx.Tx, id, recipient, replyTo, subject, body string) error {
	_, err := tx.Exec(ctx, `INSERT INTO mail_outbox(id,recipient,reply_to,subject,html) VALUES($1,$2,$3,$4,$5) ON CONFLICT(id) DO NOTHING`, id, recipient, replyTo, subject, body)
	return err
}

func Fields(heading string, fields map[string]string) string {
	var body strings.Builder
	body.WriteString("<h1>" + html.EscapeString(heading) + "</h1><dl>")
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		body.WriteString("<dt><strong>" + html.EscapeString(key) + "</strong></dt><dd style=\"white-space:pre-wrap\">" + html.EscapeString(fields[key]) + "</dd>")
	}
	body.WriteString("</dl>")
	return body.String()
}
