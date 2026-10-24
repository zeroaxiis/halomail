# contact — forms, messages, spam, forwarding

Embeddable contact-form infrastructure for portfolios and sites.

- **Dev port:** `8083`
- **Proto:** [`proto/halomail/contact/v1`](../../proto/halomail/contact/v1/contact.proto)
- **Module:** `github.com/zeroaxiis/halomail/services/contact`

## Services / RPCs

| Service          | RPCs                                                          |
| ---------------- | ------------------------------------------------------------ |
| `FormService`    | Create/Get/List/Update/Delete Form                           |
| `MessageService` | SubmitMessage (public), ListMessages, GetMessage, MarkRead, DeleteMessage |

## Submission pipeline

```
POST (widget / REST / SDK)
  ─ rate limit (per IP + per form, shared/ratelimit)
  ─ spam check:
      · honeypot field must be empty
      · heuristics (links, keywords, timing) → spam_score
      · optional reCAPTCHA verification
  ─ store message (flagged is_spam if over threshold)
  ─ notification.SendEmail → forward to form.target_email
  ─ notification.Dispatch(MESSAGE_RECEIVED) → webhooks
  ─ redirect or JSON ack
```

## Embeddable widget

A tiny script renders the form and posts to the public endpoint:

```html
<script src="https://<api-host>/widget.js" data-form="my-form-slug" defer></script>
```

No-JS fallback: a plain `<form action=".../v1/forms/{slug}/submit" method="post">`.

## Data model

| Table      | Notes                                                       |
| ---------- | ---------------------------------------------------------- |
| `forms`    | slug, target_email, spam_protection, fields (jsonb), active |
| `messages` | form_id, sender, data (jsonb), ip, ua, spam_score, is_spam, read |

## Configuration

| Env             | Purpose                          |
| --------------- | -------------------------------- |
| `RATELIMIT_*`   | public submit limits             |

## Run

```bash
cd services/contact && HTTP_PORT=8083 go run ./cmd/server
```
