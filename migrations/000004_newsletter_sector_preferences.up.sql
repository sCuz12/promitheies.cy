ALTER TABLE newsletter_subscribers
    ADD COLUMN cpv_divisions TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN min_value    NUMERIC(14,2);
