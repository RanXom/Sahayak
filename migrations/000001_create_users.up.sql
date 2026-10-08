CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_hash      TEXT UNIQUE NOT NULL,
    phone_last4     TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'citizen',
    preferred_lang  TEXT NOT NULL DEFAULT 'hi',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ
);

CREATE TABLE otp_challenges (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_hash      TEXT NOT NULL,
    phone_last4     TEXT NOT NULL,
    otp_hash        TEXT NOT NULL,
    purpose         TEXT NOT NULL DEFAULT 'login',
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 3,
    expires_at      TIMESTAMPTZ NOT NULL,
    consumed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_challenges_phone_hash ON otp_challenges(phone_hash);

CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT UNIQUE NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
