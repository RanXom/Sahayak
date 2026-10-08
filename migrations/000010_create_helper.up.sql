CREATE TABLE helper_profiles (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    helper_type  TEXT NOT NULL,
    organization TEXT,
    id_number    TEXT,
    verified     BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at  TIMESTAMPTZ,
    verified_by  UUID REFERENCES users(id),
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE TABLE helper_sessions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    helper_user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    citizen_user_id    UUID REFERENCES users(id),
    citizen_phone_hash TEXT,
    consent_given      BOOLEAN NOT NULL DEFAULT FALSE,
    consent_text       TEXT,
    consent_at         TIMESTAMPTZ,
    purpose            TEXT NOT NULL,
    status             TEXT NOT NULL DEFAULT 'active',
    conversation_id    UUID REFERENCES conversations(id),
    profile_id         UUID REFERENCES profiles(id),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at       TIMESTAMPTZ
);

CREATE INDEX idx_helper_sessions_helper ON helper_sessions(helper_user_id);
CREATE INDEX idx_helper_sessions_status ON helper_sessions(status);
