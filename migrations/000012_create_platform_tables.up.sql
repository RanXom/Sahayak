CREATE TABLE translation_cache (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_lang  TEXT NOT NULL,
    target_lang  TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    source_text  TEXT NOT NULL,
    translated   TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_lang, target_lang, content_hash)
);

CREATE TABLE llm_usage (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID,
    provider   TEXT NOT NULL,
    model      TEXT NOT NULL,
    operation  TEXT NOT NULL,
    tokens_in  INT NOT NULL DEFAULT 0,
    tokens_out INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_llm_usage_user_created ON llm_usage(user_id, created_at);
