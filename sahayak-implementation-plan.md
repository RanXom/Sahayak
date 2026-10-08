# Sahayak — Complete Implementation Plan

## Goal Description

Sahayak is an India-wide welfare-scheme discovery and application-assistance platform. It helps Indian citizens discover government welfare schemes they are eligible for and assists them through the application process. The experience is voice-first, multilingual (Hindi, English, regional languages), and designed primarily for mobile users.

This document is the complete architecture and implementation plan. It is detailed enough that a coding agent can take it and implement every component without guessing.

---

## Decisions Recorded
> [!NOTE]
> **Voice Provider**: Sarvam AI selected as initial implementation for superior low-latency, code-mixed Indian language support. Architecture remains provider-agnostic.
> **Scheme Discovery**: MVP will rely on manual curation of high-impact schemes using the source registry architecture. Broad automated scraping is deferred to Phase 2.
> **LLM Provider**: Gemini Flash selected as initial provider. Strictly restricted to fact extraction and NLG; forbidden from making eligibility decisions.
> **Job Queue**: River (PostgreSQL-backed) confirmed.

---

# Table of Contents

1. [Product Architecture](#1-product-architecture)
2. [System Architecture](#2-system-architecture)
3. [Backend Architecture](#3-backend-architecture)
4. [Frontend Architecture](#4-frontend-architecture)
5. [Database Schema](#5-database-schema)
6. [API Design](#6-api-design)
7. [Authentication and Authorization](#7-authentication-and-authorization)
8. [Voice/STT/TTS Architecture](#8-voicestttts-architecture)
9. [LLM Integration](#9-llm-integration)
10. [Deterministic Eligibility Engine](#10-deterministic-eligibility-engine)
11. [Scheme Data Model](#11-scheme-data-model)
12. [Government Source Registry](#12-government-source-registry)
13. [Source Discovery and Controlled Fetching](#13-source-discovery-and-controlled-fetching)
14. [Verification Workflow](#14-verification-workflow)
15. [Unverified Scheme Handling](#15-unverified-scheme-handling)
16. [India-Wide Location Model](#16-india-wide-location-model)
17. [Profile/Intake Flow](#17-profileintake-flow)
18. [Missing Document System](#18-missing-document-system)
19. [Document Acquisition Guidance](#19-document-acquisition-guidance)
20. [Application Form Generation](#20-application-form-generation)
21. [PDF Generation](#21-pdf-generation)
22. [Application Status Tracking](#22-application-status-tracking)
23. [WhatsApp/SMS Notifications](#23-whatsappsms-notifications)
24. [ASHA/CSC Helper Mode](#24-ashacsc-helper-mode)
25. [Privacy and Security](#25-privacy-and-security)
26. [Rate Limiting and Abuse Protection](#26-rate-limiting-and-abuse-protection)
27. [Audit Logging](#27-audit-logging)
28. [Observability](#28-observability)
29. [Database Backups](#29-database-backups)
30. [Docker Setup](#30-docker-setup)
31. [Coolify Deployment](#31-coolify-deployment)
32. [CI/CD](#32-cicd)
33. [Testing Strategy](#33-testing-strategy)
34. [Failure and Fallback Behaviour](#34-failure-and-fallback-behaviour)
35. [Data Freshness and Verification](#35-data-freshness-and-verification)
36. [Multilingual Architecture](#36-multilingual-architecture)
37. [Frontend/Backend Contract](#37-frontendbackend-contract)
38. [PWA Requirements](#38-pwa-requirements)
39. [Repository Structure](#39-repository-structure)
40. [API Endpoint Specification](#40-api-endpoint-specification)
41. [Core User Flow Diagrams](#41-core-user-flow-diagrams)
42. [Eligibility Engine Design](#42-eligibility-engine-design)
43. [Scheme Source/Discovery Architecture](#43-scheme-sourcediscovery-architecture)
44. [Voice/LLM Architecture](#44-voicellm-architecture)
45. [Notification Architecture](#45-notification-architecture)
46. [Helper Mode Architecture](#46-helper-mode-architecture)
47. [Deployment Architecture](#47-deployment-architecture)
48. [Development Milestones](#48-development-milestones)
49. [Technical Risks and Mitigations](#49-technical-risks-and-mitigations)
50. [External Services and Alternatives](#50-external-services-and-alternatives)

---

# 1. Product Architecture

## High-Level Architecture

```mermaid
flowchart TD
    subgraph Client["Mobile PWA (SvelteKit)"]
        UI["Voice-First UI"]
        MIC["Microphone / MediaRecorder"]
        SPK["Audio Playback"]
        I18N["i18n Layer"]
    end

    subgraph Backend["Go Modular Monolith"]
        API["REST API (chi router)"]
        AUTH["Auth Module"]
        CONV["Conversation / Intake Module"]
        PROF["Profile Module"]
        SCHEME["Scheme Module"]
        ELIG["Eligibility Engine"]
        DOC["Document Module"]
        APP["Application Module"]
        NOTIF["Notification Module"]
        HELPER["Helper Mode Module"]
        LOC["Location Module"]
        SOURCE["Source Registry"]
        AUDIT["Audit Module"]
    end

    subgraph External["External Services"]
        BHASHINI["Bhashini ASR/TTS/NMT"]
        LLM["LLM (Gemini/OpenAI/Claude)"]
        WHATSAPP["WhatsApp BSP"]
        SMS["SMS Provider"]
    end

    subgraph Data["Data Layer"]
        PG["PostgreSQL"]
        RIVER["River Job Queue (in PG)"]
    end

    UI --> API
    MIC --> API
    API --> AUTH
    API --> CONV
    API --> PROF
    API --> SCHEME
    API --> ELIG
    API --> DOC
    API --> APP
    API --> NOTIF
    API --> HELPER
    API --> LOC
    CONV --> BHASHINI
    CONV --> LLM
    ELIG --> SCHEME
    NOTIF --> WHATSAPP
    NOTIF --> SMS
    NOTIF --> RIVER
    AUTH --> PG
    PROF --> PG
    SCHEME --> PG
    APP --> PG
    SOURCE --> PG
```

## Key Architectural Principles

1. **Modular monolith**: Single Go binary with clear internal module boundaries. No microservices.
2. **LLM is a tool, not the authority**: The LLM extracts facts and generates explanations. Eligibility is determined by deterministic rules.
3. **Backend is source of truth**: Frontend contains zero eligibility logic or scheme rules.
4. **Provider abstraction**: STT, TTS, LLM, SMS, WhatsApp — all behind interfaces for swappability.
5. **PostgreSQL for everything**: Primary store, job queue (River), audit log. No Redis.
6. **Verified vs. unverified**: Strict separation. Only verified schemes participate in eligibility evaluation.

---

# 2. System Architecture

## Technology Stack

| Layer | Technology | Justification |
|-------|-----------|---------------|
| **Backend language** | Go 1.23+ | Strong concurrency, single binary deployment, excellent for APIs |
| **HTTP router** | chi | Idiomatic, `net/http` compatible, excellent route grouping |
| **Database** | PostgreSQL 16+ | JSONB for flexible profile data, full-text search, proven reliability |
| **DB driver** | pgx v5 | Fastest Go PG driver, native type support |
| **Query gen** | sqlc | Type-safe Go from SQL, compile-time verification |
| **Migrations** | golang-migrate | Version-ordered SQL migrations, CLI + embeddable |
| **Job queue** | River | PostgreSQL-backed, transactional job insertion, no Redis needed |
| **Auth** | golang-jwt/jwt v5 | JWT tokens for stateless API auth |
| **Rules engine** | expr-lang/expr | Expression-based rules stored in DB, safe sandboxed evaluation |
| **PDF** | maroto v2 + wkhtmltopdf | Structured PDF generation with Hindi/Unicode font support |
| **LLM client** | go-openai + direct HTTP | OpenAI-compatible APIs + Bhashini direct calls |
| **Frontend** | SvelteKit | Smallest bundles, excellent PWA support, fast on mobile |
| **PWA** | @vite-pwa/sveltekit | Service worker, manifest, installability |
| **i18n** | paraglide-js | Compile-time i18n, type-safe, tree-shakeable |
| **Containerization** | Docker + Docker Compose | Multi-stage Go builds, ~15MB images |
| **Deployment** | Coolify on OCI VPS | Simple PaaS with Docker support, HTTPS, health checks |

## Module Dependency Graph

```mermaid
flowchart LR
    AUTH["auth"]
    LOC["location"]
    PROF["profile"]
    CONV["conversation"]
    SCHEME["scheme"]
    SOURCE["source"]
    ELIG["eligibility"]
    DOC["document"]
    APP["application"]
    NOTIF["notification"]
    HELPER["helper"]
    AUDIT["audit"]

    CONV --> AUTH
    CONV --> PROF
    CONV --> LOC
    PROF --> AUTH
    PROF --> LOC
    ELIG --> SCHEME
    ELIG --> PROF
    DOC --> SCHEME
    DOC --> PROF
    APP --> SCHEME
    APP --> PROF
    APP --> DOC
    NOTIF --> APP
    NOTIF --> AUTH
    HELPER --> AUTH
    HELPER --> PROF
    HELPER --> CONV
    SOURCE --> SCHEME
    AUDIT --> AUTH
```

---

# 3. Backend Architecture

## Module Design

The Go backend is structured as a modular monolith. Each module:
- Has its own package under `internal/`
- Owns its database tables (migrations are ordered globally but conceptually belong to modules)
- Exposes a service interface consumed by HTTP handlers
- Does not import other module's internal types directly — uses shared domain types or interfaces

### Module Summary

| Module | Package | Owns | Depends On |
|--------|---------|------|------------|
| `auth` | `internal/auth` | users, sessions, OTPs, roles | — |
| `location` | `internal/location` | states, districts, coverage | — |
| `profile` | `internal/profile` | citizen profiles, profile facts | auth, location |
| `conversation` | `internal/conversation` | conversation sessions, messages | auth, profile, location |
| `scheme` | `internal/scheme` | schemes, eligibility rules, documents | — |
| `source` | `internal/source` | source registry, fetch jobs, candidate schemes | scheme |
| `eligibility` | `internal/eligibility` | evaluation results | scheme, profile |
| `document` | `internal/document` | document status, acquisition guidance | scheme, profile |
| `application` | `internal/application` | applications, form mappings, PDFs | scheme, profile, document |
| `notification` | `internal/notification` | notification jobs, templates, delivery status | auth, application |
| `helper` | `internal/helper` | helper sessions, citizen-worker links, consent records | auth, profile, conversation |
| `audit` | `internal/audit` | audit log entries | auth |

### Shared Packages

| Package | Contents |
|---------|----------|
| `internal/platform/llm` | LLM provider interface + implementations |
| `internal/platform/voice` | STT/TTS provider interface + implementations |
| `internal/platform/sms` | SMS provider interface + implementations |
| `internal/platform/whatsapp` | WhatsApp provider interface + implementations |
| `internal/platform/translate` | Translation provider interface + implementations |
| `internal/domain` | Shared domain types (ProfileFacts, SchemeID, etc.) |
| `internal/middleware` | HTTP middleware (auth, rate limiting, logging, CORS) |
| `internal/config` | Configuration loading from env vars |
| `internal/server` | HTTP server setup, graceful shutdown |

### Application Entrypoint

```
cmd/
├── server/        # Main API server
│   └── main.go
├── migrate/       # Migration runner
│   └── main.go
└── worker/        # Background job worker (can be same binary with flag)
    └── main.go
```

The server and worker can be the same binary (flag `--mode=server|worker|all`). For simplicity, default mode is `all` (both API server and River worker in one process).

---

# 4. Frontend Architecture

## Framework: SvelteKit

### Why SvelteKit

- **Smallest bundle size**: Critical for Indian mobile users (2G/3G connections)
- **No virtual DOM**: Direct DOM manipulation = faster on low-end devices
- **Built-in SSR**: Fast initial load, SEO-friendly
- **Native service worker support**: Clean PWA implementation
- **Simple component model**: Easy to maintain

### Key Frontend Responsibilities

1. **Voice capture**: Use `MediaRecorder` API to capture audio, send to backend
2. **Audio playback**: Play TTS audio responses from backend
3. **Conversation UI**: Chat-like interface with voice button
4. **Profile display**: Show extracted profile facts for user review/correction
5. **Scheme results**: Display eligible schemes with details
6. **Document checklist**: Show ✓/❌ document status
7. **Application form**: Render auto-filled form for review
8. **i18n**: All UI strings in user's language
9. **PWA**: Installable, offline shell (not offline scheme discovery)

### What the Frontend Does NOT Do

- No eligibility logic
- No scheme rule evaluation
- No API keys or secrets
- No direct calls to LLM/Bhashini
- No data persistence (all through backend APIs)

### Frontend Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── components/
│   │   │   ├── voice/
│   │   │   │   ├── VoiceButton.svelte      # Mic record/stop
│   │   │   │   ├── AudioPlayer.svelte       # TTS playback
│   │   │   │   └── WaveformVisualizer.svelte
│   │   │   ├── chat/
│   │   │   │   ├── ChatWindow.svelte
│   │   │   │   ├── MessageBubble.svelte
│   │   │   │   └── TypingIndicator.svelte
│   │   │   ├── profile/
│   │   │   │   ├── ProfileCard.svelte
│   │   │   │   ├── FactEditor.svelte
│   │   │   │   └── ProfileReview.svelte
│   │   │   ├── scheme/
│   │   │   │   ├── SchemeCard.svelte
│   │   │   │   ├── SchemeDetail.svelte
│   │   │   │   ├── SchemeList.svelte
│   │   │   │   └── DocumentChecklist.svelte
│   │   │   ├── application/
│   │   │   │   ├── ApplicationForm.svelte
│   │   │   │   └── ApplicationStatus.svelte
│   │   │   └── common/
│   │   │       ├── LanguageSelector.svelte
│   │   │       ├── LocationPicker.svelte
│   │   │       └── LoadingSpinner.svelte
│   │   ├── api/
│   │   │   └── client.ts                   # Typed API client
│   │   ├── stores/
│   │   │   ├── auth.ts
│   │   │   ├── conversation.ts
│   │   │   ├── profile.ts
│   │   │   └── language.ts
│   │   ├── i18n/
│   │   │   ├── messages/
│   │   │   │   ├── en.json
│   │   │   │   ├── hi.json
│   │   │   │   └── ta.json
│   │   │   └── index.ts
│   │   └── utils/
│   │       ├── audio.ts                    # MediaRecorder helpers
│   │       └── format.ts
│   ├── routes/
│   │   ├── +layout.svelte
│   │   ├── +page.svelte                    # Landing / conversation
│   │   ├── profile/
│   │   │   └── +page.svelte
│   │   ├── schemes/
│   │   │   ├── +page.svelte                # Scheme results
│   │   │   └── [id]/
│   │   │       └── +page.svelte            # Scheme detail
│   │   ├── apply/
│   │   │   └── [schemeId]/
│   │   │       └── +page.svelte
│   │   ├── status/
│   │   │   └── +page.svelte
│   │   └── helper/
│   │       ├── +layout.svelte
│   │       ├── +page.svelte                # Helper dashboard
│   │       └── session/
│   │           └── [id]/
│   │               └── +page.svelte
│   ├── app.html
│   └── service-worker.ts
├── static/
│   ├── manifest.json
│   └── icons/
├── svelte.config.js
├── vite.config.ts
├── package.json
└── tsconfig.json
```

---

# 5. Database Schema

## Core Schema

```sql
-- ================================================
-- AUTH MODULE
-- ================================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_hash      TEXT UNIQUE NOT NULL,           -- SHA-256 hash of phone number
    phone_last4     TEXT NOT NULL,                   -- Last 4 digits for display
    role            TEXT NOT NULL DEFAULT 'citizen', -- citizen, helper, admin
    preferred_lang  TEXT NOT NULL DEFAULT 'hi',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

CREATE TABLE otp_challenges (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_hash      TEXT NOT NULL,
    otp_hash        TEXT NOT NULL,               -- bcrypt hash of OTP
    purpose         TEXT NOT NULL DEFAULT 'login', -- login, verify
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 3,
    expires_at      TIMESTAMPTZ NOT NULL,
    consumed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    token_hash      TEXT UNIQUE NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ================================================
-- LOCATION MODULE
-- ================================================

CREATE TABLE states (
    id              INT PRIMARY KEY,              -- LGD state code
    code            TEXT UNIQUE NOT NULL,          -- ISO-like: "RJ", "TN", etc.
    name_en         TEXT NOT NULL,
    name_hi         TEXT NOT NULL,
    name_local      TEXT,                          -- Name in state's own script
    is_ut           BOOLEAN NOT NULL DEFAULT FALSE,
    active          BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE districts (
    id              INT PRIMARY KEY,              -- LGD district code
    state_id        INT NOT NULL REFERENCES states(id),
    name_en         TEXT NOT NULL,
    name_hi         TEXT NOT NULL,
    name_local      TEXT,
    census_code     TEXT,
    active          BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE coverage_status (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state_id        INT NOT NULL REFERENCES states(id),
    level           TEXT NOT NULL DEFAULT 'none', -- full, partial, none
    scheme_count    INT NOT NULL DEFAULT 0,
    last_updated    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes           TEXT,
    UNIQUE(state_id)
);

-- ================================================
-- PROFILE MODULE
-- ================================================

CREATE TABLE profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    state_id        INT REFERENCES states(id),
    district_id     INT REFERENCES districts(id),
    facts           JSONB NOT NULL DEFAULT '{}',
    -- facts schema:
    -- {
    --   "age": 45,
    --   "gender": "female",
    --   "marital_status": "widowed",
    --   "income_annual": 120000,
    --   "occupation": "farmer",
    --   "caste_category": "obc",
    --   "residence_type": "rural",
    --   "land_acres": 2.0,
    --   "ration_card_type": "bpl",
    --   "has_bank_account": true,
    --   "has_aadhaar": true,
    --   "education_level": "illiterate",
    --   "disability_status": "none",
    --   "minority_status": false,
    --   "family_members": [
    --     {"relation": "son", "age": 14, "education": "class_9"}
    --   ]
    -- }
    completed       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profiles_user_id ON profiles(user_id);
CREATE INDEX idx_profiles_facts ON profiles USING GIN (facts);

CREATE TABLE profile_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES profiles(id),
    document_type   TEXT NOT NULL,               -- aadhaar, income_cert, domicile_cert, etc.
    has_document    BOOLEAN NOT NULL DEFAULT FALSE,
    notes           TEXT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(profile_id, document_type)
);

-- ================================================
-- CONVERSATION MODULE
-- ================================================

CREATE TABLE conversations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    profile_id      UUID REFERENCES profiles(id),
    language        TEXT NOT NULL DEFAULT 'hi',
    status          TEXT NOT NULL DEFAULT 'active', -- active, completed, abandoned
    context         JSONB NOT NULL DEFAULT '{}',    -- LLM context / state machine state
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE conversation_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id),
    role            TEXT NOT NULL,                -- user, assistant, system
    content_text    TEXT NOT NULL,                -- text content (original or transcript)
    content_lang    TEXT,                         -- detected language
    facts_extracted JSONB,                        -- structured facts extracted from this message
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- NOTE: Raw audio is NOT stored in DB. If temporarily needed for STT,
-- it is processed in memory and discarded.

CREATE INDEX idx_conv_messages_conv ON conversation_messages(conversation_id, created_at);

-- ================================================
-- SCHEME MODULE
-- ================================================

CREATE TABLE schemes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            TEXT UNIQUE NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft', -- draft, verified, published, archived
    name_en         TEXT NOT NULL,
    name_hi         TEXT NOT NULL,
    description_en  TEXT,
    description_hi  TEXT,
    benefit_type    TEXT NOT NULL,                  -- pension, scholarship, insurance, housing, health, other
    benefit_amount  TEXT,                           -- e.g., "₹1000/month", "up to ₹2,00,000"
    benefit_description_en TEXT,
    benefit_description_hi TEXT,
    scope           TEXT NOT NULL DEFAULT 'central', -- central, state
    state_id        INT REFERENCES states(id),      -- NULL for central schemes
    ministry        TEXT,
    department      TEXT,
    application_url TEXT,                            -- official application URL
    official_url    TEXT,                            -- official scheme info page
    deadline        DATE,                            -- application deadline, if any
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    source_id       UUID REFERENCES sources(id),
    verified_at     TIMESTAMPTZ,
    verified_by     UUID REFERENCES users(id),
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_schemes_status ON schemes(status);
CREATE INDEX idx_schemes_state ON schemes(state_id);
CREATE INDEX idx_schemes_scope ON schemes(scope);

CREATE TABLE scheme_eligibility_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    rule_name       TEXT NOT NULL,                  -- human-readable label
    expression      TEXT NOT NULL,                  -- expr-lang expression
    -- Example expressions:
    -- "age >= 60 && gender == 'female'"
    -- "income_annual <= 200000 && caste_category in ['sc', 'st']"
    -- "state_code == 'RJ' && residence_type == 'rural'"
    -- "has_family_member_student == true && family_student_class >= 9"
    description_en  TEXT,                           -- human-readable description
    description_hi  TEXT,
    is_mandatory    BOOLEAN NOT NULL DEFAULT TRUE,  -- false = optional/bonus criteria
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_eligibility_rules_scheme ON scheme_eligibility_rules(scheme_id);

CREATE TABLE scheme_required_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    document_type   TEXT NOT NULL,                 -- matches profile_documents.document_type
    document_name_en TEXT NOT NULL,
    document_name_hi TEXT NOT NULL,
    is_mandatory    BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INT NOT NULL DEFAULT 0,
    UNIQUE(scheme_id, document_type)
);

CREATE TABLE scheme_application_fields (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    field_name      TEXT NOT NULL,                 -- field identifier
    field_label_en  TEXT NOT NULL,
    field_label_hi  TEXT NOT NULL,
    field_type      TEXT NOT NULL DEFAULT 'text',  -- text, number, date, select, boolean
    profile_mapping TEXT,                          -- JSON path into profile.facts, e.g., "age", "family_members[0].age"
    is_required     BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INT NOT NULL DEFAULT 0,
    options         JSONB,                         -- for select type: [{"value": "m", "label_en": "Male"}]
    UNIQUE(scheme_id, field_name)
);

-- ================================================
-- SOURCE / DISCOVERY MODULE
-- ================================================

CREATE TABLE sources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,                 -- e.g., "MyScheme Central", "Rajasthan Gov Portal"
    source_type     TEXT NOT NULL,                 -- official_api, gov_dataset, gov_portal, gazette, manual
    url             TEXT,
    description     TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    fetch_config    JSONB,                         -- { "method": "scrape", "selector": "...", "schedule": "weekly" }
    last_fetched_at TIMESTAMPTZ,
    last_fetch_status TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE candidate_schemes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id       UUID NOT NULL REFERENCES sources(id),
    external_id     TEXT,                          -- ID from external source, if any
    raw_data        JSONB NOT NULL,                -- raw fetched data
    name            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending', -- pending, reviewing, accepted, rejected
    content_hash    TEXT NOT NULL,                  -- SHA-256 of raw_data for dedup
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at     TIMESTAMPTZ,
    reviewed_by     UUID REFERENCES users(id),
    reviewer_notes  TEXT,
    promoted_to     UUID REFERENCES schemes(id),   -- link to scheme if accepted
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_id, content_hash)
);

-- ================================================
-- DOCUMENT GUIDANCE MODULE
-- ================================================

CREATE TABLE document_types (
    id              TEXT PRIMARY KEY,              -- e.g., "aadhaar", "income_cert", "domicile_cert"
    name_en         TEXT NOT NULL,
    name_hi         TEXT NOT NULL,
    description_en  TEXT NOT NULL,
    description_hi  TEXT NOT NULL,
    category        TEXT NOT NULL DEFAULT 'identity', -- identity, financial, residence, education, other
    issuing_authority_en TEXT,
    issuing_authority_hi TEXT,
    acquisition_url TEXT,                           -- official URL to apply for this document
    acquisition_steps_en TEXT,                      -- verified steps to obtain (markdown)
    acquisition_steps_hi TEXT,
    typical_time    TEXT,                           -- e.g., "7-15 working days"
    typical_fee     TEXT,                           -- e.g., "₹50" or "Free"
    last_verified   TIMESTAMPTZ,
    source_url      TEXT                            -- URL of the source for this guidance
);

-- CSC / office locations for document acquisition
CREATE TABLE service_locations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_type   TEXT NOT NULL,                 -- csc, tehsil, block_office, district_office
    name            TEXT NOT NULL,
    address         TEXT,
    state_id        INT REFERENCES states(id),
    district_id     INT REFERENCES districts(id),
    latitude        DECIMAL(10, 8),
    longitude       DECIMAL(11, 8),
    phone           TEXT,
    services        TEXT[],                        -- array of document_type IDs this location can issue
    active          BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_service_locations_state ON service_locations(state_id);
CREATE INDEX idx_service_locations_district ON service_locations(district_id);

-- ================================================
-- APPLICATION MODULE
-- ================================================

CREATE TABLE applications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    profile_id      UUID NOT NULL REFERENCES profiles(id),
    scheme_id       UUID NOT NULL REFERENCES schemes(id),
    status          TEXT NOT NULL DEFAULT 'draft', -- draft, ready, submitted, tracking, completed, rejected
    form_data       JSONB NOT NULL DEFAULT '{}',  -- filled form fields
    pdf_path        TEXT,                          -- path to generated PDF
    submission_ref  TEXT,                          -- external reference number if submitted
    submitted_at    TIMESTAMPTZ,
    last_status_check TIMESTAMPTZ,
    external_status TEXT,                          -- status from external tracking
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_applications_user ON applications(user_id);
CREATE INDEX idx_applications_scheme ON applications(scheme_id);

-- ================================================
-- NOTIFICATION MODULE
-- ================================================

CREATE TABLE notification_preferences (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    channel         TEXT NOT NULL,                 -- whatsapp, sms
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    phone_hash      TEXT NOT NULL,                 -- destination phone hash
    opted_in_at     TIMESTAMPTZ,
    opted_out_at    TIMESTAMPTZ,
    UNIQUE(user_id, channel)
);

CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    application_id  UUID REFERENCES applications(id),
    channel         TEXT NOT NULL,                 -- whatsapp, sms
    template_key    TEXT NOT NULL,                 -- e.g., "deadline_reminder", "status_update"
    content         TEXT NOT NULL,                 -- rendered message content
    status          TEXT NOT NULL DEFAULT 'pending', -- pending, sent, delivered, failed
    provider_ref    TEXT,                           -- external message ID
    scheduled_for   TIMESTAMPTZ,
    sent_at         TIMESTAMPTZ,
    error           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);

-- ================================================
-- HELPER MODE MODULE
-- ================================================

CREATE TABLE helper_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    helper_type     TEXT NOT NULL,                 -- asha, csc_operator, field_worker
    organization    TEXT,
    id_number       TEXT,                          -- CSC ID or ASHA ID (encrypted)
    verified        BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at     TIMESTAMPTZ,
    verified_by     UUID REFERENCES users(id),
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE helper_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    helper_user_id  UUID NOT NULL REFERENCES users(id),
    citizen_user_id UUID REFERENCES users(id),     -- may be NULL for anonymous
    citizen_phone_hash TEXT,                        -- for linking later
    consent_given   BOOLEAN NOT NULL DEFAULT FALSE,
    consent_text    TEXT,                           -- exact consent statement shown
    consent_at      TIMESTAMPTZ,
    purpose         TEXT NOT NULL,                  -- scheme_discovery, application_help, document_help
    status          TEXT NOT NULL DEFAULT 'active', -- active, completed, abandoned
    conversation_id UUID REFERENCES conversations(id),
    profile_id      UUID REFERENCES profiles(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX idx_helper_sessions_helper ON helper_sessions(helper_user_id);

-- ================================================
-- AUDIT MODULE
-- ================================================

CREATE TABLE audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id        UUID REFERENCES users(id),     -- NULL for system actions
    actor_role      TEXT,
    action          TEXT NOT NULL,                  -- e.g., "profile.view", "scheme.publish", "helper.start_session"
    resource_type   TEXT NOT NULL,                  -- e.g., "profile", "scheme", "application"
    resource_id     TEXT NOT NULL,
    details         JSONB,                          -- additional context (non-sensitive)
    ip_address      TEXT,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_actor ON audit_log(actor_id);
CREATE INDEX idx_audit_log_action ON audit_log(action);
CREATE INDEX idx_audit_log_created ON audit_log(created_at);

-- Partitioning for audit_log (monthly partitions)
-- In production, convert audit_log to a partitioned table:
-- CREATE TABLE audit_log (...) PARTITION BY RANGE (created_at);
```

---

# 6. API Design

## API Conventions

- Base path: `/api/v1`
- JSON request/response bodies
- Authentication: `Authorization: Bearer <JWT>` header
- Errors: `{ "error": { "code": "INVALID_INPUT", "message": "..." } }`
- Pagination: `?page=1&per_page=20` → response includes `{ "pagination": { "page": 1, "per_page": 20, "total": 100 } }`
- Rate limiting: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers
- Request IDs: `X-Request-Id` header on all responses
- Audio uploads: `multipart/form-data` for voice input
- Language preference: `Accept-Language: hi` header

---

# 7. Authentication and Authorization

## Auth Flow for Citizens

1. User enters phone number
2. Backend generates 6-digit OTP, stores bcrypt hash with 5-minute expiry
3. OTP sent via SMS (MSG91/Kaleyra)
4. User enters OTP
5. Backend verifies OTP hash, creates user if new
6. Backend issues JWT access token (15 min) + refresh token (30 days)
7. Refresh token stored hashed in `refresh_tokens` table

## Auth Flow for Helpers

1. Same phone OTP flow
2. Helper must also have a verified `helper_profiles` record
3. Helper tokens include `role: "helper"` claim
4. Admin manually verifies helper profiles (or via bulk import)

## Auth Flow for Admins

1. Email + password login (bcrypt-hashed passwords)
2. TOTP 2FA required
3. Admin tokens include `role: "admin"` claim

## JWT Structure

```json
{
  "sub": "<user_id>",
  "role": "citizen|helper|admin",
  "lang": "hi",
  "iat": 1234567890,
  "exp": 1234568790
}
```

## Authorization Middleware

```go
// Middleware chain per route group
r.Route("/api/v1", func(r chi.Router) {
    r.Use(middleware.RateLimit)
    r.Use(middleware.RequestID)
    r.Use(middleware.Logger)

    // Public routes
    r.Group(func(r chi.Router) {
        r.Post("/auth/otp/request", authHandler.RequestOTP)
        r.Post("/auth/otp/verify", authHandler.VerifyOTP)
        r.Get("/locations/states", locationHandler.ListStates)
    })

    // Citizen routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.Authenticate)
        r.Use(middleware.RequireRole("citizen", "helper", "admin"))
        // ... profile, conversation, schemes, applications
    })

    // Helper routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.Authenticate)
        r.Use(middleware.RequireRole("helper", "admin"))
        r.Use(middleware.RequireVerifiedHelper)
        // ... helper session management
    })

    // Admin routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.Authenticate)
        r.Use(middleware.RequireRole("admin"))
        // ... scheme management, source management, user management
    })
})
```

## Security Details

- **Phone numbers**: Never stored in plaintext. SHA-256 hashed with a server-side pepper for lookup. Last 4 digits stored separately for display.
- **OTPs**: bcrypt-hashed before storage. Max 3 attempts per challenge. Rate limited to 3 requests per phone per 10 minutes.
- **JWTs**: Signed with HMAC-SHA256. Short-lived (15 min). Refresh tokens are opaque UUIDs stored hashed.
- **HTTPS only**: All cookies marked `Secure; HttpOnly; SameSite=Strict`.

---

# 8. Voice/STT/TTS Architecture

## Architecture Overview

```mermaid
sequenceDiagram
    participant U as User (Phone)
    participant FE as Frontend (PWA)
    participant BE as Backend (Go)
    participant V as Voice Provider (Sarvam)
    participant LLM as LLM Provider

    U->>FE: Taps mic, speaks
    FE->>FE: MediaRecorder captures audio (WebM/OGG)
    FE->>BE: POST /api/v1/conversation/voice {audio, lang}
    BE->>V: ASR request (audio + lang)
    V-->>BE: Transcript text
    BE->>LLM: Extract facts from transcript
    LLM-->>BE: Structured facts + follow-up question
    BE->>BE: Update profile with extracted facts
    BE->>V: TTS request (response text + lang)
    V-->>BE: Audio response (WAV/MP3)
    BE-->>FE: {transcript, response_text, response_audio_url, facts_extracted, follow_up}
    FE->>U: Display text + play audio
```

## Provider Interfaces

To ensure the frontend and backend business logic remain provider-agnostic, voice providers are abstracted behind Go interfaces.

```go
// internal/platform/voice/voice.go

type STTProvider interface {
    Transcribe(ctx context.Context, audio []byte, lang string, format string) (*TranscribeResult, error)
}

type TTSProvider interface {
    Synthesize(ctx context.Context, text string, lang string) (*SynthesizeResult, error)
}

type TranscribeResult struct {
    Text       string
    Language   string  // detected language
    Confidence float64
}

type SynthesizeResult struct {
    Audio      []byte
    Format     string  // "wav", "mp3"
    SampleRate int
}
```

## Primary Implementation: Sarvam AI

Sarvam AI is the initial voice provider due to its specialized Saaras (STT) and Bulbul (TTS) models, which excel at Indian regional accents and code-switching (e.g., Hinglish).

*   **Configuration:** Configured via `SARVAM_API_KEY`.
*   **Latency Note:** While Sarvam advertises sub-250ms latency for its models, this does NOT guarantee end-to-end latency for Sahayak. End-to-end latency will include frontend capture, network round-trip, LLM processing, and database operations.
*   **Fallback:** If Sarvam is unavailable, the system can fallback to Bhashini (if configured) or gracefully degrade to text-only mode.

## Alternative Implementation: Bhashini

Bhashini remains a viable alternative provider and can be implemented against the same interfaces, selectable via an environment variable (`VOICE_PROVIDER=sarvam|bhashini`).

## Audio Handling

- **Frontend captures**: WebM/OGG via `MediaRecorder` (best browser support)
- **Backend converts if needed**: ffmpeg for format conversion (installed in Docker image)
- **Audio is NOT stored**: Processed in memory, transcribed, then discarded
- **Max audio duration**: 60 seconds per message (configurable)
- **Max file size**: 5MB per audio upload

---

# 9. LLM Integration

## Exact LLM Responsibilities

The LLM is used as a **structured information extraction and natural language generation tool**. It is NOT the authority on eligibility.

### What the LLM DOES

| Task | Input | Output | Example |
|------|-------|--------|---------|
| **Fact extraction** | User's natural language text | Structured profile facts (JSON) | "I'm a widow with 2 acres" → `{"marital_status":"widowed","land_acres":2.0}` |
| **Missing fact detection** | Current profile + scheme requirements | List of missing facts to ask about | Profile has age but no income → ask about income |
| **Follow-up question generation** | Missing facts + language + conversation context | Natural language question in user's language | "आपकी वार्षिक आय कितनी है?" |
| **Grounded explanation** | Scheme data + eligibility result | Natural language explanation of why user qualifies | "You qualify because you are a widow over 40 with income below ₹2L" |
| **Response rephrasing** | System response + user's language | Culturally appropriate phrasing | Rephrase English template into natural Hindi |
| **Language detection** | User text | Detected language code | "मैं किसान हूँ" → "hi" |

### What the LLM MUST NOT DO

1. ❌ Decide eligibility — only the rules engine does this
2. ❌ Invent scheme criteria not in the database
3. ❌ Claim a user is eligible/ineligible without rules evaluation
4. ❌ Fabricate document acquisition instructions
5. ❌ Generate URLs or application links not from verified data
6. ❌ Store or log sensitive user information

## LLM Provider Interface

```go
// internal/platform/llm/llm.go

type LLMProvider interface {
    ExtractFacts(ctx context.Context, req ExtractFactsRequest) (*ExtractFactsResponse, error)
    GenerateFollowUp(ctx context.Context, req FollowUpRequest) (*FollowUpResponse, error)
    GenerateExplanation(ctx context.Context, req ExplanationRequest) (*ExplanationResponse, error)
    DetectLanguage(ctx context.Context, text string) (string, error)
}

type ExtractFactsRequest struct {
    UserText       string
    CurrentProfile map[string]any
    Language       string
    ConversationHistory []Message
}

type ExtractFactsResponse struct {
    Facts          map[string]any   // extracted facts
    Confidence     float64          // 0-1
    Ambiguities    []string         // things the LLM wasn't sure about
    NeedsFollowUp  []string         // fact keys that need clarification
}
```

## System Prompts

The LLM receives carefully constructed system prompts that enforce its role:

```
You are a structured information extraction assistant for Sahayak,
an Indian welfare scheme discovery platform.

Your ONLY job is to extract factual information from what the user says
and identify what information is still missing.

RULES:
1. Extract facts as a JSON object with these possible keys:
   age, gender, marital_status, income_annual, occupation, caste_category,
   residence_type, state, district, land_acres, ration_card_type,
   has_bank_account, has_aadhaar, education_level, disability_status,
   minority_status, family_members
2. Only extract facts the user actually stated. Never assume or infer.
3. If the user's statement is ambiguous, flag it in "ambiguities".
4. Return the extracted facts as valid JSON.
5. Do NOT evaluate eligibility.
6. Do NOT mention specific schemes.
7. Do NOT fabricate information.

Current profile: {current_profile}
User's message: {user_text}
```

## LLM Cost Control

- **Caching**: Cache identical/similar extraction prompts
- **Prompt optimization**: Minimal system prompts, structured output format
- **Model selection**: Use smaller/cheaper models for fact extraction (e.g., Gemini Flash, GPT-4o-mini), larger models only for complex explanation generation
- **Rate limiting**: Max 20 LLM calls per user per session
- **Fallback**: If LLM is unavailable, fall back to form-based input

---

# 10. Deterministic Eligibility Engine

## Design

The eligibility engine evaluates a user profile against scheme rules using the `expr-lang/expr` library. Rules are stored as expressions in the database.

## How It Works

```mermaid
flowchart TD
    A["User Profile (facts JSONB)"] --> B["Flatten profile to evaluation map"]
    B --> C["Load all published schemes"]
    C --> D["For each scheme, load eligibility rules"]
    D --> E["Evaluate each rule expression against profile map"]
    E --> F{"All mandatory rules pass?"}
    F -->|Yes| G["Scheme is ELIGIBLE"]
    F -->|No| H{"Any mandatory rule has MISSING data?"}
    H -->|Yes| I["Scheme is POTENTIALLY ELIGIBLE (missing info)"]
    H -->|No| J["Scheme is NOT ELIGIBLE"]
    G --> K["Rank eligible schemes by benefit/relevance"]
    I --> K
    K --> L["Return results with explanations"]
```

## Profile Flattening

The profile's JSONB facts are flattened into a typed map for `expr` evaluation:

```go
type EvaluationContext struct {
    Age               *int     `expr:"age"`
    Gender            *string  `expr:"gender"`
    MaritalStatus     *string  `expr:"marital_status"`
    IncomeAnnual      *int     `expr:"income_annual"`
    Occupation        *string  `expr:"occupation"`
    CasteCategory     *string  `expr:"caste_category"`
    ResidenceType     *string  `expr:"residence_type"`
    StateCode         *string  `expr:"state_code"`
    LandAcres         *float64 `expr:"land_acres"`
    RationCardType    *string  `expr:"ration_card_type"`
    HasBankAccount    *bool    `expr:"has_bank_account"`
    HasAadhaar        *bool    `expr:"has_aadhaar"`
    EducationLevel    *string  `expr:"education_level"`
    DisabilityStatus  *string  `expr:"disability_status"`
    MinorityStatus    *bool    `expr:"minority_status"`
    // Derived fields
    IsWidow           *bool    `expr:"is_widow"`
    HasStudentChild   *bool    `expr:"has_student_child"`
    StudentChildClass *int     `expr:"student_child_class"`
}
```

## Three-State Evaluation

Each rule evaluation returns one of three states:

| State | Meaning |
|-------|---------|
| `PASS` | Rule condition is met |
| `FAIL` | Rule condition is definitively not met |
| `UNKNOWN` | Required data for this rule is missing from the profile |

**CRITICAL**: `UNKNOWN` does NOT mean `FAIL`. A scheme with `UNKNOWN` rules is "potentially eligible" and the user should be prompted for the missing information.

```go
func EvaluateRule(expr string, ctx *EvaluationContext) RuleResult {
    // Check if required fields are nil (missing)
    requiredFields := extractFieldsFromExpression(expr)
    for _, field := range requiredFields {
        if isNil(ctx, field) {
            return RuleResult{Status: UNKNOWN, MissingFields: []string{field}}
        }
    }

    // All required fields present — evaluate
    result, err := exprEval(expr, ctx)
    if err != nil {
        return RuleResult{Status: ERROR, Error: err}
    }
    if result {
        return RuleResult{Status: PASS}
    }
    return RuleResult{Status: FAIL}
}
```

## Scheme Evaluation Result

```go
type SchemeEvaluationResult struct {
    SchemeID       string
    Status         string // "eligible", "potentially_eligible", "not_eligible"
    RuleResults    []RuleEvaluation
    MissingFields  []string    // fields needed to determine eligibility
    MatchScore     float64     // 0-1, for ranking
    PassedRules    int
    TotalRules     int
}
```

## Ranking

Eligible schemes are ranked by:
1. **Exact eligibility**: All rules pass > some rules unknown
2. **Benefit amount**: Higher benefit ranked higher (where structured data exists)
3. **Relevance score**: Based on how many profile facts matched

## Testing the Rules Engine

Every rule expression in the database must have corresponding test cases:

```go
func TestWidowPensionEligibility(t *testing.T) {
    rule := "age >= 40 && gender == 'female' && marital_status == 'widowed' && income_annual <= 200000"

    tests := []struct{
        name    string
        ctx     EvaluationContext
        want    RuleStatus
    }{
        {
            name: "eligible widow",
            ctx:  EvaluationContext{Age: ptr(45), Gender: ptr("female"), MaritalStatus: ptr("widowed"), IncomeAnnual: ptr(120000)},
            want: PASS,
        },
        {
            name: "too young",
            ctx:  EvaluationContext{Age: ptr(30), Gender: ptr("female"), MaritalStatus: ptr("widowed"), IncomeAnnual: ptr(120000)},
            want: FAIL,
        },
        {
            name: "missing income",
            ctx:  EvaluationContext{Age: ptr(45), Gender: ptr("female"), MaritalStatus: ptr("widowed")},
            want: UNKNOWN,
        },
    }
    // ...
}
```

---

# 11. Scheme Data Model

## Scheme Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Candidate: Discovered from source
    Candidate --> Draft: Accepted by reviewer
    Draft --> Verified: All fields validated
    Verified --> Published: Admin publishes
    Published --> Archived: Scheme expires/discontinued
    Candidate --> Rejected: Reviewer rejects
    Published --> Draft: Needs update
```

## Scheme Categories

| Category | Examples |
|----------|----------|
| `pension` | Widow pension, old age pension, disability pension |
| `scholarship` | Student scholarships, merit-based, minority scholarships |
| `insurance` | Crop insurance (PMFBY), health insurance (PMJAY) |
| `housing` | PMAY (urban/rural), state housing schemes |
| `health` | Ayushman Bharat, state health schemes |
| `agriculture` | PM-KISAN, crop insurance, fertilizer subsidy |
| `employment` | MGNREGA, skill development, self-employment |
| `financial` | Mudra loans, Jan Dhan accounts |
| `women_child` | Maternity benefits, nutrition schemes |
| `other` | Miscellaneous welfare schemes |

## Scheme Data Completeness Levels

Each scheme tracks what verified data is available:

```json
{
  "completeness": {
    "basic_info": true,        // name, description, benefit type
    "eligibility_rules": true, // structured rules in DB
    "required_documents": true,
    "application_process": true,
    "application_url": true,
    "benefit_amount": false,   // not available in structured form
    "deadline": false          // not applicable or unknown
  }
}
```

---

# 12. Government Source Registry

## What It Is

The source registry tracks every external source from which scheme data can be obtained. It provides:
- Provenance: Where did scheme data come from?
- Freshness: When was it last fetched?
- Reliability: How trustworthy is this source?

## Source Types

| Type | Description | Trust Level | Example |
|------|-------------|-------------|---------|
| `official_api` | Government API (data.gov.in, API Setu) | High | data.gov.in scholarship dataset |
| `gov_dataset` | Structured dataset download | High | LGD CSV exports |
| `gov_portal` | Official scheme page (manual extraction) | High | myscheme.gov.in scheme pages |
| `gazette` | Government gazette notifications | Highest | Official scheme notification PDFs |
| `manual` | Manual data entry by verified admin | Medium | Admin enters scheme from physical document |

## Source Table

See `sources` table in schema above. Each source has:
- `fetch_config` JSONB for automated fetching parameters
- `last_fetched_at` for freshness tracking
- `last_fetch_status` for health monitoring

---

# 13. Source Discovery and Controlled Fetching

## MVP vs. Future Scope

> [!IMPORTANT]
> **MVP:** Manual curation of a small set of high-impact schemes. The source registry and controlled architecture are implemented, but broad automated web scraping is deferred to avoid building a fragile scraper.
> **Later:** Automated broad discovery and fetching via scheduled background jobs.

## Architecture (Implemented for MVP)

1. **Source Registry:** The database tracks official sources (data.gov.in, scheme PDFs) to maintain provenance.
2. **Controlled API Adapters:** Where an official structured API exists (e.g., LGD dataset), implement a controlled fetcher.
3. **Manual Entry:** Admins curate schemes manually through the admin interface, linking them to registered sources.
4. **Strict Separation:** Discovered/unverified candidate schemes (if any structured data is ingested) are kept completely separate in `candidate_schemes` and NEVER enter the eligibility engine without human verification.

## Future Automated Pipeline (Phase 2)

```mermaid
flowchart TD
    A["Scheduled Job (River)"] --> B["Load active sources"]
    B --> C["For each source"]
    C --> D["Fetch content with rate limiting"]
    D --> E["Compute content hash"]
    E --> F{"Content changed?"}
    F -->|No| G["Update last_fetched_at, skip"]
    F -->|Yes| H["Parse into candidate scheme"]
    H --> I["Store in candidate_schemes table"]
    I --> J["Flag for human review"]
    J --> K["Send admin notification"]
```

---

# 14. Verification Workflow

## Who Verifies

- **Admin users** (role: `admin`) manually verify scheme data
- Verification involves:
  1. Cross-referencing scheme details with official gazette notifications
  2. Confirming eligibility criteria are correctly structured as rules
  3. Verifying required documents list
  4. Confirming application URL is correct and active
  5. Checking benefit amounts against official sources

## Verification Checklist

Each scheme has a verification checklist tracked in its metadata:

```json
{
  "verification": {
    "name_verified": true,
    "description_verified": true,
    "eligibility_rules_verified": true,
    "documents_verified": true,
    "application_url_verified": true,
    "benefit_amount_verified": false,
    "official_source_cross_checked": true,
    "verified_by": "admin-user-id",
    "verified_at": "2026-10-01T00:00:00Z",
    "verification_notes": "Cross-checked with gazette notification dated..."
  }
}
```

## Verification Status Transitions

- `draft` → `verified`: All mandatory checklist items pass
- `verified` → `published`: Admin explicitly publishes
- `published` → `draft`: Admin flags for re-verification (e.g., scheme updated)

---

# 15. Unverified Scheme Handling

## Rules

1. **Unverified schemes (`candidate_schemes` table) NEVER participate in eligibility evaluation**
2. They may optionally be shown in a separate "Discovered / Unverified" section in the admin panel
3. They are NEVER shown to citizens as confirmed eligibility results
4. Each unverified scheme displays:
   - Source name and URL
   - Retrieval date
   - "UNVERIFIED" label
   - Content hash (for change detection)
5. Promoting a candidate to a scheme requires explicit admin action

## Frontend Handling

The frontend API response for eligible schemes includes a `verification_status` field:
- `published` → shown normally
- `draft` / `verified` → shown only in admin views
- `candidate` → shown only in admin review queue

Citizens NEVER see schemes that are not `published`.

---

# 16. India-Wide Location Model

## Data Source

LGD (Local Government Directory) — canonical source for Indian administrative divisions.

## Initial Data Load

1. Download state and district CSV from [lgdirectory.gov.in/downloadDirectory.do](https://lgdirectory.gov.in/downloadDirectory.do)
2. Parse and load into `states` and `districts` tables
3. Seed `coverage_status` with `level: "none"` for all states

## All 36 States/UTs

The system seeds all 28 states + 8 Union Territories with their LGD codes.

## Coverage vs. Validity

The API clearly communicates two separate concepts:

```json
// GET /api/v1/locations/states
{
  "states": [
    {
      "id": 8,
      "code": "RJ",
      "name_en": "Rajasthan",
      "name_hi": "राजस्थान",
      "is_ut": false,
      "coverage": {
        "level": "full",
        "scheme_count": 45,
        "last_updated": "2026-10-01T00:00:00Z"
      }
    },
    {
      "id": 33,
      "code": "TN",
      "name_en": "Tamil Nadu",
      "name_hi": "तमिल नाडु",
      "is_ut": false,
      "coverage": {
        "level": "none",
        "scheme_count": 0,
        "last_updated": "2026-10-01T00:00:00Z",
        "note": "State scheme coverage is being developed. Central schemes are still available."
      }
    }
  ]
}
```

## Important Distinctions

| Situation | UI Behavior |
|-----------|-------------|
| State has full coverage | Show all matched schemes normally |
| State has partial coverage | Show matched schemes + "Coverage is partial for your state. More schemes may exist." |
| State has no coverage | Show central schemes only + "We don't yet have state-specific scheme data for your state. Only central government schemes are shown." |
| No scheme matches | "No matching schemes found based on your current profile. This does NOT mean you are ineligible for all government schemes." |

---

# 17. Profile/Intake Flow

## Intake State Machine

```mermaid
stateDiagram-v2
    [*] --> Welcome: User opens Sahayak
    Welcome --> LanguageSelect: Select language
    LanguageSelect --> FreeformInput: "Tell me about yourself"
    FreeformInput --> FactExtraction: LLM extracts facts
    FactExtraction --> FollowUp: Missing required facts?
    FollowUp --> FactExtraction: User answers
    FactExtraction --> ProfileReview: Enough facts gathered
    ProfileReview --> ProfileConfirmed: User confirms
    ProfileReview --> FactCorrection: User corrects
    FactCorrection --> ProfileReview: Show updated
    ProfileConfirmed --> EligibilityResults: Run eligibility engine
```

## Required vs. Optional Facts

| Fact | Required for Matching | How Obtained |
|------|----------------------|--------------|
| `state` | Yes (for state schemes) | Asked or detected |
| `gender` | Yes (many schemes) | Extracted or asked |
| `age` | Yes (many schemes) | Extracted or asked |
| `marital_status` | Conditional | Extracted or asked |
| `income_annual` | Conditional | Asked |
| `caste_category` | Conditional | Asked (sensitively) |
| `residence_type` | Conditional | Asked (rural/urban) |
| `occupation` | Helpful | Extracted or asked |
| `has_aadhaar` | Document check | Asked |
| `has_bank_account` | Document check | Asked |
| `ration_card_type` | Conditional | Asked |
| `land_acres` | Conditional | Extracted or asked |
| `education_level` | Conditional | Extracted or asked |
| `family_members` | Conditional | Extracted or asked |

## Intake Question Priority

The system asks questions in this order:
1. **State/location** (needed to scope schemes)
2. **Gender + age** (filters most schemes)
3. **Marital status** (if relevant based on initial input)
4. **Income** (many schemes are income-gated)
5. **Occupation** (farmer/student/self-employed)
6. **Caste/category** (only if needed for matching; asked respectfully)
7. **Document inventory** (what documents the user has)
8. **Family details** (if relevant for family benefit schemes)

The LLM determines which facts are still missing and generates the next question.

## Conversation Context Management

Each conversation stores a `context` JSONB field tracking:
```json
{
  "state": "follow_up",
  "facts_collected": ["age", "gender", "state", "marital_status"],
  "facts_missing": ["income_annual", "caste_category"],
  "current_question_about": "income_annual",
  "turn_count": 4,
  "max_turns": 15
}
```

The backend limits conversations to 15 turns before offering a form-based fallback.

---

# 18. Missing Document System

## How It Works

1. User's profile includes document inventory (`profile_documents` table)
2. Each eligible scheme has required documents (`scheme_required_documents` table)
3. The system computes the diff:

```go
type DocumentCheckResult struct {
    SchemeID    string
    Documents   []DocumentStatus
    MissingCount int
    AllPresent   bool
}

type DocumentStatus struct {
    DocumentType    string
    DocumentName    string // localized
    IsMandatory     bool
    UserHas         bool   // from profile_documents
    AcquisitionInfo *DocumentAcquisitionInfo // from document_types table
}
```

## API Response

```json
{
  "scheme_id": "...",
  "documents": [
    {
      "type": "aadhaar",
      "name": "आधार कार्ड",
      "mandatory": true,
      "user_has": true,
      "status": "available"
    },
    {
      "type": "income_cert",
      "name": "आय प्रमाण पत्र",
      "mandatory": true,
      "user_has": false,
      "status": "missing",
      "acquisition": {
        "issuing_authority": "तहसीलदार / SDM कार्यालय",
        "steps": "1. तहसील कार्यालय जाएं\n2. आवेदन पत्र भरें\n3. आवश्यक दस्तावेज़ जमा करें\n4. सत्यापन के बाद प्रमाण पत्र प्राप्त करें",
        "typical_time": "7-15 कार्य दिवस",
        "typical_fee": "₹50",
        "source_url": "https://...",
        "nearest_offices": [
          {
            "name": "CSC Jaipur Central",
            "address": "...",
            "distance_km": 2.3,
            "lat": 26.9124,
            "lng": 75.7873
          }
        ]
      }
    }
  ]
}
```

---

# 19. Document Acquisition Guidance

## Data Source

Document acquisition guidance is stored in the `document_types` table. This data is:
- **Manually curated** from official government sources
- **Never LLM-generated** — all steps, URLs, and fees come from verified sources
- **Timestamped** with `last_verified` date
- **Source-linked** with `source_url`

## Supported Document Types (Initial Set)

| ID | Name | Issuing Authority |
|----|------|-------------------|
| `aadhaar` | Aadhaar Card | UIDAI |
| `income_cert` | Income Certificate | Revenue Department / Tehsil |
| `domicile_cert` | Domicile Certificate | Revenue Department / Tehsil |
| `caste_cert` | Caste Certificate | Revenue Department / Tehsil |
| `bank_account` | Bank Account | Any nationalized bank |
| `ration_card` | Ration Card | Food & Civil Supplies Dept |
| `voter_id` | Voter ID (EPIC) | Election Commission |
| `birth_cert` | Birth Certificate | Municipal Corporation / Gram Panchayat |
| `pan_card` | PAN Card | Income Tax Department |
| `disability_cert` | Disability Certificate | District Medical Board |
| `land_record` | Land Ownership Record | Revenue Department |
| `student_cert` | Bonafide Student Certificate | Educational Institution |
| `death_cert` | Death Certificate (of spouse, for widow schemes) | Municipal Corporation |
| `marriage_cert` | Marriage Certificate | Marriage Registrar |

## Location-Based Guidance

Where the user's district is known, the system can suggest the nearest CSC or tehsil office from the `service_locations` table. This data is seeded manually for pilot locations and expanded over time.

---

# 20. Application Form Generation

## How It Works

1. Each scheme has `scheme_application_fields` defining the form structure
2. Each field has an optional `profile_mapping` that maps to the user's profile facts
3. The backend auto-fills fields from the profile
4. The frontend renders the form for user review
5. User can edit any field
6. Finalized form data stored in `applications.form_data`

## Auto-Fill Logic

```go
func AutoFillForm(scheme *Scheme, profile *Profile) map[string]any {
    formData := make(map[string]any)
    for _, field := range scheme.ApplicationFields {
        if field.ProfileMapping != "" {
            value := extractFromProfile(profile.Facts, field.ProfileMapping)
            if value != nil {
                formData[field.FieldName] = value
            }
        }
    }
    return formData
}
```

## Form Review Flow

1. Backend returns `{ "form_fields": [...], "auto_filled": {...}, "missing_fields": [...] }`
2. Frontend renders form with auto-filled values highlighted
3. User reviews, edits, fills missing fields
4. User submits form data back to backend
5. Backend validates all required fields are present
6. Backend stores in `applications` table with `status: "ready"`

## No Silent Submission

The system NEVER submits applications without explicit user action. The flow is:
1. Generate form → 2. User reviews → 3. User downloads PDF → 4. User submits through official channel

If an official API integration exists (future), the user must explicitly confirm submission.

---

# 21. PDF Generation

## Implementation

Use `maroto` v2 for structured PDF generation:

```go
type PDFGenerator struct {
    fontPath string // path to Noto Sans Devanagari font
}

func (g *PDFGenerator) GenerateApplicationPDF(scheme *Scheme, formData map[string]any, lang string) ([]byte, error) {
    m := maroto.New()

    // Header with scheme name
    m.AddRow(15, text.NewCol(12, scheme.Name(lang), props.Text{Size: 16, Style: fontstyle.Bold}))

    // Form fields
    for _, field := range scheme.ApplicationFields {
        value := formData[field.FieldName]
        m.AddRow(10,
            text.NewCol(4, field.Label(lang), props.Text{Size: 10, Style: fontstyle.Bold}),
            text.NewCol(8, fmt.Sprintf("%v", value), props.Text{Size: 10}),
        )
    }

    // Footer with generation timestamp and disclaimer
    m.AddRow(10, text.NewCol(12,
        "Generated by Sahayak | This is not an official submission",
        props.Text{Size: 8, Color: color.Gray}))

    return m.Generate()
}
```

## Font Support

- Embed **Noto Sans Devanagari** for Hindi and related scripts
- Embed **Noto Sans Tamil**, **Noto Sans Telugu**, etc. for regional languages
- Fonts included in Docker image at `/app/fonts/`

## PDF Storage

- PDFs stored in local filesystem: `/data/pdfs/{user_id}/{application_id}.pdf`
- Served via authenticated API endpoint: `GET /api/v1/applications/{id}/pdf`
- Retained for 90 days, then cleaned up

---

# 22. Application Status Tracking

## Architecture

Status tracking has two modes:

### Mode 1: Manual Tracking (Default)

User manually enters their application reference number and current status. The system stores this and sends reminders.

### Mode 2: Automated Tracking (Future, per-scheme)

For schemes that have a publicly accessible status-check API or portal:
1. Store the reference number
2. Scheduled River job checks the status periodically (daily)
3. If status changes, notify the user

## Implementation

```go
// Scheduled job for automated tracking
type StatusCheckJob struct {
    ApplicationID string
    SchemeID      string
    ReferenceNo   string
    CheckURL      string // scheme's status check URL
}

// Status check is scheme-specific — each scheme with tracking support
// has a status checker implementation
type StatusChecker interface {
    CheckStatus(ctx context.Context, referenceNo string) (*StatusResult, error)
}
```

## Honesty Rule

The system NEVER claims an application has been submitted or approved unless:
1. There is an actual integration that verified the status
2. The user manually reported the status

The default message for tracked applications is: "Status: Awaiting update. Last checked: [date]"

---

# 23. WhatsApp/SMS Notifications

## Architecture

```mermaid
flowchart LR
    APP["Application Module"] -->|"Enqueue job"| RIVER["River Job Queue"]
    SCHEDULER["Scheduled Jobs"] -->|"Enqueue job"| RIVER
    RIVER --> WORKER["Notification Worker"]
    WORKER --> DISPATCH["Dispatch by channel"]
    DISPATCH -->|WhatsApp| WA["WhatsApp BSP"]
    DISPATCH -->|SMS| SMS["SMS Provider"]
    WA --> DELIVERY["Delivery tracking"]
    SMS --> DELIVERY
    DELIVERY --> DB["Update notification status"]
```

## Provider Interfaces

```go
// internal/platform/whatsapp/whatsapp.go
type WhatsAppProvider interface {
    SendTemplate(ctx context.Context, phone string, templateName string, params map[string]string) (*SendResult, error)
    SendText(ctx context.Context, phone string, message string) (*SendResult, error)
}

// internal/platform/sms/sms.go
type SMSProvider interface {
    Send(ctx context.Context, phone string, message string, templateID string) (*SendResult, error)
}
```

## Notification Types

| Type | Channel | Trigger |
|------|---------|---------|
| Deadline reminder | WhatsApp + SMS | 7 days before scheme deadline |
| Status update | WhatsApp | Application status change detected |
| Follow-up reminder | WhatsApp | User abandoned profile completion |
| New scheme alert | WhatsApp | New scheme matching user's profile published |

## Opt-In/Opt-Out

- Notifications are opt-in (user must explicitly enable)
- Stored in `notification_preferences` table
- WhatsApp requires user to initiate conversation first (WhatsApp policy)
- Opt-out via reply "STOP" or through the app

## DLT Compliance (SMS)

- All SMS templates must be registered with TRAI DLT
- Template IDs stored in configuration
- Template content pre-approved and stored

## Recommended Providers

| Channel | Primary | Backup |
|---------|---------|--------|
| WhatsApp | Gupshup | MSG91 |
| SMS | MSG91 | Kaleyra |

---

# 24. ASHA/CSC Helper Mode

## What It Is

Helper mode allows authorized welfare workers to assist citizens through the Sahayak process. The helper uses their own authenticated session to guide a citizen.

## How It Differs from Citizen Mode

| Aspect | Citizen Mode | Helper Mode |
|--------|-------------|-------------|
| Authentication | Phone OTP | Phone OTP + verified helper profile |
| Session scope | Own profile only | Can create/manage profiles for citizens |
| Consent | Implicit (own data) | Explicit recorded consent required |
| Multi-citizen | Single profile | Multiple citizen sessions |
| Audit | Standard logging | Enhanced audit trail |
| Data access | Own data only | Citizen's data only during active session |
| Privacy | Standard | Helper cannot see data after session ends |

## Helper Session Flow

```mermaid
sequenceDiagram
    participant H as Helper
    participant S as Sahayak
    participant C as Citizen

    H->>S: Login (helper credentials)
    S-->>H: Helper dashboard
    H->>S: Start new citizen session
    S-->>H: Consent form displayed
    H->>C: Read consent statement aloud
    C->>H: Gives verbal consent
    H->>S: Record consent (checkbox + timestamp)
    S-->>H: Citizen intake begins
    H->>S: Enter citizen info (voice or form)
    S-->>H: Follow-up questions
    H->>S: Answer on behalf of citizen
    S-->>H: Profile review
    H->>C: Verify profile with citizen
    H->>S: Confirm profile
    S-->>H: Eligible schemes displayed
    H->>C: Explain schemes to citizen
    H->>S: Select scheme, generate application
    S-->>H: Application PDF
    H->>S: End session
    S-->>H: Session summary + audit record
```

## Consent Mechanism

Before any citizen data is entered in helper mode:

1. Helper must confirm consent was obtained
2. System records:
   - Consent text (exact statement shown)
   - Timestamp
   - Helper user ID
   - Citizen identifier (phone hash or anonymous)
3. Consent record stored in `helper_sessions` table
4. Without consent, session cannot proceed

## Privacy Protections

1. **Session isolation**: Helper can only access citizen data during active session
2. **No data retention by helper**: After session ends, helper cannot retrieve citizen profile
3. **Audit trail**: Every action in helper mode is logged with enhanced detail
4. **Rate limiting**: Helpers limited to reasonable session counts per day (e.g., 50)
5. **No bulk export**: Helpers cannot export citizen data

## Helper Verification

Helpers are verified by admins through:
1. Helper registers with helper type (ASHA, CSC, field worker)
2. Provides organization and ID number (encrypted in DB)
3. Admin verifies identity (cross-reference with CSC/ASHA databases)
4. Admin sets `verified: true` in `helper_profiles`

---

# 25. Privacy and Security

## Data Minimization

| Data Type | Stored? | How? |
|-----------|---------|------|
| Phone number | Hashed only | SHA-256 + pepper; last 4 digits separate |
| Aadhaar number | NO | Only `has_aadhaar: true/false` |
| Bank account number | NO | Only `has_bank_account: true/false` |
| Voice audio | NO | Processed in memory, discarded after STT |
| Conversation transcripts | Minimal | Text only, no PII in logs; auto-deleted after 90 days |
| Profile facts | Yes (encrypted at rest) | JSONB in PostgreSQL with TDE or column encryption |
| Document images | NO | Not collected |

## Encryption

- **In transit**: TLS 1.3 (HTTPS)
- **At rest**: PostgreSQL TDE (Transparent Data Encryption) or `pgcrypto` for sensitive columns
- **Secrets**: Environment variables, never in code. Coolify manages secrets.

## Access Controls

| Role | Can Access |
|------|-----------|
| Citizen | Own profile, own applications, own notifications |
| Helper | Citizen data ONLY during active session, own helper profile |
| Admin | All data, scheme management, source management, user management |

## Retention Policies

| Data | Retention | Action After |
|------|-----------|-------------|
| Conversation messages | 90 days | Soft delete (facts preserved in profile) |
| OTP challenges | 24 hours | Hard delete |
| Refresh tokens | 30 days | Hard delete on revocation/expiry |
| Audit logs | 2 years | Archive to cold storage |
| Application PDFs | 90 days | Delete file, keep metadata |
| Notifications | 1 year | Soft delete |

## Consent

- User agrees to data collection terms on first use
- Explicit opt-in for notifications
- Helper mode requires explicit citizen consent per session
- User can request data deletion (GDPR-like right, good practice)

---

# 26. Rate Limiting and Abuse Protection

## Implementation

Use a token-bucket rate limiter implemented in Go middleware, backed by PostgreSQL counters (no Redis needed):

```go
// Middleware using in-memory rate limiter (golang.org/x/time/rate)
// with per-IP and per-user limits

type RateLimitConfig struct {
    // Anonymous (unauthenticated) requests
    AnonPerMinute    int // 30
    AnonPerHour      int // 200

    // Authenticated citizen requests
    CitizenPerMinute int // 60
    CitizenPerHour   int // 500

    // Voice/LLM requests (expensive)
    VoicePerMinute   int // 10
    VoicePerHour     int // 100

    // OTP requests
    OTPPerPhone10Min int // 3
    OTPPerPhoneDay   int // 10

    // Helper requests
    HelperSessionsPerDay int // 50
}
```

## Abuse Protection

- **Bot detection**: Require minimum time between conversation turns (2 seconds)
- **Audio validation**: Reject audio files that are silent or too short (<0.5s)
- **LLM abuse**: Cap LLM calls per user per session (20) and per day (100)
- **Registration spam**: Rate limit OTP requests aggressively
- **Content validation**: Sanitize all user input before storage

---

# 27. Audit Logging

## What Gets Logged

| Event | Actor | Details |
|-------|-------|---------|
| `auth.otp.request` | System | Phone hash, success/failure |
| `auth.login` | User | Login method, IP |
| `profile.create` | User | Profile ID |
| `profile.update` | User | Changed fields (keys only, not values) |
| `profile.view` | User/Helper | Profile ID |
| `eligibility.evaluate` | User | Profile ID, scheme count |
| `scheme.publish` | Admin | Scheme ID, name |
| `scheme.archive` | Admin | Scheme ID, reason |
| `application.create` | User | Application ID, scheme ID |
| `application.pdf.generate` | User | Application ID |
| `helper.session.start` | Helper | Session ID, consent status |
| `helper.session.end` | Helper | Session ID, duration |
| `notification.send` | System | Channel, template, status |

## Sensitive Data in Audit Logs

- **NEVER log**: PII, phone numbers, profile fact values, OTP codes
- **DO log**: Resource IDs, action types, timestamps, actor IDs, IP addresses

## Storage

- `audit_log` table in PostgreSQL
- Partitioned by month for performance
- Retained for 2 years
- Archived to compressed backup after 6 months

---

# 28. Observability

## Logging

- **Structured JSON logging** via `slog` (Go standard library, Go 1.21+)
- Log levels: DEBUG, INFO, WARN, ERROR
- Request ID on every log line
- No PII in logs

## Metrics

Use **Prometheus** client library for Go (`prometheus/client_golang`):

| Metric | Type | Labels |
|--------|------|--------|
| `http_requests_total` | Counter | method, path, status |
| `http_request_duration_seconds` | Histogram | method, path |
| `llm_requests_total` | Counter | provider, model, status |
| `llm_request_duration_seconds` | Histogram | provider, model |
| `voice_stt_requests_total` | Counter | provider, language, status |
| `voice_stt_duration_seconds` | Histogram | provider, language |
| `eligibility_evaluations_total` | Counter | - |
| `eligibility_evaluation_duration_seconds` | Histogram | - |
| `notifications_sent_total` | Counter | channel, status |
| `river_jobs_total` | Counter | queue, status |
| `active_conversations` | Gauge | - |
| `active_helper_sessions` | Gauge | - |

## Health Check

`GET /healthz` returns:

```json
{
  "status": "ok",
  "checks": {
    "database": "ok",
    "river": "ok"
  },
  "version": "0.1.0",
  "uptime_seconds": 86400
}
```

## Monitoring Stack (Simple)

For a small project:
- **Prometheus**: Metrics collection (or Coolify's built-in monitoring)
- **Grafana**: Dashboards (optional, Coolify may provide basic metrics)
- **PostgreSQL logs**: For slow query analysis
- **Application logs**: JSON to stdout, captured by Docker/Coolify

---

# 29. Database Backups

## Strategy

1. **Automated daily backups** via `pg_dump`
2. **Stored in**: OCI Object Storage or local backup volume
3. **Retention**: 30 daily + 12 weekly + 6 monthly
4. **Tested monthly**: Restore to a test database to verify integrity

## Implementation

```bash
#!/bin/bash
# backup.sh - run daily via cron or River scheduled job
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="/backups/sahayak_${TIMESTAMP}.sql.gz"

pg_dump -h localhost -U sahayak -d sahayak | gzip > "${BACKUP_FILE}"

# Upload to OCI Object Storage (optional)
# oci os object put --bucket-name sahayak-backups --file "${BACKUP_FILE}"

# Cleanup: keep last 30 days
find /backups -name "sahayak_*.sql.gz" -mtime +30 -delete
```

## Docker Volume for Backups

```yaml
volumes:
  db-data:
    driver: local
  db-backups:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/sahayak/backups
```

---

# 30. Docker Setup

## Docker Compose

```yaml
# docker-compose.yml
version: "3.9"

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://sahayak:${DB_PASSWORD}@db:5432/sahayak?sslmode=disable
      - BHASHINI_API_KEY=${BHASHINI_API_KEY}
      - BHASHINI_USER_ID=${BHASHINI_USER_ID}
      - LLM_API_KEY=${LLM_API_KEY}
      - LLM_PROVIDER=${LLM_PROVIDER:-gemini}
      - SMS_API_KEY=${SMS_API_KEY}
      - WHATSAPP_API_KEY=${WHATSAPP_API_KEY}
      - JWT_SECRET=${JWT_SECRET}
      - PHONE_HASH_PEPPER=${PHONE_HASH_PEPPER}
      - SERVER_MODE=all  # server + worker
    depends_on:
      db:
        condition: service_healthy
    volumes:
      - app-data:/data
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 5s
      retries: 3
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=sahayak
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=sahayak
    volumes:
      - db-data:/var/lib/postgresql/data
      - db-backups:/backups
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U sahayak"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - PUBLIC_API_URL=http://api:8080
    depends_on:
      - api
    restart: unless-stopped

volumes:
  db-data:
  db-backups:
  app-data:
```

## Backend Dockerfile

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /sahayak ./cmd/server

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata ffmpeg wkhtmltopdf
COPY --from=builder /sahayak /sahayak
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/fonts /fonts
EXPOSE 8080
ENTRYPOINT ["/sahayak"]
```

## Frontend Dockerfile

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json .
EXPOSE 3000
CMD ["node", "build"]
```

---

# 31. Coolify Deployment

## Setup

1. **Coolify installed on OCI VPS** (Ubuntu 22.04+, minimum 4GB RAM, 2 vCPU)
2. **Docker + Docker Compose** managed by Coolify
3. **HTTPS via Coolify's built-in Traefik** reverse proxy with Let's Encrypt
4. **Domain**: Point DNS A record to VPS IP

## Coolify Configuration

- **Service type**: Docker Compose
- **Source**: Git repository (GitHub/GitLab)
- **Build**: Automatic on push to `main` branch
- **Environment variables**: Set in Coolify UI (secrets encrypted)
- **Persistent volumes**: Map to host paths for DB data and backups
- **Health checks**: Coolify monitors `/healthz` endpoint
- **Rollback**: Coolify keeps previous images for rollback

## VPS Sizing

| Component | Min Requirement |
|-----------|----------------|
| CPU | 2 vCPU (4 recommended) |
| RAM | 4 GB (8 recommended) |
| Disk | 50 GB SSD |
| OS | Ubuntu 22.04 LTS |
| Network | Public IP with ports 80, 443 |

---

# 32. CI/CD

## Pipeline (GitHub Actions)

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test-backend:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: sahayak_test
        ports:
          - 5432:5432
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: go mod download
      - run: go vet ./...
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go build ./cmd/server

  test-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: "20"
      - run: cd frontend && corepack enable && pnpm install
      - run: cd frontend && pnpm check
      - run: cd frontend && pnpm test
      - run: cd frontend && pnpm build

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  deploy:
    needs: [test-backend, test-frontend, lint]
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Coolify deployment
        run: |
          curl -X POST "${{ secrets.COOLIFY_WEBHOOK_URL }}" \
            -H "Authorization: Bearer ${{ secrets.COOLIFY_API_KEY }}"
```

---

# 33. Testing Strategy

## Test Pyramid

| Level | Focus | Tools | Coverage Target |
|-------|-------|-------|----------------|
| **Unit** | Individual functions, services | Go `testing`, testify | 80%+ for core logic |
| **Integration** | Module interactions, DB queries | testcontainers-go, real PG | All DB operations |
| **API** | HTTP endpoints end-to-end | httptest, real DB | All API endpoints |
| **E2E** | Full user flows | Playwright (frontend) | Core user journey |

## Critical Test Areas

### Eligibility Engine (Highest Priority)

```go
// Every scheme rule must have tests:
// 1. Profile that PASSES all rules
// 2. Profile that FAILS specific rules
// 3. Profile with MISSING data (should return UNKNOWN, not FAIL)
// 4. Edge cases (boundary values)
// 5. Empty profile (all UNKNOWN)
```

### Voice Pipeline

```go
// Test with recorded audio samples in each supported language
// Mock Bhashini provider for unit tests
// Integration test with actual Bhashini API (CI only, rate limited)
```

### LLM Fact Extraction

```go
// Test with diverse natural language inputs:
// "I'm a widow, I have 2 acres" → {marital_status: widowed, land_acres: 2.0}
// "मेरी उम्र ५० साल है" → {age: 50}
// Mock LLM provider for unit tests with canned responses
```

### Auth Flow

```go
// OTP generation, verification, expiry, rate limiting
// JWT creation, validation, expiry, refresh
// Role-based access control
```

---

# 34. Failure and Fallback Behaviour

| Component | Failure | Fallback |
|-----------|---------|----------|
| **Bhashini ASR** | API timeout/error | Offer text input; queue retry |
| **Bhashini TTS** | API timeout/error | Return text-only response |
| **LLM** | API timeout/error | Fall back to form-based profile input |
| **PostgreSQL** | Connection lost | Return 503; health check reports unhealthy |
| **SMS provider** | Delivery failure | Retry via River job (3 attempts); try backup provider |
| **WhatsApp provider** | Delivery failure | Retry; fall back to SMS |
| **PDF generation** | Render error | Return form data as JSON; suggest user downloads later |
| **Source fetch** | HTTP error | Log error; keep existing data; retry with backoff |
| **Eligibility eval** | Expression error | Log error; skip that rule; show "could not evaluate" for that scheme |

## Circuit Breaker Pattern

For external APIs (Bhashini, LLM, SMS):
- Track failure rate in a 1-minute window
- If failure rate > 50% (min 5 requests), open circuit
- When open: immediately return fallback (no external call)
- After 30 seconds: half-open (try one request)
- If success: close circuit; if fail: stay open

```go
type CircuitBreaker struct {
    Name          string
    FailureThreshold int
    ResetTimeout     time.Duration
    state         State // closed, open, half-open
    failures      int
    lastFailure   time.Time
}
```

---

# 35. Data Freshness and Verification

## Scheme Data Freshness

| Data Point | Refresh Frequency | Method |
|-----------|-------------------|--------|
| Central scheme details | Monthly | Manual review by admin |
| State scheme details | Monthly | Manual review by admin |
| Eligibility rules | On gazette notification | Manual update |
| Document requirements | Quarterly | Manual review |
| Application URLs | Weekly | Automated link checker (River job) |
| Benefit amounts | On official update | Manual update |

## Link Verification Job

A scheduled River job checks scheme application URLs weekly:

```go
type LinkCheckJob struct{}

func (j *LinkCheckJob) Work(ctx context.Context) error {
    schemes := loadPublishedSchemes(ctx)
    for _, scheme := range schemes {
        if scheme.ApplicationURL != "" {
            resp, err := http.Head(scheme.ApplicationURL)
            if err != nil || resp.StatusCode >= 400 {
                flagSchemeForReview(ctx, scheme.ID, "application_url_broken")
            }
        }
    }
    return nil
}
```

## Freshness Display

Every scheme shown to the user includes:
```json
{
  "verified_at": "2026-10-01T00:00:00Z",
  "last_checked": "2026-10-05T00:00:00Z",
  "source": "Official Gazette Notification, dated 2026-09-15",
  "freshness_note": "Verified 4 days ago"
}
```

---

# 36. Multilingual Architecture

## Design Principles

1. **Language is a first-class parameter** on every API call
2. **All user-facing text** has at least English + Hindi versions
3. **Regional language support** added incrementally without code changes
4. **Scheme data** stored in multiple languages in the database
5. **UI strings** managed via compile-time i18n (paraglide-js)
6. **Dynamic content** (LLM responses, explanations) generated in the user's language

## Language Code Convention

Use ISO 639-1/639-3 codes:
`en`, `hi`, `ta`, `te`, `kn`, `bn`, `mr`, `gu`, `pa`, `or`, `as`, `ur`, `bho`, `raj`

## Multilingual Data Storage

Scheme names, descriptions, and document names are stored with language suffixes:
```sql
name_en TEXT NOT NULL,
name_hi TEXT NOT NULL,
-- Future: name_ta, name_te, etc.
```

For now, English and Hindi are required. Other languages use Bhashini NMT for on-the-fly translation with caching:

```go
type TranslationCache struct {
    // key: source_lang + target_lang + content_hash
    // value: translated text
    // stored in PostgreSQL table for persistence
}
```

## Frontend i18n

Using `paraglide-js`:
```
frontend/src/lib/i18n/messages/
├── en.json    # English UI strings
├── hi.json    # Hindi UI strings
├── ta.json    # Tamil UI strings (initial)
└── ...
```

All UI strings go through paraglide. No hardcoded strings in components.

---

# 37. Frontend/Backend Contract

## API Contract

- **OpenAPI 3.0 spec** generated from Go code annotations or maintained manually
- Frontend uses a **generated TypeScript API client** from the OpenAPI spec
- Contract tests ensure backend responses match the spec

## Key Principles

1. Frontend calls backend for ALL data
2. Frontend renders what the backend returns — no client-side filtering or rule evaluation
3. Audio uploads go to the backend; backend handles STT/TTS
4. Eligibility results come pre-computed from the backend
5. Form field definitions come from the backend
6. Language content comes from the backend (or from i18n files for UI chrome)

## Error Contract

```typescript
interface ApiError {
    error: {
        code: string;        // e.g., "INVALID_INPUT", "UNAUTHORIZED", "RATE_LIMITED"
        message: string;     // human-readable message
        details?: Record<string, string>;  // field-level errors
    }
}
```

---

# 38. PWA Requirements

## Manifest

```json
{
  "name": "Sahayak - सहायक",
  "short_name": "Sahayak",
  "description": "Discover government welfare schemes you're eligible for",
  "start_url": "/",
  "display": "standalone",
  "orientation": "portrait",
  "background_color": "#ffffff",
  "theme_color": "#16a34a",
  "icons": [
    { "src": "/icons/icon-192.png", "sizes": "192x192", "type": "image/png" },
    { "src": "/icons/icon-512.png", "sizes": "512x512", "type": "image/png" }
  ]
}
```

## Service Worker

- Cache app shell (HTML, CSS, JS, icons)
- Cache API responses for offline viewing of previously loaded schemes
- Do NOT promise offline scheme discovery (eligibility requires backend)
- Show "You are offline" message when network unavailable
- Queue notification opt-in for when connection returns

## Performance Targets

| Metric | Target |
|--------|--------|
| First Contentful Paint | < 2s on 3G |
| Largest Contentful Paint | < 4s on 3G |
| Time to Interactive | < 5s on 3G |
| Total bundle size (initial) | < 200KB gzipped |
| Lighthouse PWA score | > 90 |

---

# 39. Repository Structure

```
sahayak/
├── cmd/
│   └── server/
│       └── main.go                    # Application entrypoint
├── internal/
│   ├── config/
│   │   └── config.go                  # Environment-based configuration
│   ├── server/
│   │   └── server.go                  # HTTP server setup, routes, middleware
│   ├── domain/
│   │   ├── profile.go                 # Shared domain types
│   │   ├── scheme.go
│   │   └── errors.go
│   ├── middleware/
│   │   ├── auth.go                    # JWT validation middleware
│   │   ├── ratelimit.go               # Rate limiting
│   │   ├── logging.go                 # Request logging
│   │   ├── requestid.go               # Request ID
│   │   └── cors.go                    # CORS headers
│   ├── auth/
│   │   ├── handler.go                 # HTTP handlers
│   │   ├── service.go                 # Business logic
│   │   ├── repository.go             # Database operations
│   │   └── auth_test.go
│   ├── location/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── seed.go                    # LGD data seeding
│   │   └── location_test.go
│   ├── profile/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── profile_test.go
│   ├── conversation/
│   │   ├── handler.go
│   │   ├── service.go                 # Orchestrates voice→STT→LLM→facts→TTS
│   │   ├── intake.go                  # Intake state machine
│   │   ├── repository.go
│   │   └── conversation_test.go
│   ├── scheme/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── scheme_test.go
│   ├── eligibility/
│   │   ├── handler.go
│   │   ├── engine.go                  # Rules evaluation engine
│   │   ├── evaluator.go              # expr-lang integration
│   │   ├── ranker.go                  # Result ranking
│   │   ├── service.go
│   │   └── eligibility_test.go       # Extensive rule tests
│   ├── document/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── document_test.go
│   ├── application/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── pdf.go                     # PDF generation
│   │   ├── repository.go
│   │   └── application_test.go
│   ├── notification/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── worker.go                  # River job workers
│   │   ├── templates.go              # Notification templates
│   │   ├── repository.go
│   │   └── notification_test.go
│   ├── source/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── fetcher.go                # Source fetching logic
│   │   ├── repository.go
│   │   └── source_test.go
│   ├── helper/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── consent.go                # Consent management
│   │   ├── repository.go
│   │   └── helper_test.go
│   ├── audit/
│   │   ├── logger.go                  # Audit log writer
│   │   ├── repository.go
│   │   └── audit_test.go
│   └── platform/
│       ├── llm/
│       │   ├── llm.go                 # Interface
│       │   ├── gemini.go              # Google Gemini implementation
│       │   ├── openai.go              # OpenAI implementation
│       │   └── llm_test.go
│       ├── voice/
│       │   ├── voice.go               # Interface
│       │   ├── bhashini.go            # Bhashini implementation
│       │   ├── google.go              # Google Cloud fallback
│       │   └── voice_test.go
│       ├── sms/
│       │   ├── sms.go                 # Interface
│       │   ├── msg91.go               # MSG91 implementation
│       │   └── sms_test.go
│       ├── whatsapp/
│       │   ├── whatsapp.go            # Interface
│       │   ├── gupshup.go            # Gupshup implementation
│       │   └── whatsapp_test.go
│       └── translate/
│           ├── translate.go           # Interface
│           ├── bhashini_nmt.go       # Bhashini NMT implementation
│           └── translate_test.go
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_locations.up.sql
│   ├── 000002_create_locations.down.sql
│   ├── ...
│   └── seed/
│       ├── states.sql                 # LGD state/UT data
│       ├── districts.sql              # LGD district data
│       ├── document_types.sql         # Document type reference data
│       └── initial_schemes.sql        # Initial curated schemes
├── sqlc/
│   ├── sqlc.yaml                      # sqlc configuration
│   └── queries/
│       ├── users.sql
│       ├── profiles.sql
│       ├── schemes.sql
│       ├── eligibility.sql
│       └── ...
├── fonts/
│   ├── NotoSansDevanagari-Regular.ttf
│   ├── NotoSansTamil-Regular.ttf
│   └── ...
├── frontend/                          # SvelteKit PWA (see Section 4)
├── scripts/
│   ├── backup.sh
│   ├── seed.sh
│   └── generate-sqlc.sh
├── .github/
│   └── workflows/
│       └── ci.yml
├── Dockerfile
├── docker-compose.yml
├── docker-compose.prod.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

# 40. API Endpoint Specification

## Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/auth/otp/request` | None | Request OTP for phone number |
| `POST` | `/api/v1/auth/otp/verify` | None | Verify OTP, get tokens |
| `POST` | `/api/v1/auth/refresh` | Refresh token | Refresh access token |
| `POST` | `/api/v1/auth/logout` | Bearer | Revoke refresh token |

### `POST /api/v1/auth/otp/request`

```json
// Request
{ "phone": "+919876543210" }
// Response 200
{ "challenge_id": "uuid", "expires_in": 300 }
// Error 429 (rate limited)
{ "error": { "code": "RATE_LIMITED", "message": "Too many OTP requests. Try again in 10 minutes." } }
```

### `POST /api/v1/auth/otp/verify`

```json
// Request
{ "challenge_id": "uuid", "otp": "123456" }
// Response 200
{
  "access_token": "eyJ...",
  "refresh_token": "uuid",
  "expires_in": 900,
  "user": { "id": "uuid", "role": "citizen", "preferred_lang": "hi" }
}
```

## Location

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/locations/states` | None | List all states/UTs with coverage |
| `GET` | `/api/v1/locations/states/{id}/districts` | None | List districts for a state |

## Profile

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/profile` | Bearer | Get current user's profile |
| `PUT` | `/api/v1/profile` | Bearer | Update profile |
| `PATCH` | `/api/v1/profile/facts` | Bearer | Merge new facts into profile |
| `GET` | `/api/v1/profile/documents` | Bearer | Get document inventory |
| `PUT` | `/api/v1/profile/documents` | Bearer | Update document inventory |

### `GET /api/v1/profile`

```json
// Response 200
{
  "id": "uuid",
  "state": { "id": 8, "code": "RJ", "name": "Rajasthan" },
  "district": { "id": 123, "name": "Jaipur" },
  "facts": {
    "age": 45,
    "gender": "female",
    "marital_status": "widowed",
    "income_annual": 120000,
    "land_acres": 2.0,
    "occupation": "farmer",
    "residence_type": "rural"
  },
  "documents": [
    { "type": "aadhaar", "has": true },
    { "type": "bank_account", "has": true },
    { "type": "income_cert", "has": false }
  ],
  "completed": true,
  "updated_at": "2026-10-05T12:00:00Z"
}
```

## Conversation

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/conversations` | Bearer | Start new conversation |
| `POST` | `/api/v1/conversations/{id}/message` | Bearer | Send text message |
| `POST` | `/api/v1/conversations/{id}/voice` | Bearer | Send voice message (multipart) |
| `GET` | `/api/v1/conversations/{id}` | Bearer | Get conversation state |

### `POST /api/v1/conversations/{id}/voice`

```
Content-Type: multipart/form-data
- audio: (binary, WebM/OGG, max 5MB)
- language: "hi" (optional, auto-detected if omitted)
```

```json
// Response 200
{
  "transcript": "मैं विधवा हूँ, मेरे पास 2 एकड़ ज़मीन है",
  "detected_language": "hi",
  "response_text": "समझ गई। आपकी उम्र क्या है?",
  "response_audio_url": "/api/v1/audio/temp/uuid.wav",
  "facts_extracted": {
    "marital_status": "widowed",
    "land_acres": 2.0
  },
  "profile_snapshot": { ... },
  "follow_up": {
    "question_about": "age",
    "is_last_question": false
  }
}
```

## Eligibility

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/eligibility/evaluate` | Bearer | Evaluate profile against schemes |
| `GET` | `/api/v1/eligibility/results` | Bearer | Get cached evaluation results |

### `POST /api/v1/eligibility/evaluate`

```json
// Request (empty - uses current profile)
{}
// Response 200
{
  "evaluated_at": "2026-10-05T12:00:00Z",
  "profile_completeness": 0.85,
  "coverage_note": "Full scheme coverage for Rajasthan",
  "results": [
    {
      "scheme": {
        "id": "uuid",
        "name": "Widow Pension Scheme",
        "benefit_type": "pension",
        "benefit_amount": "₹1,500/month",
        "ministry": "Social Justice",
        "scope": "state",
        "state": "Rajasthan",
        "official_url": "https://...",
        "application_url": "https://...",
        "deadline": null,
        "verified_at": "2026-10-01T00:00:00Z"
      },
      "status": "eligible",
      "match_score": 0.95,
      "rule_results": [
        { "rule": "Age >= 40", "status": "pass" },
        { "rule": "Gender is female", "status": "pass" },
        { "rule": "Marital status is widowed", "status": "pass" },
        { "rule": "Income <= ₹2,00,000", "status": "pass" }
      ],
      "explanation": "आप इस योजना के लिए पात्र हैं क्योंकि आप 40 वर्ष से अधिक आयु की विधवा महिला हैं और आपकी वार्षिक आय ₹2,00,000 से कम है।",
      "documents": {
        "available": ["aadhaar", "bank_account"],
        "missing": ["income_cert", "death_cert"],
        "all_present": false
      }
    },
    {
      "scheme": { ... },
      "status": "potentially_eligible",
      "missing_fields": ["caste_category"],
      "explanation": "We need your caste/category information to determine eligibility for this scheme."
    }
  ],
  "summary": {
    "eligible": 3,
    "potentially_eligible": 5,
    "not_eligible_count": 12,
    "total_evaluated": 20
  }
}
```

## Schemes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/schemes/{id}` | Bearer | Get scheme details |
| `GET` | `/api/v1/schemes/{id}/documents` | Bearer | Get scheme document requirements with user status |

## Applications

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/applications` | Bearer | Create application draft |
| `GET` | `/api/v1/applications/{id}` | Bearer | Get application details |
| `PUT` | `/api/v1/applications/{id}/form` | Bearer | Update form data |
| `POST` | `/api/v1/applications/{id}/pdf` | Bearer | Generate PDF |
| `GET` | `/api/v1/applications/{id}/pdf` | Bearer | Download PDF |
| `GET` | `/api/v1/applications` | Bearer | List user's applications |

## Notifications

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/notifications/preferences` | Bearer | Get notification preferences |
| `PUT` | `/api/v1/notifications/preferences` | Bearer | Update preferences |
| `GET` | `/api/v1/notifications` | Bearer | List notifications |

## Helper Mode

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/helper/dashboard` | Helper | Get helper dashboard |
| `POST` | `/api/v1/helper/sessions` | Helper | Start citizen session |
| `GET` | `/api/v1/helper/sessions/{id}` | Helper | Get session details |
| `POST` | `/api/v1/helper/sessions/{id}/consent` | Helper | Record consent |
| `POST` | `/api/v1/helper/sessions/{id}/end` | Helper | End session |
| `GET` | `/api/v1/helper/sessions` | Helper | List today's sessions |

## Admin (Scheme Management)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/admin/schemes` | Admin | List all schemes (any status) |
| `POST` | `/api/v1/admin/schemes` | Admin | Create scheme |
| `PUT` | `/api/v1/admin/schemes/{id}` | Admin | Update scheme |
| `POST` | `/api/v1/admin/schemes/{id}/publish` | Admin | Publish scheme |
| `POST` | `/api/v1/admin/schemes/{id}/archive` | Admin | Archive scheme |
| `GET` | `/api/v1/admin/candidates` | Admin | List candidate schemes |
| `POST` | `/api/v1/admin/candidates/{id}/accept` | Admin | Accept candidate |
| `POST` | `/api/v1/admin/candidates/{id}/reject` | Admin | Reject candidate |
| `GET` | `/api/v1/admin/sources` | Admin | List sources |
| `POST` | `/api/v1/admin/sources` | Admin | Create source |
| `GET` | `/api/v1/admin/audit` | Admin | Query audit log |

---

# 41. Core User Flow Diagrams

## Primary User Journey

```
┌─────────────────────────────────────────────────────────┐
│                    USER OPENS SAHAYAK                     │
│                          │                               │
│                    ┌─────▼─────┐                         │
│                    │ Language   │                         │
│                    │ Selection  │                         │
│                    └─────┬─────┘                         │
│                          │                               │
│              ┌───────────▼───────────┐                   │
│              │ "Tell me about        │                   │
│              │  yourself" (voice)    │                   │
│              └───────────┬───────────┘                   │
│                          │                               │
│     ┌────────────────────▼────────────────────┐          │
│     │  STT → LLM Fact Extraction → Profile    │          │
│     └────────────────────┬────────────────────┘          │
│                          │                               │
│              ┌───────────▼───────────┐                   │
│              │  Missing facts?       │──Yes──►  Follow-up│
│              │                       │           question │
│              └───────────┬───────────┘         (loop)    │
│                     No   │                               │
│              ┌───────────▼───────────┐                   │
│              │  Profile Review       │                   │
│              │  (user confirms)      │                   │
│              └───────────┬───────────┘                   │
│                          │                               │
│     ┌────────────────────▼────────────────────┐          │
│     │  Eligibility Engine Evaluation          │          │
│     │  (deterministic rules)                  │          │
│     └────────────────────┬────────────────────┘          │
│                          │                               │
│              ┌───────────▼───────────┐                   │
│              │  Eligible Schemes     │                   │
│              │  + Document Check     │                   │
│              └───────────┬───────────┘                   │
│                          │                               │
│              ┌───────────▼───────────┐                   │
│              │  User selects scheme  │                   │
│              └───────────┬───────────┘                   │
│                          │                               │
│     ┌────────────────────▼────────────────────┐          │
│     │  Auto-filled Application Form           │          │
│     │  (user reviews/edits)                   │          │
│     └────────────────────┬────────────────────┘          │
│                          │                               │
│              ┌───────────▼───────────┐                   │
│              │  Generate PDF         │                   │
│              │  Download/Print       │                   │
│              └───────────┬───────────┘                   │
│                          │                               │
│              ┌───────────▼───────────┐                   │
│              │  Apply via official   │                   │
│              │  channel              │                   │
│              └───────────┬───────────┘                   │
│                          │                               │
│              ┌───────────▼───────────┐                   │
│              │  Track + Reminders    │                   │
│              │  (WhatsApp/SMS)       │                   │
│              └───────────────────────┘                   │
└─────────────────────────────────────────────────────────┘
```

---

# 42. Eligibility Engine Design

(See Section 10 for full details)

## Summary Architecture

```
Profile (JSONB) → Flatten → EvaluationContext (typed struct)
                              │
                              ▼
         For each published scheme:
            Load eligibility rules (expr expressions)
            Evaluate each rule: PASS / FAIL / UNKNOWN
                              │
                              ▼
         Classify: eligible / potentially_eligible / not_eligible
                              │
                              ▼
         Rank by: exact eligibility > partial > benefit amount
                              │
                              ▼
         Return results with per-rule explanations
```

## Expression Examples

| Scheme | Expression |
|--------|-----------|
| Widow Pension | `age >= 40 && gender == "female" && marital_status == "widowed" && income_annual <= 200000` |
| PM-KISAN | `occupation == "farmer" && land_acres <= 5.0` |
| PM-JAY (health) | `ration_card_type in ["bpl", "aay"]` |
| SC/ST Scholarship | `caste_category in ["sc", "st"] && has_student_child == true` |
| PMAY-G (housing) | `residence_type == "rural" && income_annual <= 300000 && !has_pucca_house` |

---

# 43. Scheme Source/Discovery Architecture

(See Sections 12-14 for full details)

```
┌──────────────────────────────┐
│       Source Registry         │
│  (official sources only)      │
└──────────────┬───────────────┘
               │
    ┌──────────▼──────────┐
    │  Scheduled Fetcher   │
    │  (River job, weekly)  │
    │  Rate-limited, cached │
    └──────────┬──────────┘
               │
    ┌──────────▼──────────┐
    │  Candidate Schemes   │
    │  (unverified staging) │
    └──────────┬──────────┘
               │
    ┌──────────▼──────────┐
    │  Human Review Queue  │
    │  (admin panel)        │
    └──────────┬──────────┘
               │
    ┌──────────▼──────────┐
    │  Draft → Verified    │
    │  → Published Scheme  │
    └─────────────────────┘
```

---

# 44. Voice/LLM Architecture

(See Sections 8-9 for full details)

```
┌─────────┐     audio      ┌──────────┐     text       ┌─────────┐
│ Frontend │ ──────────────►│ Backend  │ ──────────────►│   LLM   │
│ (mic)    │                │ (Go)     │                │ Provider │
└─────────┘                │          │                └────┬────┘
                           │          │◄───── facts JSON ───┘
                           │          │
                           │   ┌──────▼──────┐
                           │   │   Bhashini   │
                           │   │  ASR / TTS   │
                           │   └─────────────┘
                           │          │
                           │◄─── transcript ──┘
                           │
                           │ generate response text
                           │
                           │──► Bhashini TTS ──► audio response
                           │
                           └──────────────────► Frontend (text + audio)
```

---

# 45. Notification Architecture

(See Section 23 for full details)

```
┌──────────────┐     enqueue     ┌──────────────┐
│ Trigger Event │ ──────────────►│ River Queue   │
│ (deadline,    │                │ (PostgreSQL)  │
│  status, etc) │                └──────┬───────┘
└──────────────┘                       │
                                ┌──────▼───────┐
                                │   Worker      │
                                │  (Go routine) │
                                └──────┬───────┘
                                       │
                         ┌─────────────┼─────────────┐
                         │                           │
                  ┌──────▼──────┐             ┌─────▼──────┐
                  │  WhatsApp   │             │    SMS     │
                  │  (Gupshup)  │             │  (MSG91)   │
                  └──────┬──────┘             └─────┬──────┘
                         │                           │
                  ┌──────▼───────────────────────────▼──────┐
                  │        Delivery Status Update           │
                  │      (notifications table update)       │
                  └────────────────────────────────────────┘
```

---

# 46. Helper Mode Architecture

(See Section 24 for full details)

```
┌─────────────────────────────────────────────────┐
│                HELPER MODE                       │
│                                                  │
│  ┌──────────────┐                               │
│  │ Helper Login  │ (OTP + verified helper)        │
│  └──────┬───────┘                               │
│         │                                        │
│  ┌──────▼───────┐                               │
│  │  Dashboard    │ (today's sessions, stats)      │
│  └──────┬───────┘                               │
│         │                                        │
│  ┌──────▼───────┐                               │
│  │ New Session   │                               │
│  │ + Consent     │ (mandatory before proceeding) │
│  └──────┬───────┘                               │
│         │                                        │
│  ┌──────▼───────┐                               │
│  │ Same intake   │ (voice/text for citizen)       │
│  │ flow as       │                               │
│  │ citizen mode  │                               │
│  └──────┬───────┘                               │
│         │                                        │
│  ┌──────▼───────┐                               │
│  │ Results +     │ (helper explains to citizen)   │
│  │ Application   │                               │
│  └──────┬───────┘                               │
│         │                                        │
│  ┌──────▼───────┐                               │
│  │ End Session   │ (audit record, data sealed)    │
│  └──────────────┘                               │
│                                                  │
│  ┌──────────────────────────────────────────┐   │
│  │ CONSTRAINTS:                              │   │
│  │ - Cannot access past session data         │   │
│  │ - Cannot export citizen data              │   │
│  │ - All actions audit-logged                │   │
│  │ - Rate limited (max 50 sessions/day)      │   │
│  │ - Cannot bypass eligibility rules         │   │
│  └──────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
```

---

# 47. Deployment Architecture

```
┌─────────────────────────────────────────────────────┐
│                    OCI VPS                            │
│                                                      │
│  ┌──────────────────────────────────────────────┐   │
│  │                 Coolify                        │   │
│  │                                                │   │
│  │  ┌──────────┐  ┌──────────┐  ┌────────────┐  │   │
│  │  │ Traefik  │  │ Sahayak  │  │ PostgreSQL │  │   │
│  │  │ (HTTPS)  │──│ API+Worker│──│   16       │  │   │
│  │  │          │  │ (Go)     │  │            │  │   │
│  │  └──────────┘  └──────────┘  └─────┬──────┘  │   │
│  │        │                            │         │   │
│  │  ┌─────▼──────┐              ┌─────▼──────┐  │   │
│  │  │ Frontend   │              │ PG Volumes │  │   │
│  │  │ (SvelteKit)│              │ (persistent)│  │   │
│  │  └────────────┘              └────────────┘  │   │
│  │                                                │   │
│  │  ┌────────────┐                               │   │
│  │  │ Backup Cron│──► /backups (host volume)      │   │
│  │  └────────────┘                               │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  DNS: sahayak.example.com → VPS IP                   │
│  HTTPS: Let's Encrypt via Traefik                    │
└─────────────────────────────────────────────────────┘
```

---

# 48. Development Milestones

## Milestone Order (Dependency-Correct)

### M0: Foundation (Week 1-2)

**Deliverables:**
- Go project scaffolding with module structure
- PostgreSQL setup with Docker Compose
- Migration framework with initial schemas (users, states, districts)
- Configuration system (env vars)
- HTTP server with chi router
- Health check endpoint
- Basic middleware (logging, request ID, CORS)
- CI pipeline (build, test, lint)
- LGD data seeding (all 36 states/UTs + districts)

**Definition of Done:**
- `go build` succeeds
- `go test ./...` passes
- Docker Compose starts API + PostgreSQL
- `GET /healthz` returns 200
- All 36 states/UTs seeded in database
- CI pipeline green

---

### M1: Authentication (Week 2-3)

**Deliverables:**
- Phone OTP auth flow (request, verify)
- JWT access + refresh tokens
- Auth middleware
- Role-based access control (citizen, helper, admin)
- OTP rate limiting
- SMS provider integration (MSG91)

**Definition of Done:**
- User can register and login via OTP
- JWT tokens issued and validated
- Refresh token flow works
- Rate limiting prevents OTP abuse
- Admin users can be created via seed script

**Dependencies:** M0

---

### M2: Profile & Location (Week 3-4)

**Deliverables:**
- Profile CRUD API
- Profile facts JSONB model
- Document inventory model
- Location API (states, districts, coverage)
- Coverage status display

**Definition of Done:**
- User can create/update profile via API
- Profile stores structured facts
- Document inventory tracks has/doesn't have
- Location API returns all states with coverage info
- Profile validation works

**Dependencies:** M0, M1

---

### M3: Scheme Data Model & Admin (Week 4-6)

**Deliverables:**
- Scheme CRUD (admin API)
- Eligibility rules storage (expr expressions)
- Required documents per scheme
- Application fields per scheme
- Source registry
- Candidate scheme pipeline
- Verification workflow
- Initial scheme data curation (30+ central schemes)

**Definition of Done:**
- Admin can create/edit/publish schemes
- Each scheme has structured eligibility rules
- Each scheme has required documents
- 30+ central schemes loaded with verified data
- Candidate scheme acceptance/rejection works

**Dependencies:** M0, M1

---

### M4: Eligibility Engine (Week 6-7)

**Deliverables:**
- `expr-lang/expr` integration
- Profile flattening to evaluation context
- Three-state evaluation (PASS/FAIL/UNKNOWN)
- Scheme ranking
- Eligibility API endpoint
- Extensive rule tests

**Definition of Done:**
- Eligibility evaluation returns correct results for all 30+ schemes
- Missing data returns UNKNOWN, not FAIL
- Every rule has at least 3 test cases
- Results include per-rule explanations
- Performance: evaluate 30 schemes in < 100ms

**Dependencies:** M2, M3

---

### M5: LLM Integration (Week 7-9)

**Deliverables:**
- LLM provider interface
- Gemini implementation (primary)
- OpenAI implementation (fallback)
- Fact extraction from natural language
- Missing fact detection
- Follow-up question generation
- Grounded explanation generation
- System prompt engineering
- LLM response validation (reject hallucinated data)

**Definition of Done:**
- LLM correctly extracts facts from 20+ test inputs in Hindi and English
- Follow-up questions generated naturally in correct language
- Grounded explanations use only verified scheme data
- LLM does NOT decide eligibility
- Provider can be swapped via config

**Dependencies:** M2, M3

---

### M6: Voice Integration (Week 9-11)

**Deliverables:**
- Bhashini API integration (ASR + TTS + NMT)
- Voice upload endpoint
- Audio processing pipeline (in-memory, no storage)
- STT provider interface + Bhashini implementation
- TTS provider interface + Bhashini implementation
- Translation provider interface + Bhashini NMT implementation
- Fallback to Google Cloud Speech/TTS
- Language detection

**Definition of Done:**
- Audio uploaded → transcribed to text via Bhashini
- Text response → audio via Bhashini TTS
- Works for Hindi and English
- At least 2 additional Indian languages tested
- Fallback works when Bhashini is unavailable
- Audio is NOT stored (verified by testing)

**Dependencies:** M5, (Bhashini approval needed — start in M0)

---

### M7: Conversation Flow (Week 11-13)

**Deliverables:**
- Conversation state machine (intake flow)
- Voice/text message handling
- STT → LLM fact extraction → profile update pipeline
- Follow-up question loop
- Profile review and confirmation
- Automatic eligibility evaluation on profile completion
- Conversation API endpoints
- Max turn limit with form fallback

**Definition of Done:**
- User can complete full intake via voice in Hindi
- User can complete full intake via text in English
- Profile progressively builds from conversation
- Follow-up questions are contextually appropriate
- Profile review screen shows accurate data
- Eligibility results shown after confirmation

**Dependencies:** M4, M5, M6

---

### M8: Document System (Week 13-14)

**Deliverables:**
- Document requirement matching (scheme vs. profile)
- Missing document detection
- Document type reference data (acquisition guidance)
- Service location data (initial pilot: 1-2 states)
- Document checklist API

**Definition of Done:**
- For each eligible scheme, correctly shows ✓/❌ documents
- Missing documents include acquisition guidance
- Guidance is sourced from verified data (not LLM-generated)
- Nearest service locations shown for pilot states

**Dependencies:** M2, M3, M4

---

### M9: Application Forms & PDF (Week 14-16)

**Deliverables:**
- Application form generation from scheme fields + profile
- Auto-fill logic
- Form review API
- maroto v2 PDF generation with Hindi font support
- PDF download endpoint
- Application CRUD

**Definition of Done:**
- Application form auto-fills from profile data
- User can edit fields before finalization
- PDF generates correctly with Hindi text
- PDF includes disclaimer about unofficial status
- PDF stored temporarily and downloadable

**Dependencies:** M2, M3, M8

---

### M10: Frontend PWA (Week 12-18, parallel)

**Deliverables:**
- SvelteKit project setup
- PWA configuration (manifest, service worker, icons)
- API client (typed, from OpenAPI spec)
- i18n setup (Hindi + English)
- Auth flow UI (phone OTP)
- Voice input component (MediaRecorder)
- Chat/conversation UI
- Profile display and editing
- Scheme results display
- Document checklist UI
- Application form UI
- Language selector
- Location picker
- Responsive mobile-first design

**Definition of Done:**
- Installable PWA from Chrome on Android
- Voice input works on mobile browsers
- Full user journey completable in Hindi and English
- Lighthouse PWA score > 90
- Works on 3G (LCP < 4s)
- All data from backend API

**Dependencies:** M1-M9 (API endpoints), can start parallel from M3

---

### M11: Notifications (Week 18-20)

**Deliverables:**
- River job queue setup for notifications
- WhatsApp BSP integration (Gupshup)
- SMS provider integration (MSG91 for notifications)
- Notification templates (Hindi + English)
- Notification preferences API
- Deadline reminders (scheduled jobs)
- Opt-in/opt-out flow
- DLT template registration

**Definition of Done:**
- Deadline reminders sent via WhatsApp 7 days before deadline
- SMS fallback works when WhatsApp fails
- User can opt in/out of notifications
- Notification delivery status tracked
- DLT compliance verified

**Dependencies:** M1, M9 (applications exist)

---

### M12: Helper Mode (Week 20-22)

**Deliverables:**
- Helper profile management
- Helper session flow
- Consent mechanism
- Helper dashboard
- Session isolation (can't access past session data)
- Enhanced audit logging for helper actions
- Helper frontend views

**Definition of Done:**
- Helper can start session, record consent, assist citizen
- Citizen data inaccessible after session ends
- All helper actions appear in audit log
- Rate limiting enforced (50 sessions/day)
- Helper cannot bypass eligibility rules

**Dependencies:** M1, M2, M7, M10

---

### M13: Status Tracking (Week 22-23)

**Deliverables:**
- Manual status tracking (user enters reference number)
- Status update notifications
- Application status tracking UI
- Scheduled status check jobs (for supported schemes)

**Definition of Done:**
- User can record application reference number
- Status changes trigger notifications
- UI shows application status timeline
- No false claims about submission or approval status

**Dependencies:** M9, M11

---

### M14: Production Hardening (Week 23-25)

**Deliverables:**
- Security audit (OWASP checklist)
- Penetration testing (basic)
- Load testing (target: 100 concurrent users)
- Database backup automation
- Monitoring dashboards (Prometheus + Grafana or Coolify built-in)
- Log aggregation
- Error alerting
- Rate limiting tuning
- Performance optimization
- Data retention job (cleanup old conversations, expired OTPs)
- Privacy review

**Definition of Done:**
- No critical security vulnerabilities
- Handles 100 concurrent users without degradation
- Backups run daily and are tested
- Monitoring alerts on error rate spikes
- Data retention policies enforced
- Ready for production deployment

**Dependencies:** All previous milestones

---

### M15: Launch (Week 25-26)

**Deliverables:**
- Production deployment on Coolify
- DNS + HTTPS configured
- Smoke tests pass on production
- Initial scheme data verified
- User acceptance testing
- Documentation (README, API docs, admin guide)

**Definition of Done:**
- Sahayak accessible at production domain
- Full user journey works on mobile
- 30+ schemes with verified eligibility rules
- At least 2 states with state-specific scheme coverage
- Monitoring active
- Backups verified

---

# 49. Technical Risks and Mitigations

| # | Risk | Severity | Likelihood | Mitigation |
|---|------|----------|------------|------------|
| 1 | **Bhashini API unavailability** | High | Medium | Fallback to Google Cloud Speech/TTS. Circuit breaker prevents cascading failures. Text-only mode always available. |
| 2 | **LLM hallucination in fact extraction** | High | Medium | Validate LLM output against known fact schema. Reject any facts with unknown keys. Never trust LLM for eligibility decisions. User reviews all extracted facts. |
| 3 | **Scheme data staleness** | High | High | Last-verified timestamps displayed to user. Weekly link checks. Monthly admin review cycle. Clear "verified as of" dates. |
| 4 | **LLM API cost overrun** | Medium | Medium | Per-user daily limits. Caching for repeated queries. Use cheaper models (Gemini Flash) for extraction. Rate limiting. Budget alerts. |
| 5 | **No public MyScheme API** | High | Certain | Manual scheme curation from official sources. NOT scraping. Start with small, high-impact set. Build admin tools for efficient data entry. |
| 6 | **Voice recognition accuracy for dialects** | Medium | High | Support primary languages well. Dialect speakers fall back to closest recognized language. Text input always available. User can correct transcription. |
| 7 | **WhatsApp Business verification delay** | Low | Medium | Start verification process early (M0). SMS as primary channel initially. |
| 8 | **Privacy/data protection compliance** | High | Low | Data minimization by design. No Aadhaar storage. Phone hashing. Consent mechanisms. Audit logging. Regular privacy reviews. |
| 9 | **expr rule expression errors** | Medium | Medium | All rules validated on save. Mandatory test cases for each rule. Sandbox evaluation prevents arbitrary code execution. |
| 10 | **Single VPS failure** | High | Low | Daily backups to separate storage. Coolify supports quick redeployment. Database restore tested monthly. |
| 11 | **Bhashini approval delay** | Medium | Medium | Apply immediately in M0. Use Google Cloud Speech/TTS as initial provider, swap to Bhashini when approved. Architecture is provider-agnostic. |
| 12 | **Regional language content quality** | Medium | Medium | Start with Hindi + English (strong support). Add regional languages incrementally. Use Bhashini NMT for translation, cache results. Human review of translated scheme content. |
| 13 | **PDF Hindi font rendering** | Low | Medium | Use Noto Sans Devanagari fonts. Test PDF output early in M9. Alternative: HTML-to-PDF via wkhtmltopdf gives better typography control. |

---

# 50. External Services and Alternatives

| Service | Primary | Alternatives | Notes |
|---------|---------|-------------|-------|
| **STT (Speech-to-Text)** | Bhashini ASR | Google Cloud Speech-to-Text, Azure Speech, Whisper API | Bhashini is free and has best Indian language coverage |
| **TTS (Text-to-Speech)** | Bhashini TTS | Google Cloud TTS, Azure TTS | Bhashini has Rajasthani and Bhojpuri TTS |
| **Translation** | Bhashini NMT | Google Translate API, Azure Translator | Bhashini designed for Indian languages |
| **LLM (Extraction/NLG)** | Google Gemini (Flash) | OpenAI GPT-4o-mini, Anthropic Claude Haiku, Groq (Llama) | Gemini: good Hindi, generous free tier. OpenAI: best overall. Groq: fastest/cheapest |
| **SMS** | MSG91 | Kaleyra, Twilio, Textlocal, Amazon SNS | MSG91: Indian company, DLT support, OTP specialization |
| **WhatsApp** | Gupshup | MSG91, Twilio, AiSensy, Interakt | Gupshup: established in India, good API |
| **Database** | PostgreSQL 16 | — | No alternative recommended. PG is the right choice. |
| **Job Queue** | River (PG-backed) | Asynq (Redis), Machinery (Redis) | River avoids Redis dependency |
| **PDF** | maroto v2 | wkhtmltopdf, unipdf (commercial) | maroto for structured; wkhtmltopdf for complex layouts |
| **Hosting** | OCI VPS + Coolify | DigitalOcean, Hetzner, AWS Lightsail | OCI has generous free tier |
| **Maps** | OpenStreetMap + Leaflet | Google Maps API | OSM is free, no API key needed for basic display |
| **Rules Engine** | expr-lang/expr | google/cel-go, custom | expr is simpler; CEL is more robust for complex rules |
| **Monitoring** | Prometheus + Grafana | Coolify built-in, Uptime Kuma | Start with Coolify built-in, add Prometheus if needed |

---

# Verification Plan

## Automated Tests

```bash
# Backend tests (unit + integration)
go test -v -race -coverprofile=coverage.out ./...

# Frontend tests
cd frontend && pnpm test

# Frontend type checking
cd frontend && pnpm check

# Linting
golangci-lint run

# Build verification
go build ./cmd/server
cd frontend && pnpm build

# Database migration test
# (uses testcontainers to start PG, run all migrations up, then all down)
go test -v ./migrations/...
```

### Manual Verification

After each milestone:
1. Run the full Docker Compose stack locally
2. Execute the core user journey manually
3. Verify voice input works on mobile browser
4. Check scheme eligibility results against manual calculation
5. Verify audit log entries for all actions
6. Test rate limiting by exceeding thresholds
7. Verify no PII in application logs

---

---

# 51. Configuration and Local Development

## Environment Variables

The application is configured entirely via environment variables (12-factor app).

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | API Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/sahayak?sslmode=disable` |
| `JWT_SECRET` | Secret key for signing JWTs | (32+ character random string) |
| `PHONE_HASH_PEPPER` | Secret pepper for hashing phone numbers | (32+ character random string) |
| `VOICE_PROVIDER` | Active STT/TTS provider | `sarvam` (default) or `bhashini` |
| `SARVAM_API_KEY` | API Key for Sarvam AI | `sk-...` |
| `LLM_PROVIDER` | Active LLM provider | `gemini` (default) |
| `GEMINI_API_KEY` | API key for Google Gemini | `AIza...` |
| `SMS_PROVIDER` | Active SMS provider | `msg91` |
| `MSG91_AUTH_KEY` | MSG91 Authentication key | `...` |
| `WHATSAPP_API_KEY` | Gupshup/WhatsApp API key | `...` |
| `SERVER_MODE` | Mode to run the binary | `all` (api + worker), `api`, or `worker` |

## Local Development Setup

To go from an empty repository to a running application, the implementation agent will:

1. **Database:** Run `docker-compose up db -d` to start PostgreSQL locally.
2. **Migrations:** Run the migration tool to create tables: `go run cmd/migrate/main.go up`
3. **Seed Data:** Run the seed script to populate LGD locations and document types: `go run cmd/seed/main.go`
4. **Backend:** Start the Go server: `go run cmd/server/main.go`
5. **Frontend:** Navigate to `frontend/`, run `pnpm install` and `pnpm dev`.

---

# 52. Background Jobs (River) Detailed Architecture

The system utilizes `riverqueue/river` embedded within the Go monolith, backed by PostgreSQL. No Redis is required.

## Job Types

1. **`NotificationJob`**: Dispatches WhatsApp or SMS messages.
2. **`StatusCheckJob`**: Periodically checks external application statuses (where supported).
3. **`LinkCheckJob`**: Weekly sweep to ensure scheme application URLs return 200 OK.
4. **`DataRetentionJob`**: Cleans up expired OTPs and old temporary PDF files.

## Behavior and Reliability

*   **Idempotency:** All job processing logic MUST be idempotent. If a `NotificationJob` runs twice due to a crash, it should check the `notifications` table status before sending.
*   **Transactional Enqueueing:** Jobs are enqueued within the same PostgreSQL transaction as the business logic. If a user completes an application, the DB commit and the notification job enqueue happen atomically.
*   **Retry and Failure:** Jobs configure exponential backoff. Example: A failed SMS delivery will retry 3 times (e.g., +1m, +5m, +15m). After max retries, the job is marked `discarded` and triggers an alert.
*   **Cleanup:** River automatically sweeps and deletes successfully completed jobs after a configured retention period (e.g., 24 hours) to prevent DB bloat.


**End of Implementation Plan.** Please review the "User Review Required" and "Open Questions" sections at the top and provide your approval to proceed.
