# Forms and Meetings

## Existing code audit

The repository has Go identity, contact, scheduling, notification and template
services, mounted together by the gateway, with a Next.js dashboard. Before this
change, contact forwarding only logged messages, scheduling had no completed
calendar callback or event creation, the widget sent an outdated field, and
general-purpose API keys were unsuitable for embedding publicly. Public booking
lookups exposed booking-management tokens, and mail dispatch was publicly callable.

The new dashboard centers on **Forms** and **Meetings**. Legacy template,
analytics and message routes remain in the repository; they are not new product
features or a paid-plan implementation.

## Accounts, keys and allowances

- Dashboard requests require an authenticated user. Browser tokens are stored
  in HttpOnly, SameSite=Lax cookies, Secure in production, behind same-origin
  Next.js API routes. Old localStorage tokens are discarded.
- Each account can have one active Forms key and one active Meetings key.
  A PostgreSQL unique index enforces this even for concurrent requests.
- A full key is returned only on creation. The database stores its SHA-256 hash,
  prefix and last four characters. Later dashboard visits show only a mask;
  there is no reveal endpoint. Generate stays disabled until Delete succeeds.
- Embedded keys are **public identifiers with restricted capabilities**, not
  dashboard credentials. Forms keys submit forms; Meetings keys read public
  event availability and create bookings for their owner. Neither authorizes
  account management, reading private submissions or listing bookings.
- Default free allowances are **25 accepted, non-spam forms and 10 new bookings
  per UTC calendar month**. These are configurable defaults, not confirmed
  pricing decisions. Set `FREE_LIMITS_MONTHLY=false` for a lifetime allowance.
- Usage is charged atomically with a stored submission or reservation, not on
  email delivery. Key deletion, regeneration, cancellation and message deletion
  do not refund/reset it. Rescheduling does not charge a second booking.

## Forms flow

1. Register/sign in, open Forms and generate a key. Save it immediately.
2. Copy the generated HTML snippet into a website.
3. `POST /submit` accepts HTML form encoding, multipart text fields without
   files, or a flat JSON object containing `access_key` and arbitrary fields.
4. A valid submission is stored and its email is queued in the same database
   transaction. Notification workers send it to the account email, with the
   sender email as Reply-To. Form fields are escaped before HTML rendering.
5. The API returns JSON acceptance, not proof of email delivery. The dashboard
   lists recent messages; Mailpit catches local development mail.

Limits: 64 KiB requests, 50 fields, 100-character field names, 10,000-character
field values. `_hl_hp` is an optional honeypot. Nested JSON and uploads are
rejected. Spam is retained but not forwarded or charged. The existing widget
also accepts the new Forms key in `data-halomail`.

## Meetings flow

1. Generate a Meetings key and connect Google Calendar from the Meetings page.
2. Create a meeting type, enable your working days and save hours/timezone.
   Empty availability means no bookable slots.
3. Copy the button link containing the key. Guests select an event type/time,
   then provide name, email and discussion notes.
4. Availability checks local bookings and the connected primary Google calendar.
   Owner-level transaction locks prevent concurrent HaloMail double bookings.
5. An accepted reservation consumes credit and enters calendar synchronization.
   The worker creates a deterministic Google event with a Meet conference,
   then queues host/guest confirmation emails containing the link and a
   cancellation link. Google also sends its calendar invitation.
6. The UI distinguishes a reservation awaiting calendar confirmation from one
   with a meeting link. Calendar/email failures retry without charging again.
   Cancellation releases the HaloMail slot immediately and queues calendar and
   email updates. Opening a cancellation link alone does not cancel a booking.

Google Calendar and Google Meet are the implemented provider. Outlook is not
implemented in this flow. This is not full cal.com parity: routing forms,
custom booking questions, teams, billing and a reschedule UI are not included.
The existing reschedule RPC checks calendar availability and queues an update.

## Setup

1. Back up an existing database, then run `task migrate` before starting the
   updated services. New migrations are identity `00002_feature_usage`,
   scheduling `00002_calendar_jobs` and notification `00002_mail_outbox`.
   All services currently share one database. Apply all migrations before
   accepting traffic, including when running standalone services.
2. Set a random `JWT_SECRET` of at least 32 characters. Configure allowances in
   the root `.env` using `.env.example`. Do not deploy its placeholder secret.
3. Configure `RESEND_API_KEY` and a verified `EMAIL_FROM`, or local Mailpit SMTP.
   Run the gateway (includes workers), or run notification and scheduling
   workers alongside their standalone services.
4. Enable Google Calendar API for your OAuth application. Configure its consent
   screen and authorized users as appropriate. Set `GOOGLE_CLIENT_ID`,
   `GOOGLE_CLIENT_SECRET`, and `GOOGLE_REDIRECT_URL` to the **web app** callback:
   `http://localhost:3000/api/calendar/callback` locally, HTTPS in production.
5. Generate a dedicated 32-byte encryption key, base64 encoded, for
   `CALENDAR_ENCRYPTION_KEY` (for example `openssl rand -base64 32`). Keep it
   stable and secret. Changing it requires calendar reconnection. OAuth refresh
   tokens are AES-GCM encrypted with the account ID bound to the ciphertext.
6. Copy `apps/web/.env.example` to `apps/web/.env.local`. `API_BASE_URL` is the
   server-to-server gateway URL. `NEXT_PUBLIC_API_URL` is the browser-reachable
   API origin used in embedded form snippets. `PUBLIC_WEB_URL` must be the
   exact public web origin, without a trailing slash, and must also match the
   backend setting used for cancellation links.

Existing generic API keys no longer authenticate dashboard management or these
public features. Generate feature-specific keys and replace old website embeds.
The new unique index deliberately fails if historical data already contains
duplicate active single-feature keys: review and revoke duplicates first rather
than silently choosing a key to invalidate.

## Validation and production boundaries

Run backend tests from the repository root:

```sh
go test ./services/identity/... ./services/contact/... ./services/scheduling/... ./services/notification/... ./services/shared/... ./services/gateway/... ./services/template/...
corepack pnpm --filter @halomail/web build
```

Set `TEST_DATABASE_URL` to a **local test PostgreSQL database** and run
`go test ./services/shared/usage -run TestPostgres -v` to test migrations, key
revocation/scope isolation and concurrent quota consumption. The test creates
and drops only its randomly named schema. Without this variable it is skipped.

Before public launch, perform an end-to-end test using your OAuth application,
calendar and mail provider. No software can honestly be certified “100% secure”
by this implementation or a code scan. In particular:

- Add account email verification and stronger automated-abuse controls (such
  as CAPTCHA) before unrestricted registration. Public keys alone do not
  prevent spam, deliberate quota exhaustion or fake invitee addresses.
- Configure HTTPS, secure secret storage, retention/cleanup, monitoring and
  backups. Audit the remaining legacy endpoints and dependencies separately.
- The new HTTP throttle uses the actual network peer, not untrusted forwarding
  headers; requests through the web proxy share that peer budget. Production
  needs trusted-edge, per-client and distributed throttling appropriate to its
  proxy topology. It is not a DDoS defense.
- Google busy checks and local reservations cannot be one atomic transaction.
  An external calendar edit racing a booking can still introduce a conflict.
  Only the primary connected calendar is checked.
- Monitor pending `bookings.calendar_pending` and undelivered `mail_outbox`
  rows. Revoked Google consent, unsupported conferencing or provider outages
  require operator/user action; jobs retry rather than pretending to confirm.
- Mail delivery is at least once. Resend receives an idempotency key; SMTP or
  ambiguous failures can still duplicate mail. A queued reservation is charged
  even if an external provider later fails permanently.
- Logout revokes refresh access; an already issued bearer JWT remains valid
  until its 15-minute expiry. Reset/recovery, verified emails and billing are
  not implemented by this change.

Google conference creation follows the
[Calendar events insert API](https://developers.google.com/workspace/calendar/api/v3/reference/events/insert)
and authorization follows the
[OAuth web-server flow](https://developers.google.com/identity/protocols/oauth2/web-server).
