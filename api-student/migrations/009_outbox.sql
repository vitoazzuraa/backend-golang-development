CREATE TABLE IF NOT EXISTS outbox (
    id           BIGSERIAL   PRIMARY KEY,
    event_type   VARCHAR(50) NOT NULL,
    payload      JSONB       NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS outbox_unpublished_idx
    ON outbox (id) WHERE published_at IS NULL;

CREATE TABLE IF NOT EXISTS notifications (
    id         BIGSERIAL   PRIMARY KEY,
    outbox_id  BIGINT      NOT NULL REFERENCES outbox(id) ON DELETE CASCADE,
    user_id    INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject    VARCHAR(200) NOT NULL,
    body       TEXT        NOT NULL,
    sent_at    TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS notifications_outbox_idx ON notifications (outbox_id);
CREATE INDEX IF NOT EXISTS notifications_user_idx ON notifications (user_id);
