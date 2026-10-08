# Sahayak · सहायक

India-wide welfare-scheme discovery and application-assistance platform. Voice-first,
multilingual (Hindi, English, regional languages), mobile-first PWA.

This repository implements the architecture described in `sahayak-implementation-plan.md`
as a Go modular monolith plus a SvelteKit PWA.

## Architecture at a glance

| Layer | Technology |
|-------|------------|
| Backend | Go 1.26 modular monolith, `chi` router |
| Database | PostgreSQL 16, `pgx` v5 |
| Migrations | Embedded SQL migrations (`cmd/migrate`) |
| Job queue | River (PostgreSQL-backed) |
| Rules engine | `expr-lang/expr` (deterministic, 3-state) |
| Voice | Sarvam AI (primary) / Bhashini (alternative), provider-abstracted |
| LLM | Gemini (primary) / OpenAI-compatible, fallback deterministic provider |
| PDF | `maroto` v2 |
| Frontend | SvelteKit 2 + Svelte 5, `@vite-pwa/sveltekit` |

**Core principle:** the LLM extracts facts and phrases responses; it never decides
eligibility. Eligibility is computed exclusively by the deterministic rules engine.

## Repository layout

```
cmd/server        API + worker entrypoint
cmd/migrate       up | down | seed
cmd/seed          migrate + seed (+ --admin-phone)
internal/
  auth location profile scheme source eligibility document
  conversation application notification helper audit maintenance
  rules           expr-lang wrapper (validation, field extraction, caching)
  platform/       llm | voice | sms | whatsapp | translate providers
  middleware      auth, rate limiting, logging, CORS, request id
  server          router + dependency wiring
  jobs            River job types, workers, client
migrations/       versioned SQL + seed data
frontend/         SvelteKit PWA
```

## Quick start (local)

Requirements: Go 1.26+, PostgreSQL 16, Node 20+ (pnpm or npm).

```bash
cp .env.example .env

# 1. Start PostgreSQL (or use your own)
docker compose up -d db

# 2. Apply migrations (includes River's schema) and seed reference data
make migrate
make seed
# optional: create an admin
make seed-admin PHONE=+919876543210

# 3. Run the API (server + River worker)
make run

# 4. Run the frontend
cd frontend && npm install && npm run dev
```

API: http://localhost:8080 · Frontend dev: http://localhost:5173 (proxies `/api` to the API).

Without SMS credentials, OTPs are printed to the server log (console SMS provider).
Without a voice/LLM key, the app degrades to text/form input and deterministic prompts.

## Development environment (Nix + direnv)

`flake.nix` pins the full toolchain (Go 1.26, Node 22 + pnpm, PostgreSQL 16,
golangci-lint, ffmpeg, jq, openssl, python3). `.envrc` loads the flake and your
`.env`.

```bash
direnv allow            # or: nix develop
pg-start && pg-create   # local PostgreSQL on port 55432
make migrate && make seed && make run
```

The dev shell exports `DATABASE_URL`-compatible Postgres settings (`PGPORT=55432`)
and puts `bin/` on `PATH`. `pg-start`, `pg-stop`, `pg-create` and `pg-init` are
real scripts (not shell functions), so they work in bash, zsh and fish alike.
Postgres data lives outside the repo (default `~/.local/share/sahayak/pgdata`, or
`$XDG_DATA_HOME`), because a live socket inside the flake source directory breaks
`nix develop`.

