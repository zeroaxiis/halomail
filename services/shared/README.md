# shared — platform library

Common building blocks every HaloMail service uses to boot consistently. Not a
running service; imported as a Go module:
`github.com/aashishrajdev/halomail/services/shared`.

## Packages

| Package         | What it provides                                                   |
| --------------- | ----------------------------------------------------------------- |
| `config`        | `Load()` — env + `.env` into a typed `Config` (superset for all)  |
| `log`           | `New(Options)` — structured slog; trace-correlated, secret-redacting, context-scoped. See [log/README.md](log/README.md) |
| `observability` | `Setup(ctx, Config)` — OpenTelemetry tracer → OTLP                 |
| `postgres`      | `New(ctx, url)` — verified pgx pool                                |
| `redis`         | `New(ctx, url)` — verified go-redis client                        |
| `ratelimit`     | token-bucket limiter; Redis-backed or in-memory fallback          |
| `connectutil`   | `Default(logger)` interceptors (otel + recovery + logging); `ToConnect(err)` error mapping |
| `health`        | `Checker` with `/healthz` (liveness) + `/readyz` (readiness)      |
| `server`        | `Run(ctx, addr, handler, logger)` — h2c server, graceful shutdown |
| `errs`          | transport-agnostic domain errors (`NotFound`, `Invalid`, …)       |
| `idgen`         | `New()` — sortable UUIDv7 ids                                      |
| `authn`         | `Verifier` — local HS256 access-token validation                   |
| `httpx`         | JSON responses, error mapping and bearer auth for plain HTTP routes |
| `accesskey`     | `Verify(...)` — resolve a scoped API key to its owner              |
| `usage`         | free-tier allowance: `Policy`, `Consume`, `Read`                   |
| `outbox`        | `Queue(...)` — transactional mail outbox; `Fields` HTML body       |
| `gen`           | generated protobuf + ConnectRPC code (do not edit)                |

## Typical service bootstrap

```go
func main() {
	ctx := context.Background()
	cfg, _ := config.Load()
	logger := log.New(log.Options{
		Level: cfg.App.LogLevel, Format: cfg.App.LogFormat,
		Service: "identity", Env: cfg.App.Env,
	})
	log.SetDefault(logger)

	shutdown, _ := observability.Setup(ctx, observability.Config{
		Enabled: cfg.OTel.Enabled, Endpoint: cfg.OTel.Endpoint,
		ServiceName: "identity", SampleRatio: cfg.OTel.SampleRatio,
	})
	defer shutdown(ctx)

	pool, _ := postgres.New(ctx, cfg.Postgres.URL)
	defer pool.Close()

	hc := health.New()
	hc.Register("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })

	interceptors, _ := connectutil.Default(logger)

	mux := http.NewServeMux()
	mux.Handle("/healthz", hc.Liveness())
	mux.Handle("/readyz", hc.Readiness())
	// mux.Handle(identityv1connect.NewAuthServiceHandler(handler, connect.WithInterceptors(interceptors...)))

	_ = server.Run(ctx, cfg.HTTP.Addr(), mux, logger)
}
```

## Design rules

- Keep this library small and dependency-light; it is on every service's import path.
- No business logic here — only cross-cutting infrastructure.
- Generated code lives in `gen/` and is owned by `buf`, not humans.
