CREATE TABLE schemes (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                    TEXT UNIQUE NOT NULL,
    status                  TEXT NOT NULL DEFAULT 'draft',
    name_en                 TEXT NOT NULL,
    name_hi                 TEXT NOT NULL,
    description_en          TEXT,
    description_hi          TEXT,
    benefit_type            TEXT NOT NULL,
    benefit_amount          TEXT,
    benefit_description_en  TEXT,
    benefit_description_hi  TEXT,
    scope                   TEXT NOT NULL DEFAULT 'central',
    state_id                INT REFERENCES states(id),
    ministry                TEXT,
    department              TEXT,
    application_url         TEXT,
    official_url            TEXT,
    deadline                DATE,
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    completeness            JSONB NOT NULL DEFAULT '{}',
    verification            JSONB NOT NULL DEFAULT '{}',
    source_id               UUID REFERENCES sources(id),
    verified_at             TIMESTAMPTZ,
    verified_by             UUID REFERENCES users(id),
    published_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_schemes_status ON schemes(status);
CREATE INDEX idx_schemes_state ON schemes(state_id);
CREATE INDEX idx_schemes_scope ON schemes(scope);

CREATE TABLE scheme_eligibility_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scheme_id       UUID NOT NULL REFERENCES schemes(id) ON DELETE CASCADE,
    rule_name       TEXT NOT NULL,
    expression      TEXT NOT NULL,
    description_en  TEXT,
    description_hi  TEXT,
    is_mandatory    BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(scheme_id, rule_name)
);

CREATE INDEX idx_eligibility_rules_scheme ON scheme_eligibility_rules(scheme_id);

CREATE TABLE scheme_required_documents (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scheme_id        UUID NOT NULL REFERENCES schemes(id) ON DELETE CASCADE,
    document_type    TEXT NOT NULL,
    document_name_en TEXT NOT NULL,
    document_name_hi TEXT NOT NULL,
    is_mandatory     BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order       INT NOT NULL DEFAULT 0,
    UNIQUE(scheme_id, document_type)
);

CREATE TABLE scheme_application_fields (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scheme_id       UUID NOT NULL REFERENCES schemes(id) ON DELETE CASCADE,
    field_name      TEXT NOT NULL,
    field_label_en  TEXT NOT NULL,
    field_label_hi  TEXT NOT NULL,
    field_type      TEXT NOT NULL DEFAULT 'text',
    profile_mapping TEXT,
    is_required     BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INT NOT NULL DEFAULT 0,
    options         JSONB,
    UNIQUE(scheme_id, field_name)
);

ALTER TABLE candidate_schemes
    ADD CONSTRAINT fk_candidate_promoted_to FOREIGN KEY (promoted_to) REFERENCES schemes(id);
