# apps/docs

Documentation for HaloMail.

## Now: raw static page

Unified, self-contained, dependency-free documentation:

- [`index.html`](index.html) — the complete developer guide: quickstart, embed widgets, framework integration, ConnectRPC API reference, and signed webhooks.

[`index.html`](index.html) is a **self-contained, dependency-free** docs page —
real content, premium minimal styling, light/dark, Geist. Open it directly:

```bash
# just open the file
start index.html            # Windows
open index.html             # macOS

# or serve it
npx serve .                 # → http://localhost:3000
python -m http.server 3001  # → http://localhost:3001
```

No build step, no framework — intentionally "raw and real" for now.

## Roadmap: HaloMail's own docs engine

This page is the seed of a **home-grown documentation product** — a HaloMail SaaS
surface that works like a docs engine (search, versioning, navigation, API
reference generated from the proto/OpenAPI), built in-house rather than adopting
an external framework. The raw page stays the source of truth for content until
that lands.
