# Development guide

Everything you need to run HaloMail — the whole app and each individual service —
on your machine.

## 1. Requirements

| Tool       | Version | Why                              | Install                                                   |
| ---------- | ------- | -------------------------------- | --------------------------------------------------------- |
| **Go**     | 1.24+   | backend services                 | `winget install GoLang.Go` · https://go.dev/dl            |
| **Node**   | 20+     | web, docs, SDK                   | https://nodejs.org                                        |
| **pnpm**   | 9+      | JS package manager               | `npm i -g pnpm`                                           |
| **Docker** | latest  | Postgres, Redis, Mailpit, Jaeger | https://docs.docker.com/get-docker                        |
| **buf**    | latest  | proto → Go/TS codegen            | https://buf.build/docs/installation                       |
| **goose**  | latest  | DB migrations                    | `go install github.com/pressly/goose/v3/cmd/goose@latest` |
| **task**   | latest  | task runner (optional)           | `go install github.com/go-task/task/v3/cmd/task@latest`   |

> ⚠️ **Go is not installed on this machine yet.** Install it first, then
> `go version` should print 1.24+. The web app and docs run without Go.

Verify:

```bash
go version && node --version && pnpm --version && docker --version && buf --version
```

## 2. Clone & configure

```bash
git clone https://github.com/aashishrajdev/halomail
cd halomail
cp .env.example .env
```

The defaults in `.env.example` work out of the box against the local Docker infra.
Fill in `RESEND_API_KEY`, Google/Microsoft OAuth, etc. only when you need those
integrations (locally, email is captured by Mailpit without a Resend key).

## 3. Start infra

```bash
task up        # or: docker compose up -d
```

| Container | Port(s)       | Purpose                         | UI                      |
| --------- | ------------- | ------------------------------- | ----------------------- |
| postgres  | 5432          | primary database                | —                       |
| redis     | 6379          | cache + rate limiting (optional)| —                       |
| mailpit   | 1025 / 8025   | catches outbound email in dev   | http://localhost:8025   |
| jaeger    | 4317/4318/16686 | OpenTelemetry traces          | http://localhost:16686  |

## 4. Generate code from proto

Required before the first backend build (creates `services/shared/gen/` and the
SDK in `packages/sdk-js/src/gen/`).

```bash
task proto      # or: buf generate
task proto:lint # validate contracts
```

## 5. Install deps & migrate the database

```bash
task bootstrap  # pnpm install + go mod download
task migrate    # goose up — applies each service's migrations
```

## 6. Run the backend

### Monolith mode (recommended for dev)

All services in one process on `:8080`:

```bash
task api:run
```

### A single service (distributed mode)

Each service is a standalone binary. Run only what you're working on.

| Service        | Command                                          | Port | Extra env it reads                          |
| -------------- | ------------------------------------------------ | ---- | ------------------------------------------- |
| `identity`     | `cd services/identity && go run ./cmd/server`     | 8081 | `JWT_SECRET`                                |
| `scheduling`   | `cd services/scheduling && go run ./cmd/server`   | 8082 | `GOOGLE_*`, `MICROSOFT_*`                    |
| `contact`      | `cd services/contact && go run ./cmd/server`      | 8083 | `RATELIMIT_*`                               |
| `template`     | `cd services/template && go run ./cmd/server`     | 8084 | —                                           |
| `notification` | `cd services/notification && go run ./cmd/server` | 8085 | `RESEND_API_KEY` or `SMTP_*`                |
| `gateway`      | `cd services/gateway && go run ./cmd/server`      | 8080 | `*_URL` for each upstream service            |

Set the port per service with `HTTP_PORT`, e.g.:

```bash
HTTP_PORT=8081 go run ./cmd/server
```

In distributed mode the gateway needs each upstream URL:

```bash
IDENTITY_URL=http://localhost:8081 \
SCHEDULING_URL=http://localhost:8082 \
CONTACT_URL=http://localhost:8083 \
TEMPLATE_URL=http://localhost:8084 \
NOTIFICATION_URL=http://localhost:8085 \
go run ./cmd/server
```

## 7. Run the frontend & docs

```bash
task web         # Next.js dashboard → http://localhost:3000
task docs        # docs site         → http://localhost:3001
```

## 8. Quality gates

```bash
task test        # go test ./... -race in every service module
task api:test    # same, gateway module only
task api:lint    # go vet + golangci-lint
task proto:lint  # buf lint
pnpm lint        # biome (web, docs, sdk)
```

## 9. Calling the API directly

ConnectRPC speaks plain JSON over HTTP, so `curl` works:

```bash
curl http://localhost:8080/halomail.identity.v1.AuthService/Login \
  -H 'Content-Type: application/json' \
  -d '{"email":"me@example.com","password":"secret"}'
```

Health probes:

```bash
curl http://localhost:8080/healthz   # liveness
curl http://localhost:8080/readyz    # readiness (db/redis)
```

## Troubleshooting

| Symptom                                   | Fix                                                                 |
| ----------------------------------------- | ------------------------------------------------------------------ |
| `go: command not found`                   | Install Go 1.24+ and reopen the terminal.                          |
| Build errors referencing `shared/gen`     | Run `task proto` — generated code is created by codegen.           |
| `postgres: ping: connection refused`      | `task up` and wait for the healthcheck; check `DATABASE_URL`.      |
| Port already in use                       | Set `HTTP_PORT` or stop the conflicting process.                   |
| Emails "disappear" locally                | They're in Mailpit → http://localhost:8025 (no real send in dev).  |
| Traces missing                            | Ensure Jaeger is up and `OTEL_EXPORTER_OTLP_ENDPOINT` is set.      |
