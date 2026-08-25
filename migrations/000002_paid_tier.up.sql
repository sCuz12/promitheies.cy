CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    email           TEXT NOT NULL UNIQUE,
    tier            TEXT NOT NULL DEFAULT 'free',
    telegram_id     BIGINT,
    trial_ends_at   TIMESTAMPTZ,
    stripe_customer_id TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE magic_link_tokens (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL UNIQUE,
    expires_at      TIMESTAMPTZ NOT NULL,
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_magic_link_tokens_user_id ON magic_link_tokens(user_id);

CREATE TABLE user_alerts (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    cpv_codes       TEXT[] NOT NULL DEFAULT '{}',
    keywords        TEXT[] NOT NULL DEFAULT '{}',
    min_value       NUMERIC(14,2),
    max_value       NUMERIC(14,2),
    authority_ids   BIGINT[] NOT NULL DEFAULT '{}',
    frequency       TEXT NOT NULL DEFAULT 'realtime',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_user_alerts_user_id ON user_alerts(user_id);

CREATE TABLE alert_log (
    id              BIGSERIAL PRIMARY KEY,
    user_alert_id   BIGINT NOT NULL REFERENCES user_alerts(id) ON DELETE CASCADE,
    tender_id       BIGINT NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
    channel         TEXT NOT NULL,
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_alert_id, tender_id, channel)
);
CREATE INDEX idx_alert_log_user_alert_id ON alert_log(user_alert_id);

CREATE TABLE watchlist (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contractor_id   BIGINT NOT NULL REFERENCES contractors(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, contractor_id)
);

CREATE TABLE saved_searches (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    query_json      JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_saved_searches_user_id ON saved_searches(user_id);
