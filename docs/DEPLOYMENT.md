# Deployment — free & low-cost

HaloMail is designed to run at **$0/month** at small scale by using free tiers and
**monolith mode** (one container instead of six). Scale to distributed services
only when you actually need to.

## The free stack

| Concern        | Service            | Free tier                          |
| -------------- | ------------------ | ---------------------------------- |
| Backend compute| **Fly.io** / **Render** | shared CPU, scale-to-zero          |
| PostgreSQL     | **Neon**           | 0.5 GB, autosuspend                |
| Redis (opt.)   | **Upstash**        | 10k commands/day — or omit         |
| Email          | **Resend**         | 100/day, 3k/month                  |
| Web + docs     | **Vercel**         | hobby projects                     |
| Traces (opt.)  | **Grafana Cloud** / **Honeycomb** | free OTLP ingest    |

> Redis is **optional**: when `REDIS_URL` is empty, services use an in-memory
> rate limiter (per instance). Omit it entirely on free tier.

## 0. Prerequisites

- Accounts on the providers above.
- Generated code committed (`task proto`) so the Docker build is reproducible.
- A strong `JWT_SECRET`: `openssl rand -hex 32`.

## 1. Database — Neon

1. Create a Neon project; copy the pooled connection string.
2. Set it as `DATABASE_URL` (append `?sslmode=require`).
3. Run migrations against it once:
   ```bash
   DATABASE_URL="postgres://…neon…/db?sslmode=require" task migrate
   ```

## 2. Redis — Upstash (optional)

Create a database, copy the `rediss://` URL into `REDIS_URL`. Skip to use the
in-memory limiter.

## 3. Email — Resend

Create an API key and verify your sending domain. Set `RESEND_API_KEY` and
`EMAIL_FROM="HaloMail <noreply@yourdomain>"`. Without a key, the notification
service falls back to SMTP (`SMTP_HOST`/`SMTP_PORT`).

## 4. Backend container

A single multi-stage Dockerfile builds any service (or the monolith):

```bash
# Build the all-in-one monolith image (cheapest)
docker build -f deploy/Dockerfile --build-arg SERVICE=gateway -t halomail-api .
docker run -p 8080:8080 --env-file .env halomail-api
```

The runtime includes PostgreSQL backup/restore clients and the `/backup` CLI.
Neon is the only active database. With R2 credentials configured, the gateway
creates a full backup every hour and retains the latest two verified batches.
See [backup and recovery setup](BACKUP_AND_RECOVERY.md) for configuration,
monitoring and restoring the latest batch into a replacement Neon database.

### Deploy to Fly.io

```bash
fly launch --copy-config --no-deploy        # uses deploy/fly.toml
fly secrets set \
  JWT_SECRET=$(openssl rand -hex 32) \
  DATABASE_URL="postgres://…neon…?sslmode=require" \
  REDIS_URL="rediss://…upstash…" \
  RESEND_API_KEY="re_…" \
  EMAIL_FROM="HaloMail <noreply@yourdomain>"
fly deploy
```

`deploy/fly.toml` sets `auto_stop_machines` + `min_machines_running = 0` so the
app scales to zero and costs nothing when idle.

### Deploy to Render

Point Render at the repo; it reads `deploy/render.yaml` (a free Docker web
service with a `/healthz` check). Set the same secrets in the dashboard.

## 5. Frontend & docs — Vercel

Two Vercel projects from the same repo:

| Project | Root directory | Env                                   |
| ------- | -------------- | ------------------------------------- |
| web     | `apps/web`     | `NEXT_PUBLIC_API_URL=https://<api-host>` |
| docs    | `apps/docs`    | —                                     |

Vercel auto-detects Next.js and the pnpm workspace.

## 6. OAuth callback URLs

After the API has a public URL, register these redirect URIs:

- Google:    `https://<api-host>/v1/oauth/google/callback`
- Microsoft: `https://<api-host>/v1/oauth/microsoft/callback`

and set `GOOGLE_REDIRECT_URL` / `MICROSOFT_REDIRECT_URL` to match.

## 7. Observability (optional, free)

Point OTLP at a free backend instead of local Jaeger:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT="https://otlp-gateway-<region>.grafana.net/otlp"
# plus the provider's auth header env vars
```

## Production env checklist

```bash
APP_ENV=production
LOG_FORMAT=json
JWT_SECRET=<32+ byte secret>
DATABASE_URL=<neon pooled url, sslmode=require>
REDIS_URL=<upstash url, or empty>
RESEND_API_KEY=<resend key>
EMAIL_FROM="HaloMail <noreply@yourdomain>"
PUBLIC_API_URL=https://<api-host>
PUBLIC_WEB_URL=https://<web-host>
OTEL_ENABLED=true
```

## Scaling up later

When one component needs to scale independently, switch that service to its own
deployment (build with `--build-arg SERVICE=<name>`), give the gateway the new
`<SERVICE>_URL`, and leave the rest in the monolith. No code changes — see
[ARCHITECTURE.md](ARCHITECTURE.md#deployment-modes).

## Cost summary

| Scale                         | Monthly cost |
| ----------------------------- | ------------ |
| Hobby (free tiers, monolith)  | **$0**       |
| Small (always-on 256MB + Neon)| ~$5–10       |
| Growth (distributed services) | pay per service |
