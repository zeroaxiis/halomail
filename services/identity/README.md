# identity — auth, users, API keys, audit

Owns who the user is and how callers prove it.

- **Dev port:** `8081`
- **Proto:** [`proto/halomail/identity/v1`](../../proto/halomail/identity/v1/identity.proto)
- **Module:** `github.com/aashishrajdev/halomail/services/identity`

## Services / RPCs

| Service          | RPCs                                                            |
| ---------------- | -------------------------------------------------------------- |
| `AuthService`    | Register, Login, Logout, RefreshSession, GetCurrentUser, VerifyToken (internal) |
| `UserService`    | GetUser, GetUserByHandle (public profile), UpdateUser          |
| `ApiKeyService`  | CreateApiKey, ListApiKeys, RevokeApiKey, VerifyApiKey (internal)|
| `AuditService`   | ListAuditLogs, RecordAuditLog (internal)                       |

## Data model

| Table          | Notes                                                         |
| -------------- | ------------------------------------------------------------ |
| `orgs`         | tenant; a user belongs to one org                            |
| `users`        | email (unique), password hash (argon2id), handle, timezone  |
| `sessions`     | refresh tokens (hashed), expiry, revoked                    |
| `api_keys`     | hashed secret, prefix, last_four, scopes, last_used_at      |
| `audit_logs`   | actor, action, target, metadata (jsonb), ip, created_at     |

## Security notes

- Login requires a password followed by a six-digit email OTP. Registration
  creates an account without issuing a session; the legacy password-only Login
  RPC rejects requests. Existing sessions retain their normal expiry.
- Redis stores an HMAC of each OTP for **600 seconds**. Verification atomically
  consumes the code, and five incorrect attempts invalidate it. Resending
  replaces the previous code, with a 60-second cooldown and at most ten requests
  per email per ten minutes. Redis failure prevents new logins.

- Passwords hashed with **argon2id**; never logged (redacted by `shared/log`).
- API key secrets and refresh tokens are stored **hashed**; the plaintext is
  returned exactly once at creation.
- JWT access tokens are short-lived; refresh tokens rotate on use.

## Configuration

### Email OTP setup

1. Set `REDIS_URL` in the gateway/identity environment.
2. Generate a random secret (`openssl rand -hex 32`) and set the same
   `OTP_DELIVERY_SECRET` in the gateway/identity environment and `apps/web/.env`.
3. In Resend Domains, verify **email.zeroaxiis.tech** for sending using the DNS
   records Resend provides. The sender's exact domain must match; verifying
   only `zeroaxiis.tech` does not verify this subdomain. No separate `no-reply`
   mailbox or sender identity needs to be created.
4. Set `RESEND_API_KEY` and `EMAIL_FROM="HaloMail <no-reply@email.zeroaxiis.tech>"`
   in `apps/web/.env`. The key must have sending permission for that domain.
   Next.js reads this app's environment, not the repository root `.env`.
   Configure the same variables on the web deployment's hosting platform.
5. The OTP sender uses the Resend HTTP API when the key is configured and
   requires it in production. Without a key in development, Nodemailer uses
   local Mailpit at localhost:1025; captured emails appear at http://localhost:8025.
6. Restart the gateway and Next.js after changing configuration. Try logging in,
   enter the code, and check Resend Emails for delivery or bounce details.
   To preview the branded email without creating a login session, run
   `pnpm --filter @halomail/web sample:otp recipient@example.com`; the code in
   that sample is clearly marked as unusable for sign in.

The browser posts `{email, password}` to `/api/rpc/v1/auth/otp/request`, then
`{challengeId, code}` to `/api/rpc/v1/auth/otp/verify`. Only verification sets
the HttpOnly session cookies. Request responses contain a challenge ID and
expiry duration, never the code or session tokens.

The Go `/v1/auth/otp/request` and `/v1/auth/otp/cancel` endpoints are internal
delivery APIs authenticated with `X-OTP-Delivery-Secret`. The Next.js server
sends the code through Resend and cancels the challenge on delivery failure.
Use HTTPS between services in production and never log delivery response bodies.

Sender setup: https://resend.com/docs/knowledge-base/how-do-I-create-an-email-address-or-sender-in-resend

Run auth checks from the repository root with
`go test ./services/identity/... ./services/shared/server/...` and
`pnpm --filter @halomail/web test:auth`.

| Env          | Purpose                          |
| ------------ | -------------------------------- |
| `JWT_SECRET` | HMAC signing key (32+ bytes)     |
| `SESSION_TTL`| refresh token lifetime           |
| `API_KEY_PREFIX` | key prefix, e.g. `hl_`       |

## Run

```bash
cd services/identity && HTTP_PORT=8081 go run ./cmd/server
```
