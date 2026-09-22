CREATE TABLE newsletter_subscribers (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT NOT NULL,
    language      TEXT NOT NULL DEFAULT 'el' CHECK (language IN ('el', 'en')),
    consented_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (email = lower(email)),
    CHECK (length(email) BETWEEN 3 AND 254)
);

CREATE UNIQUE INDEX idx_newsletter_subscribers_email
    ON newsletter_subscribers (email);
