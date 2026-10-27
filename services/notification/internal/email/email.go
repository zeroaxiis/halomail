// Package email sends transactional mail via Resend (HTTP API) or SMTP
// (Mailpit locally). The Sender interface lets callers stay provider-agnostic.
package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/zeroaxiis/halomail/services/shared/emailpool"
	"github.com/zeroaxiis/halomail/services/shared/idgen"
)

// Message is a transactional email.
type Message struct {
	ID      string
	To      []string
	From    string
	ReplyTo string
	Subject string
	HTML    string
	Text    string
}

// Sender delivers a Message, returning a provider message id and the provider name.
type Sender interface {
	Send(ctx context.Context, msg Message) (id, provider string, err error)
}

// ---- Resend --------------------------------------------------------------

// ResendSender delivers mail through the Resend HTTP API using a multi-account pool.
type ResendSender struct {
	pool *emailpool.Pool
}

func cleanSender(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	return strings.TrimSpace(s)
}

// NewResend returns a sender that uses a multi-account pool loaded from environment.
func NewResend(apiKey, defaultFrom string) *ResendSender {
	return &ResendSender{pool: emailpool.NewFromEnv()}
}

// Send posts msg to Resend using round-robin and automatic failover across accounts.
func (s *ResendSender) Send(ctx context.Context, msg Message) (string, string, error) {
	return s.pool.Send(ctx, emailpool.Message{
		ID:      msg.ID,
		To:      msg.To,
		From:    msg.From,
		ReplyTo: msg.ReplyTo,
		Subject: msg.Subject,
		HTML:    msg.HTML,
		Text:    msg.Text,
	})
}

// ---- SMTP (Mailpit / any relay) -----------------------------------------

// SMTPSender delivers mail over SMTP, upgrading to TLS when the server offers
// STARTTLS.
type SMTPSender struct {
	addr        string
	defaultFrom string
}

// NewSMTP returns a sender for the relay at host:port.
func NewSMTP(host string, port int, defaultFrom string) *SMTPSender {
	return &SMTPSender{addr: fmt.Sprintf("%s:%d", host, port), defaultFrom: defaultFrom}
}

// Send delivers msg as an HTML mail. Addresses containing line breaks are
// rejected to prevent header injection.
func (s *SMTPSender) Send(ctx context.Context, msg Message) (string, string, error) {
	from := firstNonEmpty(msg.From, s.defaultFrom)
	fromAddr := addressOnly(from)
	for _, header := range append(append([]string{}, msg.To...), from, msg.ReplyTo) {
		if strings.ContainsAny(header, "\r\n") {
			return "", "smtp", fmt.Errorf("invalid mail address")
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(msg.To, ", "))
	if msg.ReplyTo != "" {
		fmt.Fprintf(&b, "Reply-To: %s\r\n", msg.ReplyTo)
	}
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", strings.NewReplacer("\r", " ", "\n", " ").Replace(msg.Subject)))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	if msg.HTML != "" {
		b.WriteString(msg.HTML)
	} else {
		b.WriteString(msg.Text)
	}

	connection, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", s.addr)
	if err != nil {
		return "", "smtp", err
	}
	defer connection.Close()
	stop := context.AfterFunc(ctx, func() { _ = connection.Close() })
	defer stop()
	if err = connection.SetDeadline(time.Now().Add(20 * time.Second)); err != nil {
		return "", "smtp", err
	}
	host, _, err := net.SplitHostPort(s.addr)
	if err != nil {
		return "", "smtp", err
	}
	client, err := smtp.NewClient(connection, host)
	if err != nil {
		return "", "smtp", err
	}
	defer client.Close()
	if supported, _ := client.Extension("STARTTLS"); supported {
		if err = client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return "", "smtp", err
		}
	}
	if err = client.Mail(fromAddr); err != nil {
		return "", "smtp", err
	}
	for _, recipient := range msg.To {
		if err = client.Rcpt(recipient); err != nil {
			return "", "smtp", err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return "", "smtp", err
	}
	if _, err = io.WriteString(writer, b.String()); err != nil {
		return "", "smtp", err
	}
	if err = writer.Close(); err != nil {
		return "", "smtp", err
	}
	_ = client.Quit()
	return idgen.Prefixed("eml_"), "smtp", nil
}

func addressOnly(from string) string {
	if a, err := mail.ParseAddress(from); err == nil {
		return a.Address
	}
	return from
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}
