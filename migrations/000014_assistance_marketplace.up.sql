-- ============================================================================
-- Client-helper assistance marketplace: request queue, sessions, chat,
-- signaling, shared drafts and consent ledger.
-- ============================================================================

-- Constrain account roles to the known set (previously convention only).
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('citizen', 'helper', 'admin'));

-- Helper profile: languages, categories, availability, terms.
ALTER TABLE helper_profiles ADD COLUMN IF NOT EXISTS display_name TEXT;
ALTER TABLE helper_profiles ADD COLUMN IF NOT EXISTS languages TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE helper_profiles ADD COLUMN IF NOT EXISTS categories TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE helper_profiles ADD COLUMN IF NOT EXISTS experience TEXT;
ALTER TABLE helper_profiles ADD COLUMN IF NOT EXISTS is_available BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE helper_profiles ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMPTZ;

-- Assistance requests created by citizens.
CREATE TABLE IF NOT EXISTS assistance_requests (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    citizen_user_id      UUID NOT NULL REFERENCES users(id),
    application_id       UUID REFERENCES applications(id),
    scheme_id            UUID REFERENCES schemes(id),
    language             TEXT NOT NULL,
    category             TEXT NOT NULL,
    details              TEXT NOT NULL DEFAULT '',
    share_profile_facts  BOOLEAN NOT NULL DEFAULT TRUE,
    share_documents      BOOLEAN NOT NULL DEFAULT FALSE,
    status               TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'in_session', 'completed', 'cancelled', 'expired')),
    matched_helper_user_id UUID REFERENCES users(id),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    matched_at           TIMESTAMPTZ,
    closed_at            TIMESTAMPTZ,
    expires_at           TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours'
);
CREATE INDEX IF NOT EXISTS idx_assistance_requests_citizen ON assistance_requests(citizen_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_assistance_requests_queue ON assistance_requests(status, language, created_at)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_assistance_requests_helper ON assistance_requests(matched_helper_user_id, status);

-- Link helper sessions to the request that created them.
ALTER TABLE helper_sessions ADD COLUMN IF NOT EXISTS request_id UUID REFERENCES assistance_requests(id);
ALTER TABLE helper_sessions ADD COLUMN IF NOT EXISTS ended_by TEXT;

-- Session chat messages. client_msg_id gives idempotent sends.
CREATE TABLE IF NOT EXISTS session_messages (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id     UUID NOT NULL REFERENCES helper_sessions(id) ON DELETE CASCADE,
    sender_user_id UUID NOT NULL REFERENCES users(id),
    sender_role    TEXT NOT NULL CHECK (sender_role IN ('citizen', 'helper')),
    client_msg_id  TEXT,
    body           TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 2000),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_session_messages_client ON session_messages(session_id, client_msg_id)
    WHERE client_msg_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_session_messages_session ON session_messages(session_id, created_at, id);

-- WebRTC signaling messages (offer/answer/ice/hangup/join/leave).
CREATE TABLE IF NOT EXISTS session_signals (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id     UUID NOT NULL REFERENCES helper_sessions(id) ON DELETE CASCADE,
    sender_user_id UUID NOT NULL REFERENCES users(id),
    kind           TEXT NOT NULL CHECK (kind IN ('offer', 'answer', 'ice', 'hangup', 'join', 'leave')),
    payload        JSONB NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_session_signals_session ON session_signals(session_id, created_at, id);

-- Shared draft values. Helper proposes, client confirms; confirmed values
-- are applied to the application form. Last write wins per field with
-- author/timestamp visible to both parties.
CREATE TABLE IF NOT EXISTS session_draft_values (
    session_id   UUID NOT NULL REFERENCES helper_sessions(id) ON DELETE CASCADE,
    field_name   TEXT NOT NULL,
    value        JSONB,
    proposed_by  UUID NOT NULL REFERENCES users(id),
    status       TEXT NOT NULL DEFAULT 'proposed'
        CHECK (status IN ('proposed', 'confirmed', 'applied')),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (session_id, field_name)
);

-- Consent ledger: one row per grant, revoked by default on session end.
CREATE TABLE IF NOT EXISTS consent_records (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id   UUID NOT NULL REFERENCES helper_sessions(id) ON DELETE CASCADE,
    request_id   UUID NOT NULL REFERENCES assistance_requests(id),
    granted_by   UUID NOT NULL REFERENCES users(id),
    scope        JSONB NOT NULL,
    statement    TEXT NOT NULL,
    ip           TEXT NOT NULL DEFAULT '',
    user_agent   TEXT NOT NULL DEFAULT '',
    valid        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_consent_session ON consent_records(session_id, valid);
