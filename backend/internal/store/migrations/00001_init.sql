-- Initial schema for the meetup app.
--
-- Conventions: the app generates UUIDs; the DB owns timestamps
-- (DEFAULT now(), read back via RETURNING). DB constraints cover closed value

-- +goose Up

CREATE TABLE users (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL CHECK (name <> ''),
    label      TEXT NOT NULL DEFAULT '',
    -- No avatar column: avatars live at the deterministic object key
    -- avatar/{user_id} in R2, with the client falling back to avatar/default
    -- on 404. Reads need no DB field or presigned URL.
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE events (
    id             UUID PRIMARY KEY,
    name           TEXT NOT NULL CHECK (name <> ''),
    type           TEXT NOT NULL CHECK (type <> ''),
    starts_at      TIMESTAMPTZ,                    -- NULL until scheduled/confirmed
    lat            DOUBLE PRECISION NOT NULL CHECK (lat BETWEEN -90 AND 90),
    lon            DOUBLE PRECISION NOT NULL CHECK (lon BETWEEN -180 AND 180),
    location_label TEXT NOT NULL DEFAULT '',
    created_by     UUID NOT NULL REFERENCES users (id),   -- ON DELETE RESTRICT (default)
    status         TEXT NOT NULL CHECK (status IN ('draft', 'polling', 'scheduled', 'past')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Event.Validate lifecycle invariant: a date exists iff scheduled/past;
    -- polling must not carry one; draft may or may not.
    CONSTRAINT event_startsat_matches_status CHECK (
        (status IN ('scheduled', 'past') AND starts_at IS NOT NULL) OR
        (status = 'polling'              AND starts_at IS NULL)     OR
        (status = 'draft')
    )
);

CREATE INDEX idx_events_status ON events (status);
CREATE INDEX idx_events_created_by ON events (created_by);

CREATE TABLE device_tokens (
    token      TEXT PRIMARY KEY,   -- FCM tokens are globally unique
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    platform   TEXT NOT NULL CHECK (platform IN ('android', 'web')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_device_tokens_user_id ON device_tokens (user_id);

CREATE TABLE subscriptions (
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    event_id   UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    status     TEXT NOT NULL CHECK (status IN ('going', 'maybe', 'declined')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, event_id)
);

CREATE INDEX idx_subscriptions_event_id ON subscriptions (event_id);

CREATE TABLE messages (
    id         UUID PRIMARY KEY,
    event_id   UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id),   -- RESTRICT: keep the author
    kind       TEXT NOT NULL CHECK (kind IN ('text', 'image', 'poll')),
    body       TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_messages_event_created_at ON messages (event_id, created_at);

CREATE TABLE attachments (
    id         UUID PRIMARY KEY,
    message_id UUID NOT NULL UNIQUE REFERENCES messages (id) ON DELETE CASCADE,
    object_key TEXT NOT NULL CHECK (object_key <> ''),
    mime_type  TEXT NOT NULL CHECK (mime_type LIKE 'image/%')
);

CREATE TABLE polls (
    id                  UUID PRIMARY KEY,
    message_id          UUID NOT NULL UNIQUE REFERENCES messages (id) ON DELETE CASCADE,
    question            TEXT NOT NULL CHECK (question <> ''),
    closes_at           TIMESTAMPTZ,   -- NULL: open until confirmed
    confirmed_option_id UUID           -- NULL until confirmed; FK added after poll_options
);

CREATE TABLE poll_options (
    id            UUID PRIMARY KEY,
    poll_id       UUID NOT NULL REFERENCES polls (id) ON DELETE CASCADE,
    proposed_date TIMESTAMPTZ,           -- set on date polls, NULL on opinion polls
    label         TEXT NOT NULL DEFAULT '',

    -- PollOption.Validate: exactly one of proposed_date / label.
    CONSTRAINT poll_option_date_xor_label CHECK (
        (proposed_date IS NOT NULL AND label = '') OR
        (proposed_date IS NULL     AND label <> '')
    )
);

CREATE INDEX idx_poll_options_poll_id ON poll_options (poll_id);

-- Break the polls <-> poll_options cycle: this FK can only be added once
-- poll_options exists. Not deferrable — confirmed_option_id is NULL at creation
-- and set later by UPDATE, by which point the referenced option already exists.
ALTER TABLE polls
    ADD CONSTRAINT fk_polls_confirmed_option
    FOREIGN KEY (confirmed_option_id) REFERENCES poll_options (id);

CREATE TABLE poll_votes (
    option_id  UUID NOT NULL REFERENCES poll_options (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (option_id, user_id)   -- one vote per option per user; multi-option allowed
);

-- +goose Down

-- Drop the cyclic FK first so polls and poll_options detach, then drop tables in
-- reverse dependency order. Indexes and constraints drop with their tables.
ALTER TABLE polls DROP CONSTRAINT IF EXISTS fk_polls_confirmed_option;

DROP TABLE IF EXISTS poll_votes;
DROP TABLE IF EXISTS poll_options;
DROP TABLE IF EXISTS polls;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS device_tokens;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS users;
