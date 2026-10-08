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

## Configuration

All configuration is environment-based (see `.env.example`). Key variables:

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET`, `PHONE_HASH_PEPPER` | Required in production |
| `SERVER_MODE` | `all` (api+worker), `api`, `worker` |
| `VOICE_PROVIDER`, `SARVAM_API_KEY` | Voice backend |
| `LLM_PROVIDER`, `GEMINI_API_KEY` / `OPENAI_API_KEY` | LLM backend (failover chain) |
| `SMS_PROVIDER`, `MSG91_AUTH_KEY` | SMS backend |
| `WHATSAPP_API_KEY` | Gupshup WhatsApp backend |
| `DEMO_OTP` | Fixed login code for demos (never set in production) |
| `ADMIN_PHONE` | Phone promoted to admin during seed |

## Hackathon / demo mode

Set `DEMO_OTP` to any 6-digit code (e.g. `123456`). Then:

- `POST /auth/otp/request` returns the code in `demo_otp` and skips SMS, so judges
  can log in with no SMS provider. The login screen shows and pre-fills it.
- Per-phone OTP limits and the anonymous IP rate limits are relaxed, so a shared
  venue network cannot lock anyone out.
- `GET /api/v1/config` reports `demo_mode`, `voice_enabled` and `llm_enabled`, and
  the UI hides the microphone when no voice provider is configured.

The frontend also hides the mic unless a voice key is present, so a text-only demo
never shows a broken button.

## Provider reliability

`llmProvider` builds a **failover chain**: the configured LLM first, the other
configured LLM second, then a deterministic non-networked fallback. Voice STT/TTS
do the same across Sarvam and Bhashini. A single provider outage degrades quality
instead of breaking a live demo.

For a reliable paid setup at low cost:

- **LLM**: OpenAI-compatible providers work out of the box via `OPENAI_BASE_URL`.
  Groq (`https://api.groq.com/openai/v1`, model `llama-3.3-70b-versatile`) and
  OpenAI `gpt-4o-mini` are cheap and fast. Set both `GEMINI_API_KEY` and
  `OPENAI_API_KEY` to get automatic failover. On Gemini's free tier prefer
  `gemini-3.5-flash-lite` (RPM 15 / RPD 500) over the standard Flash models
  (RPD 20).
- **Voice**: Sarvam models are `saaras:v4` (STT) and `bulbul:v3` (TTS); the TTS
  speaker must be one valid for `bulbul:v3` (e.g. `priya`). Sarvam requires
  credits — a demo needs only a few hundred rupees. Bhashini is free but
  approval-gated. Configure both for failover.
- **SMS**: only needed for real OTP delivery; `DEMO_OTP` removes that dependency
  for demos.

LLM-extracted facts are normalized into the controlled vocabulary and anything
unmappable is dropped (`internal/profile.SanitizeExtractedFacts`), so a model
returning `"widow"` or `"farming"` still produces correct `widowed`/`farmer`
facts instead of failing the request.

## Data model highlights

- **Phone numbers** are never stored in plaintext. A SHA-256 HMAC (peppered) is used for
  lookup; last 4 digits for display; and an AES-256-GCM ciphertext (`phone_encrypted`)
  is retained solely to dispatch notifications.
- **Schemes** move through `draft → verified → published → archived`. Only `published`
  schemes participate in eligibility. Discovered candidates live in `candidate_schemes`
  and never reach citizens.
- **Eligibility rules** are `expr` expressions referencing a controlled fact vocabulary.
  Rules are validated (syntax + known fields) on save.
