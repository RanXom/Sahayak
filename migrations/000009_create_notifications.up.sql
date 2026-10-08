CREATE TABLE notification_preferences (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel      TEXT NOT NULL,
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    phone_hash   TEXT NOT NULL,
    opted_in_at  TIMESTAMPTZ,
    opted_out_at TIMESTAMPTZ,
    UNIQUE(user_id, channel)
);

CREATE TABLE notifications (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    application_id UUID REFERENCES applications(id) ON DELETE SET NULL,
    channel        TEXT NOT NULL,
    template_key   TEXT NOT NULL,
    content        TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'pending',
    provider_ref   TEXT,
    scheduled_for  TIMESTAMPTZ,
    sent_at        TIMESTAMPTZ,
    error          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_scheduled ON notifications(scheduled_for);
