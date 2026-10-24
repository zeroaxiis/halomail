# notification — email delivery & webhooks

Transactional email (Resend / SMTP) and developer webhook fan-out.

- **Dev port:** `8085`
- **Proto:** [`proto/halomail/notification/v1`](../../proto/halomail/notification/v1/notification.proto)
- **Module:** `github.com/zeroaxiis/halomail/services/notification`

## Services / RPCs

| Service           | RPCs                                                       |
| ----------------- | --------------------------------------------------------- |
| `EmailService`    | SendEmail (internal)                                       |
| `WebhookService`  | CreateWebhook, ListWebhooks, DeleteWebhook, RotateSecret, Dispatch (internal) |

## Email delivery

```
SendEmail(to, subject, html, text)
  ─ if RESEND_API_KEY set → Resend API
  ─ else                  → SMTP (Mailpit locally)
  ─ record delivery; return provider + id
```

## Webhooks

- Subscriptions store a signing secret (shown once). Deliveries are signed with
  `HMAC-SHA256` over the body in an `X-HaloMail-Signature` header.
- `Dispatch(owner, event, payload)` enqueues deliveries to each matching
  subscription; a worker delivers with retries and exponential backoff.
- Delivery attempts and status codes are recorded for the dashboard.

## Events

`BOOKING_CREATED` · `BOOKING_CANCELLED` · `BOOKING_RESCHEDULED` · `MESSAGE_RECEIVED`

## Data model

| Table                | Notes                                            |
| -------------------- | ------------------------------------------------ |
| `webhooks`           | url, events, hashed secret, active               |
| `webhook_deliveries` | webhook_id, event, status, response_code, attempts |

## Configuration

| Env                | Purpose                                  |
| ------------------ | ---------------------------------------- |
| `RESEND_API_KEY`   | use Resend when set                      |
| `EMAIL_FROM`       | default From address                     |
| `SMTP_HOST`/`PORT` | fallback transport (Mailpit in dev)      |

## Run

```bash
cd services/notification && HTTP_PORT=8085 go run ./cmd/server
```
