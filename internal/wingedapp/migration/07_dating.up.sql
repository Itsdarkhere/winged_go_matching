-- Migration 07: Dating System
-- Redesigned for scheduling flow with string enums

--------------------------------------------------------------------------------
-- VENUE (cached from external APIs)
--------------------------------------------------------------------------------

CREATE TABLE venue
(
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- External source
    external_provider VARCHAR(64)  NOT NULL,
    external_id       VARCHAR(255) NOT NULL,

    -- Display
    name              VARCHAR(255) NOT NULL,
    display_name      VARCHAR(255),
    address           TEXT,

    -- Location
    latitude          DECIMAL(10, 8),
    longitude         DECIMAL(11, 8),

    -- Metadata
    venue_data        JSONB,
    dietary_tags      TEXT[],
    date_type_fit     TEXT[],

    -- Places API fields
    rating            DECIMAL(2, 1),
    price_level       INTEGER,
    user_rating_count INTEGER,
    google_maps_url   TEXT,
    website_url       TEXT,
    open_now          BOOLEAN,

    -- LLM ranking
    sort_order        INTEGER,

    -- Venue Intelligence
    bucket_for_venue  VARCHAR(32),
    subtype_for_venue VARCHAR(64),
    primary_type      VARCHAR(128),
    is_chain          BOOLEAN DEFAULT FALSE,
    noise_level       VARCHAR(32) DEFAULT 'unknown',

    -- Food/drink service flags
    serves_breakfast  BOOLEAN,
    serves_brunch     BOOLEAN,
    serves_lunch      BOOLEAN,
    serves_dinner     BOOLEAN,
    serves_beer       BOOLEAN,
    serves_wine       BOOLEAN,
    serves_cocktails  BOOLEAN,
    serves_coffee     BOOLEAN,

    -- LLM-generated text
    overview_text     TEXT,
    description_text  TEXT,
    review_summary    TEXT,
    sample_reviews    TEXT[],

    -- Cache control
    cached_at         TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    refresh_after     TIMESTAMPTZ,

    UNIQUE (external_provider, external_id)
);

CREATE INDEX idx_venue_location ON venue (latitude, longitude);
CREATE INDEX idx_venue_dietary ON venue USING GIN (dietary_tags);
CREATE INDEX idx_venue_date_type ON venue USING GIN (date_type_fit);
CREATE INDEX idx_venue_sort_order ON venue (sort_order) WHERE sort_order IS NOT NULL;

--------------------------------------------------------------------------------
-- DATE INSTANCE
--------------------------------------------------------------------------------

CREATE TABLE date_instance
(
    id                             UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Relations
    match_result_ref_id            UUID        NOT NULL REFERENCES match_result (id),
    venue_ref_id                   UUID        REFERENCES venue (id),

    -- Scheduling (string enums)
    date_type_core                 VARCHAR(32)
        CHECK (date_type_core IS NULL OR date_type_core IN ('coffee', 'drinks', 'meal', 'walk', 'activity')),
    status                         VARCHAR(64) NOT NULL DEFAULT 'Proposed'
        CHECK (status IN (
            'Proposed', 'Time Chosen', 'Venue Chosen', 'Date Set',
            'Completed', 'Cancelled', 'Expired', 'No Show'
        )),
    scheduled_time_utc             TIMESTAMPTZ,
    duration_minutes               INTEGER,

    -- Booking
    booking_status                 VARCHAR(64) DEFAULT 'Unknown'
        CHECK (booking_status IS NULL OR booking_status IN (
            'Unknown', 'Booked', 'No Booking Needed', 'Booking Failed'
        )),

    -- Feedback: User A
    feedback_status_user_a         VARCHAR(64) DEFAULT 'Pending'
        CHECK (feedback_status_user_a IS NULL OR feedback_status_user_a IN (
            'Pending', 'Submitted', 'Auto Closed', 'Submitted By Agent'
        )),
    decision_user_a                VARCHAR(64)
        CHECK (decision_user_a IS NULL OR decision_user_a IN (
            'Schedule Second Date', 'Close Connection'
        )),
    did_meet_user_a                VARCHAR(64)
        CHECK (did_meet_user_a IS NULL OR did_meet_user_a IN (
            'Yes', 'No', 'Prefer Not To Say'
        )),
    feedback_text_user_a           TEXT,

    -- Feedback: User B
    feedback_status_user_b         VARCHAR(64) DEFAULT 'Pending'
        CHECK (feedback_status_user_b IS NULL OR feedback_status_user_b IN (
            'Pending', 'Submitted', 'Auto Closed', 'Submitted By Agent'
        )),
    decision_user_b                VARCHAR(64)
        CHECK (decision_user_b IS NULL OR decision_user_b IN (
            'Schedule Second Date', 'Close Connection'
        )),
    did_meet_user_b                VARCHAR(64)
        CHECK (did_meet_user_b IS NULL OR did_meet_user_b IN (
            'Yes', 'No', 'Prefer Not To Say'
        )),
    feedback_text_user_b           TEXT,

    -- Timer
    decision_window_end            TIMESTAMPTZ NOT NULL,

    -- UI State Support
    initiator_confirmed_at         TIMESTAMPTZ,
    receiver_confirmed_at          TIMESTAMPTZ,
    booking_failure_reason         VARCHAR(64)
        CHECK (booking_failure_reason IS NULL OR booking_failure_reason IN (
            'Fully Booked', 'No Answer', 'Closed', 'Other'
        )),
    venue_proposal_status          VARCHAR(32)
        CHECK (venue_proposal_status IS NULL OR venue_proposal_status IN (
            'Pending', 'Accepted', 'Rejected'
        )),
    availability_sync_mode         VARCHAR(32)
        CHECK (availability_sync_mode IS NULL OR availability_sync_mode IN (
            'Calendar', 'Manual'
        )),

    -- Meta
    created_at                     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                     TIMESTAMPTZ
);

CREATE INDEX idx_date_instance_match ON date_instance (match_result_ref_id);
CREATE INDEX idx_date_instance_status ON date_instance (status);
CREATE INDEX idx_date_instance_scheduled ON date_instance (scheduled_time_utc) WHERE scheduled_time_utc IS NOT NULL;

-- Add FK from match_result to date_instance (circular reference)
ALTER TABLE match_result
    ADD CONSTRAINT fk_match_result_current_date_instance
    FOREIGN KEY (current_date_instance_id) REFERENCES date_instance (id);

--------------------------------------------------------------------------------
-- DATE INSTANCE LOG (event sourcing)
--------------------------------------------------------------------------------

CREATE TABLE date_instance_log
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    date_instance_ref_id UUID        NOT NULL REFERENCES date_instance (id) ON DELETE CASCADE,
    user_ref_id          UUID        REFERENCES users (id) ON DELETE SET NULL,
    event_type           VARCHAR(64) NOT NULL,
    old_value            JSONB,
    new_value            JSONB,
    details              TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_date_instance_log_instance ON date_instance_log (date_instance_ref_id, created_at);

--------------------------------------------------------------------------------
-- NOTIFICATION
--------------------------------------------------------------------------------

CREATE TABLE notification
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_ref_id     UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    notification_type VARCHAR(64),
    title           VARCHAR(255),
    message         TEXT        NOT NULL,
    payload         JSONB,
    read_at         TIMESTAMPTZ,
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_user_unread ON notification (user_ref_id, created_at DESC) WHERE read_at IS NULL;

--------------------------------------------------------------------------------
-- VENUE RANKING CACHE (TasteBrain results)
--------------------------------------------------------------------------------

CREATE TABLE venue_ranking_cache
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_instance_ref_id UUID        NOT NULL REFERENCES date_instance (id) ON DELETE CASCADE,
    venue_ref_id         UUID        NOT NULL REFERENCES venue (id) ON DELETE CASCADE,
    rank_position        INTEGER     NOT NULL,
    rank_score           DECIMAL(5,4),
    user_a_profile_hash  VARCHAR(64),
    user_b_profile_hash  VARCHAR(64),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at           TIMESTAMPTZ,

    UNIQUE (date_instance_ref_id, venue_ref_id)
);

CREATE INDEX idx_venue_ranking_cache_date_instance
    ON venue_ranking_cache (date_instance_ref_id, rank_position);

--------------------------------------------------------------------------------
-- BOOKING REMINDER
--------------------------------------------------------------------------------

CREATE TABLE booking_reminder
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_instance_ref_id UUID        NOT NULL REFERENCES date_instance (id) ON DELETE CASCADE,
    status               VARCHAR(32) NOT NULL DEFAULT 'Pending'
        CHECK (status IN ('Pending', 'Fired', 'Dismissed', 'Completed')),
    remind_at            TIMESTAMPTZ NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fired_at             TIMESTAMPTZ
);

CREATE INDEX idx_booking_reminder_date_instance ON booking_reminder (date_instance_ref_id);
CREATE INDEX idx_booking_reminder_pending ON booking_reminder (remind_at) WHERE fired_at IS NULL;

--------------------------------------------------------------------------------
-- SCHEDULING CARD (UI state tracking)
--------------------------------------------------------------------------------

CREATE TABLE scheduling_card
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_instance_ref_id UUID        NOT NULL REFERENCES date_instance (id) ON DELETE CASCADE,
    user_ref_id          UUID        NOT NULL REFERENCES users (id),
    card_type            VARCHAR(64) NOT NULL
        CHECK (card_type IN (
            'Dietary Restrictions', 'Calendar Connect', 'Time Proposal', 'Time Confirmation',
            'Venue Proposal', 'Booking Confirmation', 'Date Booked', 'Venue Revision',
            'Feedback Request', 'Predate Reminder'
        )),
    card_state           VARCHAR(32) NOT NULL DEFAULT 'Pending'
        CHECK (card_state IN ('Pending', 'Completed', 'Expired')),
    payload              JSONB,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at         TIMESTAMPTZ,
    expired_at           TIMESTAMPTZ
);

CREATE INDEX idx_scheduling_card_instance_user ON scheduling_card (date_instance_ref_id, user_ref_id);
CREATE INDEX idx_scheduling_card_pending ON scheduling_card (user_ref_id, card_state) WHERE card_state = 'Pending';
