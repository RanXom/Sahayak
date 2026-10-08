CREATE TABLE states (
    id              INT PRIMARY KEY,
    code            TEXT UNIQUE NOT NULL,
    name_en         TEXT NOT NULL,
    name_hi         TEXT NOT NULL,
    name_local      TEXT,
    is_ut           BOOLEAN NOT NULL DEFAULT FALSE,
    active          BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE districts (
    id              INT PRIMARY KEY,
    state_id        INT NOT NULL REFERENCES states(id),
    name_en         TEXT NOT NULL,
    name_hi         TEXT NOT NULL,
    name_local      TEXT,
    census_code     TEXT,
    active          BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_districts_state ON districts(state_id);

CREATE TABLE coverage_status (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state_id        INT NOT NULL REFERENCES states(id),
    level           TEXT NOT NULL DEFAULT 'none',
    scheme_count    INT NOT NULL DEFAULT 0,
    last_updated    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes           TEXT,
    UNIQUE(state_id)
);
