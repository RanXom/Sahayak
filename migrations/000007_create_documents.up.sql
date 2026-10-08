CREATE TABLE document_types (
    id                   TEXT PRIMARY KEY,
    name_en              TEXT NOT NULL,
    name_hi              TEXT NOT NULL,
    description_en       TEXT NOT NULL DEFAULT '',
    description_hi       TEXT NOT NULL DEFAULT '',
    category             TEXT NOT NULL DEFAULT 'identity',
    issuing_authority_en TEXT,
    issuing_authority_hi TEXT,
    acquisition_url      TEXT,
    acquisition_steps_en TEXT,
    acquisition_steps_hi TEXT,
    typical_time         TEXT,
    typical_fee          TEXT,
    last_verified        TIMESTAMPTZ,
    source_url           TEXT
);

CREATE TABLE service_locations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_type   TEXT NOT NULL,
    name            TEXT NOT NULL,
    address         TEXT,
    state_id        INT REFERENCES states(id),
    district_id     INT REFERENCES districts(id),
    latitude        DECIMAL(10, 8),
    longitude       DECIMAL(11, 8),
    phone           TEXT,
    services        TEXT[],
    active          BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_service_locations_state ON service_locations(state_id);
CREATE INDEX idx_service_locations_district ON service_locations(district_id);
