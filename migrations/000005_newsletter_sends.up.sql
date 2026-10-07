CREATE TABLE newsletter_sends (
    id              BIGSERIAL PRIMARY KEY,
    subscriber_id   BIGINT NOT NULL REFERENCES newsletter_subscribers(id) ON DELETE CASCADE,
    tender_id       BIGINT NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (subscriber_id, tender_id)
);

CREATE INDEX idx_newsletter_sends_subscriber_id ON newsletter_sends(subscriber_id);
