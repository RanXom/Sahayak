CREATE TABLE sources (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              TEXT NOT NULL,
    source_type       TEXT NOT NULL,
    url               TEXT,
    description       TEXT,
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    fetch_config      JSONB,
    last_fetched_at   TIMESTAMPTZ,
    last_fetch_status TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE candidate_schemes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id       UUID NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    external_id     TEXT,
    raw_data        JSONB NOT NULL,
    name            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    content_hash    TEXT NOT NULL,
    fetched_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at     TIMESTAMPTZ,
    reviewed_by     UUID REFERENCES users(id),
    reviewer_notes  TEXT,
    promoted_to     UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_id, content_hash)
);

CREATE INDEX idx_candidate_schemes_status ON candidate_schemes(status);
