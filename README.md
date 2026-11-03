<div align="center">

# HaloMail

**API-first platform for meeting scheduling + portfolio contact forms.**

Minimal. Premium. Developer-focused.

[Architecture](docs/ARCHITECTURE.md) · [Development](docs/DEVELOPMENT.md) · [Deployment](docs/DEPLOYMENT.md) · [Contributing](CONTRIBUTING.md) · [Backend services](services/README.md)

![status](https://img.shields.io/badge/status-alpha-orange) ![go](https://img.shields.io/badge/Go-1.24-00ADD8) ![next](https://img.shields.io/badge/Next.js-15-black) ![contributions](https://img.shields.io/badge/contributions-welcome-brightgreen) ![license](https://img.shields.io/badge/license-Proprietary%20%7C%20Open%20Contributions-blue)

</div>

---

> [!NOTE]
> **Source-Available & Open for Community Contributions:**  
> HaloMail is proprietary software developed and owned by **ZeroAxiis** (`zeroaxiis`). While the core codebase, commercial distribution, and branding are private and proprietary, we actively welcome and encourage open-source community contributions! Developers are warmly invited to report bugs, submit pull requests, improve documentation, and collaborate on features. See [Contributing](#contributing) and [CONTRIBUTING.md](CONTRIBUTING.md).

---

## What is HaloMail?

HaloMail gives every user two things, behind one clean API:

1. **Forms** — generate one public submission key, embed it in website forms,
   and receive submitted fields at your account email.
2. **Meetings** — generate a separate booking key, connect Google Calendar,
   configure meeting types and availability, and share a booking button.
   Calendar-confirmed bookings receive Google Meet links and email notifications.

Keys are shown once, masked thereafter, and must be deleted before replacement.
See [Forms and Meetings setup, behavior and limitations](docs/FORMS_AND_MEETINGS.md)
before running or deploying the new flow.

Everything is driven by a typed **ConnectRPC** API (gRPC + gRPC-Web + JSON/REST
from one definition), so the dashboard, the SDK, and your own integrations all
speak the same contract. It ships with API keys, webhooks, audit logs, an email
theme designer, and first-class OpenTelemetry observability.

> **Names:** the project/brand is **HaloMail**; the repository is `halomail`.

## Feature tour

| Area               | What you get                                                             |
| ------------------ | ----------------------------------------------------------------------- |
| **Scheduling**     | Key-based booking pages, availability rules + date overrides, Google Calendar/Meet integration, queued confirmations and cancellation links |
| **Contact forms**  | HTML/JSON submission endpoint, embeddable widget, honeypot/heuristic spam protection, quotas, storage and queued email forwarding |
| **Email designer** | Built-in themes — Minimal, Apple, Notion, Glass, Terminal — plus custom HTML and live preview |
| **Developer**      | API keys, signed webhooks, generated TypeScript SDK, OpenAPI docs, audit logs |
| **Operations**     | OpenTelemetry traces, structured JSON logs, liveness/readiness probes |

## How it's built

A **monorepo** of independent, modular Go microservices that share a common
platform library and speak ConnectRPC, plus a Next.js 15 frontend and an in-house
docs site. See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

```
proto/        API contracts (source of truth)
services/     Go services: gateway, identity, scheduling, contact, template, notification + shared lib
apps/web/     Next.js 15 dashboard, marketing, public booking + contact pages
apps/docs/    In-house documentation site (raw now, docs engine later)
packages/     generated TypeScript SDK
deploy/       Dockerfiles, infra configs
```

| Backend | Protocol | Data | Email | Frontend | Docs | Observability |
| ------- | -------- | ---- | ----- | -------- | ---- | ------------- |
| Go 1.24 | ConnectRPC | PostgreSQL + Redis | Resend | Next.js 15 · Tailwind · shadcn/ui · Geist | In-house docs | OpenTelemetry · slog · health checks |

## Requirements

| Tool       | Version | Needed for                          | Install                                  |
| ---------- | ------- | ----------------------------------- | ---------------------------------------- |
| **Go**     | 1.24+   | backend services                    | `winget install GoLang.Go`               |
| **Node**   | 20+     | frontend, docs, SDK                 | https://nodejs.org                        |
| **pnpm**   | 9+      | JS workspace package manager        | `npm i -g pnpm`                          |
| **Docker** | latest  | Postgres, Redis, Mailpit, Jaeger    | https://docs.docker.com/get-docker        |
| **buf**    | latest  | generate code from proto            | https://buf.build/docs/installation       |
| **goose**  | latest  | database migrations                 | `go install github.com/pressly/goose/v3/cmd/goose@latest` |
| **task**   | latest  | task runner (optional, convenient)  | `go install github.com/go-task/task/v3/cmd/task@latest`   |

## Quickstart (local)

```bash
git clone https://github.com/zeroaxiis/halomail
cd halomail
cp .env.example .env            # fill in secrets (works with defaults locally)

# 1. Infra: Postgres, Redis, Mailpit (email), Jaeger (traces)
task up                         # or: docker compose up -d

# 2. Generate Go + TS code from the proto contracts
task proto                      # or: buf generate

# 3. Install deps + run database migrations
task bootstrap
task migrate

# 4. Run the backend (all services in one process — see services/README.md)
task api:run

# 5. Run the dashboard
task web                        # http://localhost:3000
```

Local UIs while developing:

| URL                      | What                |
| ------------------------ | ------------------- |
| http://localhost:3000    | Web dashboard       |
| http://localhost:8080    | API (ConnectRPC)    |
| http://localhost:8025    | Mailpit (caught email) |
| http://localhost:16686   | Jaeger (traces)     |

Full per-service instructions: **[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)**.

## Documentation map

| Doc | Purpose |
| --- | ------- |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)   | System design, service boundaries, data flow, deploy modes |
| [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)     | Requirements + how to run the app and every service locally |
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)       | Production deployment |
| [docs/BACKUP_AND_RECOVERY.md](docs/BACKUP_AND_RECOVERY.md) | Hourly Neon backups in R2, latest two batches, and recovery into a replacement database |
| [services/README.md](services/README.md)       | Backend overview; each service has its own README |
| [proto/README.md](proto/README.md)             | API contracts and code generation |
| [CONTRIBUTING.md](CONTRIBUTING.md)             | How to contribute, coding standards, adding a service |

## Contributing

HaloMail is built in the open for community participation and contributions. Even though HaloMail is proprietary to **ZeroAxiis**, we believe in open collaboration with developers worldwide.

You can contribute by:
- Reporting bugs, suggestions, and submitting feedback via Issues.
- Submitting Pull Requests for bug fixes, performance improvements, and new capabilities.
- Expanding documentation, guides, and example implementations.

Start with **[CONTRIBUTING.md](CONTRIBUTING.md)** and check out open issues marked `good first issue`. All contributors must follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## License & Ownership

Copyright © 2026 **ZeroAxiis**. All rights reserved.

HaloMail is proprietary software. The source code is made accessible for inspection and community contributions. Unauthorized commercial redistribution, resale, sublicensing, or proprietary forks without explicit written permission from ZeroAxiis is strictly prohibited.

For inquiries, contact `contact@zeroaxiis.tech` or `legal@zeroaxiis.tech`.
