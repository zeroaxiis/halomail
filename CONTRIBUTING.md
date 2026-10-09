# Contributing to HaloMail

Thanks for your interest in contributing! HaloMail is MIT-licensed and built to
be modular and approachable. This guide covers setup, conventions, and how to
add new functionality.

## Ground rules

- Be respectful — see [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
- Discuss large changes in an issue before opening a big PR.
- Keep PRs focused: one logical change per PR.
- Every change keeps the build green: `task test`, `task proto:lint`, `pnpm lint`.

## Project layout

```
proto/        API contracts — the source of truth. Code is generated from here.
services/     One Go module per service + a shared platform library.
apps/web      Next.js 15 frontend.
apps/docs     Documentation site (in-house, no external framework).
packages/     Generated TypeScript SDK.
deploy/       Dockerfiles + infra.
```

Read [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the why.

## Development setup

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for the full walkthrough. TL;DR:

```bash
cp .env.example .env
task up          # postgres, redis, mailpit, jaeger
task proto       # generate code
task bootstrap   # install deps
task migrate     # db schema
```

## Proto-first workflow

The API contract lives in `proto/`. **Never hand-edit generated code** under
`services/shared/gen/` or `packages/sdk-js/src/gen/`.

1. Edit the relevant `.proto` file.
2. `task proto:lint` then `task proto` to regenerate.
3. Implement the new RPC in the owning service.
4. Wire it into the gateway and (if user-facing) the web app + SDK.

## Coding standards

**Go**

- `gofmt` / `goimports` clean; `task api:lint` (vet + golangci-lint) passes.
- Clean architecture per service: `internal/domain` (pure), `internal/app`
  (use cases), `internal/adapters` (db, rpc, external). Dependencies point
  inward; the domain imports nothing from `adapters`.
- Return domain errors from `services/shared/errs`; let the RPC layer map them.
- Table-driven tests; name them `TestXxx`. Run `task test` (`-race`).

**TypeScript**

- Biome for format + lint (`pnpm lint`).
- Prefer server components in the Next.js app; keep client components small.

## Commits & branches

- Branch names: `feat/…`, `fix/…`, `docs/…`, `chore/…`.
- [Conventional Commits](https://www.conventionalcommits.org/): `feat(scheduling): add date overrides`.

## Adding a new service

1. `mkdir -p services/<name>/{cmd/server,internal/{domain,app,adapters},migrations}`.
2. `cd services/<name> && go mod init github.com/aashishrajdev/halomail/services/<name>`.
3. Add it to `go.work`.
4. Define its contract in `proto/halomail/<name>/v1/<name>.proto`; run `task proto`.
5. Copy the bootstrap pattern from an existing `cmd/server/main.go` (config, log,
   otel, health, ConnectRPC handlers via `services/shared`).
6. Register its handler in the gateway and the monolith binary.
7. Add `services/<name>/README.md`.

## Pull request checklist

- [ ] Tests added/updated and passing (`task test`).
- [ ] `task proto:lint` and `pnpm lint` pass.
- [ ] Docs/READMEs updated for user-facing or structural changes.
- [ ] No secrets committed; `.env` stays local.

## License

By contributing, you agree your contributions are licensed under the
[MIT License](LICENSE).
