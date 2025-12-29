-- Migration 05: Matching System

--------------------------------------------------------------------------------
-- MATCH SET
--------------------------------------------------------------------------------

CREATE TABLE match_set
(
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                   VARCHAR(256) NOT NULL UNIQUE,
    number_of_participants INTEGER      NOT NULL,
    match_configuration    JSONB        NOT NULL,
    time_start             TIMESTAMPTZ,
    time_end               TIMESTAMPTZ,
    created_at             TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMPTZ
);

--------------------------------------------------------------------------------
-- MATCH RESULT
--------------------------------------------------------------------------------

CREATE TABLE match_result
(
    id                           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),

    -- refs
    match_set_ref_id             UUID        NOT NULL REFERENCES match_set (id),
    user_a_ref_id                UUID        NOT NULL REFERENCES users (id),
    user_b_ref_id                UUID        NOT NULL REFERENCES users (id),
    initiator_user_ref_id        UUID        NOT NULL REFERENCES users (id),
    receiver_user_ref_id         UUID        NOT NULL REFERENCES users (id),

    -- status (string enums instead of FK)
    match_status                 VARCHAR(32) NOT NULL DEFAULT 'Active'
        CHECK (match_status IN ('Active', 'Expired')),
    match_lifecycle_status       VARCHAR(64)
        CHECK (match_lifecycle_status IS NULL OR match_lifecycle_status IN (
            'Confirmed', 'Scheduling', 'Date Set', 'Date Complete Pending Feedback',
            'Decision Pending Window', 'Queued', 'Closed'
        )),
    current_date_instance_id     UUID,  -- FK added in 07_dating after date_instance exists

    -- per-user actions (string enums)
    user_a_action                VARCHAR(32) NOT NULL DEFAULT 'Pending'
        CHECK (user_a_action IN ('Pending', 'Proposed', 'Passed')),
    user_a_action_at             TIMESTAMPTZ,
    user_a_seen_at               TIMESTAMPTZ,

    user_b_action                VARCHAR(32) NOT NULL DEFAULT 'Pending'
        CHECK (user_b_action IN ('Pending', 'Proposed', 'Passed')),
    user_b_action_at             TIMESTAMPTZ,
    user_b_seen_at               TIMESTAMPTZ,

    -- domain: matching
    qualifier_results            JSONB,
    matched_qualitatively        BOOLEAN     NOT NULL DEFAULT FALSE,
    delivered_to_user_at         TIMESTAMPTZ,
    last_proposer_user_ref_id    UUID,
    last_proposed_at             TIMESTAMPTZ,
    chat_unlocked_at             TIMESTAMPTZ,
    is_approved                  BOOLEAN     NOT NULL DEFAULT FALSE,
    is_dropped                   BOOLEAN     NOT NULL DEFAULT FALSE,
    dropped_ts                   TIMESTAMPTZ,
    is_possible_match            BOOLEAN     NOT NULL DEFAULT FALSE,
    is_expired                   BOOLEAN     NOT NULL DEFAULT FALSE,

    -- expiration
    expires_at                   TIMESTAMPTZ,

    -- meta
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                   TIMESTAMPTZ,

    UNIQUE (match_set_ref_id, user_a_ref_id, user_b_ref_id)
);

CREATE INDEX idx_match_result_user_a_dropped ON match_result (user_a_ref_id, is_dropped, is_expired) WHERE is_approved = TRUE;
CREATE INDEX idx_match_result_user_b_dropped ON match_result (user_b_ref_id, is_dropped, is_expired) WHERE is_approved = TRUE;

--------------------------------------------------------------------------------
-- MATCH CONFIG
--------------------------------------------------------------------------------

CREATE TABLE match_config
(
    id                          UUID PRIMARY KEY       DEFAULT gen_random_uuid(),
    age_range_start             INTEGER                DEFAULT -5,
    age_range_end               INTEGER                DEFAULT 5,
    age_range_woman_older_by    INTEGER       NOT NULL DEFAULT 5,
    age_range_man_older_by      INTEGER       NOT NULL DEFAULT 10,
    height_male_greater_by_cm   DECIMAL(5, 2) NOT NULL DEFAULT 1,
    location_radius_km          DECIMAL(5, 2) NOT NULL DEFAULT 200,
    location_adaptive_expansion INTEGER[]     NOT NULL DEFAULT ARRAY [200, 350, 500],
    match_hours                 TEXT[]        NOT NULL DEFAULT ARRAY ['00:00'],
    drop_hours                  TEXT[]        NOT NULL DEFAULT ARRAY ['19:00', '20:00', '23:00', '22:00'],
    drop_hours_utc              TEXT[]        NOT NULL DEFAULT ARRAY ['GMT+3'],
    stale_chat_nudge            INTEGER       NOT NULL DEFAULT 24,
    stale_chat_agent_setup      INTEGER       NOT NULL DEFAULT 84,
    match_expiration_hours      INTEGER       NOT NULL DEFAULT 72,
    match_block_declined        INTEGER       NOT NULL DEFAULT 168,
    match_block_ignored         INTEGER       NOT NULL DEFAULT 168,
    match_block_closed          INTEGER       NOT NULL DEFAULT 168,
    score_range_start           DECIMAL(5, 2) NOT NULL DEFAULT 0.52,
    score_range_end             DECIMAL(5, 2) NOT NULL DEFAULT 0.60
);

INSERT INTO match_config (age_range_start, age_range_end, age_range_woman_older_by, age_range_man_older_by,
                          height_male_greater_by_cm, location_radius_km, location_adaptive_expansion,
                          drop_hours, drop_hours_utc, stale_chat_nudge, stale_chat_agent_setup,
                          match_expiration_hours, match_block_declined, match_block_ignored,
                          match_block_closed, score_range_start, score_range_end)
VALUES (-5, 5, 5, 10, 1, 200, ARRAY [200, 350, 500],
        ARRAY ['19:00', '20:00', '23:00', '22:00'], ARRAY ['GMT+3'],
        24, 84, 72, 168, 168, 168, 0.52, 0.60);

--------------------------------------------------------------------------------
-- MATCH CHAT MESSAGE
--------------------------------------------------------------------------------

CREATE TABLE match_chat_message
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    match_result_id UUID        NOT NULL REFERENCES match_result (id) ON DELETE CASCADE,
    sender_id       UUID        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    message         TEXT        NOT NULL,
    seen_at         TIMESTAMPTZ,
    edited_at       TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ
);

CREATE INDEX idx_chat_message_match_created ON match_chat_message (match_result_id, created_at);
CREATE INDEX idx_chat_message_unseen ON match_chat_message (match_result_id, sender_id) WHERE seen_at IS NULL AND deleted_at IS NULL;
CREATE INDEX idx_chat_message_sender ON match_chat_message (sender_id, created_at DESC);
