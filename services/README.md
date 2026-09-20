# HaloMail — Backend services

The backend is a set of independent, modular Go services that share one platform
library (`shared`) and communicate over ConnectRPC. Each service is its own Go
module with clean-architecture internals.

## Services

| Service        | Dir                       | Dev port | Responsibility                                              | README |
| -------------- | ------------------------- | -------- | ---------------------------------------------------------- | ------ |
| `gateway`      | [gateway](gateway)         | 8080     | Public edge: auth, rate limit, REST/OpenAPI, BFF aggregation | [↗](gateway/README.md) |
| `identity`     | [identity](identity)       | 8081     | Auth (sessions/JWT), users, orgs, API keys, audit log       | [↗](identity/README.md) |
| `scheduling`   | [scheduling](scheduling)   | 8082     | Availability, bookings, timezone, Google sync               | [↗](scheduling/README.md) |
| `contact`      | [contact](contact)         | 8083     | Contact forms, messages, spam protection, forwarding        | [↗](contact/README.md) |
| `template`     | [template](template)       | 8084     | Email themes, rendering, live preview                       | [↗](template/README.md) |
| `notification` | [notification](notification) | 8085   | Resend email delivery, webhook dispatch                     | [↗](notification/README.md) |
| `shared`       | [shared](shared)           | —        | Platform lib: config, logging, OTel, db, redis, interceptors, gen code | [↗](shared/README.md) |

## Clean architecture (per service)

```
services/<name>/
├─ cmd/server/main.go      # composition root: wire config → adapters → app → rpc
├─ internal/
│  ├─ domain/              # entities + business rules (no I/O, no imports out)
│  ├─ app/                 # use cases / application services
│  └─ adapters/
│     ├─ postgres/         # repository implementations (pgx)
│     └─ rpc/              # ConnectRPC handlers (map proto ⇄ domain)
└─ migrations/             # goose SQL
```

Dependencies point inward. The domain knows nothing about Postgres or ConnectRPC;
adapters depend on the domain, never the reverse.

## Two deployment modes (same code)

- **Monolith mode (default, cheapest):** one binary mounts every service's
  ConnectRPC handler on a single port. Ideal for local dev and free-tier hosting.
  → `task api:run`
- **Distributed mode:** run each service as its own process/container on its own
  port; the gateway calls them over the network.
  → `cd services/identity && go run ./cmd/server` (repeat per service)

See [../docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md#deployment-modes).

## Running

```bash
# from repo root
task up            # infra (postgres, redis, mailpit, jaeger)
task proto         # generate gen/ code (required before first build)
task migrate       # apply migrations
task api:run       # monolith mode on :8080

# a single service (distributed mode)
cd services/identity && go run ./cmd/server
```

## Shared library

`services/shared` provides everything a service needs to boot consistently:

| Package          | Purpose                                            |
| ---------------- | -------------------------------------------------- |
| `config`         | Env/.env configuration (superset for all services) |
| `log`            | Structured slog: trace-correlated, redacting       |
| `observability`  | OpenTelemetry tracer setup                          |
| `postgres`       | pgx pool                                            |
| `redis`          | go-redis client                                    |
| `connectutil`    | RPC interceptors (otel, recovery, logging) + error mapping |
| `health`         | Liveness/readiness HTTP probes                     |
| `server`         | h2c HTTP server with graceful shutdown             |
| `errs`           | Transport-agnostic domain errors                   |
| `idgen`          | Sortable UUIDv7 ids                                |
| `gen`            | Generated protobuf + ConnectRPC code               |

## Testing

```bash
task api:test      # go test ./... -race across all modules
```
