CREATE TABLE applications (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id        UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    scheme_id         UUID NOT NULL REFERENCES schemes(id),
    status            TEXT NOT NULL DEFAULT 'draft',
    form_data         JSONB NOT NULL DEFAULT '{}',
    pdf_path          TEXT,
    submission_ref    TEXT,
    submitted_at      TIMESTAMPTZ,
    last_status_check TIMESTAMPTZ,
    external_status   TEXT,
    notes             TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_applications_user ON applications(user_id);
CREATE INDEX idx_applications_scheme ON applications(scheme_id);
