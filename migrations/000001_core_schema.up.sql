CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE authorities (
    id              BIGSERIAL PRIMARY KEY,
    canonical_name_en TEXT,
    canonical_name_el TEXT,
    type            TEXT NOT NULL DEFAULT '',
    region          TEXT NOT NULL DEFAULT '',
    slug            TEXT NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (canonical_name_en IS NOT NULL OR canonical_name_el IS NOT NULL)
);

CREATE TABLE authority_aliases (
    id              BIGSERIAL PRIMARY KEY,
    authority_id    BIGINT NOT NULL REFERENCES authorities(id) ON DELETE CASCADE,
    alias           TEXT NOT NULL,
    normalized_alias TEXT NOT NULL,
    source          TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (normalized_alias)
);
CREATE INDEX idx_authority_aliases_authority_id ON authority_aliases(authority_id);
CREATE INDEX idx_authority_aliases_trgm ON authority_aliases USING gin (normalized_alias gin_trgm_ops);

CREATE TABLE contractors (
    id              BIGSERIAL PRIMARY KEY,
    canonical_name  TEXT NOT NULL,
    normalized_name TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE contractor_aliases (
    id              BIGSERIAL PRIMARY KEY,
    contractor_id   BIGINT NOT NULL REFERENCES contractors(id) ON DELETE CASCADE,
    alias           TEXT NOT NULL,
    normalized_alias TEXT NOT NULL,
    source          TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (normalized_alias)
);
CREATE INDEX idx_contractor_aliases_contractor_id ON contractor_aliases(contractor_id);
CREATE INDEX idx_contractor_aliases_trgm ON contractor_aliases USING gin (normalized_alias gin_trgm_ops);

CREATE TABLE cpv_categories (
    code            TEXT PRIMARY KEY,
    description_en  TEXT NOT NULL,
    description_el  TEXT
);

CREATE TABLE tenders (
    id              BIGSERIAL PRIMARY KEY,
    external_ids    JSONB NOT NULL DEFAULT '{}',
    title_en        TEXT,
    title_el        TEXT,
    authority_id    BIGINT REFERENCES authorities(id),
    cpv_code        TEXT,
    cpv_division    TEXT, -- not FK'd: source CPV codes occasionally fall outside the seeded division list
    estimated_value NUMERIC(14,2),
    currency        TEXT NOT NULL DEFAULT 'EUR',
    deadline        DATE,
    status          TEXT NOT NULL DEFAULT 'unknown',
    procedure_type  TEXT,
    source          TEXT NOT NULL,
    published_at    DATE,
    raw_payload     JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_tenders_authority_id ON tenders(authority_id);
CREATE INDEX idx_tenders_cpv_code ON tenders(cpv_code);
CREATE INDEX idx_tenders_cpv_division ON tenders(cpv_division);
CREATE INDEX idx_tenders_status ON tenders(status);
CREATE INDEX idx_tenders_published_at ON tenders(published_at);
CREATE INDEX idx_tenders_external_ids ON tenders USING gin (external_ids);
-- Supports idempotent per-source upserts keyed on a source-specific external
-- id (e.g. data.gov.cy's CFTID) stored in external_ids.
CREATE UNIQUE INDEX idx_tenders_source_cftid ON tenders (source, (external_ids->>'cftid'))
    WHERE external_ids ? 'cftid';
-- Same, for TED's publication-number.
CREATE UNIQUE INDEX idx_tenders_source_ted_id ON tenders (source, (external_ids->>'ted'))
    WHERE external_ids ? 'ted';
CREATE INDEX idx_tenders_title_fts ON tenders USING gin (
    to_tsvector('simple', coalesce(title_en, '') || ' ' || coalesce(title_el, ''))
);

CREATE TABLE awards (
    id              BIGSERIAL PRIMARY KEY,
    tender_id       BIGINT NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
    contractor_id   BIGINT NOT NULL REFERENCES contractors(id),
    value           NUMERIC(14,2),
    award_date      DATE,
    source          TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tender_id, source)
);
CREATE INDEX idx_awards_tender_id ON awards(tender_id);
CREATE INDEX idx_awards_contractor_id ON awards(contractor_id);
CREATE INDEX idx_awards_award_date ON awards(award_date);

CREATE TABLE ingest_runs (
    id                  BIGSERIAL PRIMARY KEY,
    source              TEXT NOT NULL,
    started_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at         TIMESTAMPTZ,
    records_processed   INTEGER NOT NULL DEFAULT 0,
    records_inserted    INTEGER NOT NULL DEFAULT 0,
    status              TEXT NOT NULL DEFAULT 'running',
    error               TEXT
);
CREATE INDEX idx_ingest_runs_source ON ingest_runs(source, started_at DESC);
