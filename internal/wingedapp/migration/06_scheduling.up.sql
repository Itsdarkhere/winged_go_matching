-- Migration 06: Scheduling and Agent Log

--------------------------------------------------------------------------------
-- USER AVAILABILITY
--------------------------------------------------------------------------------

CREATE TABLE user_availability
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    time_block TSTZRANGE   NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT user_availability_no_overlap EXCLUDE USING gist (
        user_id WITH =,
        time_block WITH &&
        )
);

CREATE INDEX idx_user_availability_user_time ON user_availability USING gist (user_id, time_block);
CREATE INDEX idx_user_availability_time_block ON user_availability USING gist (time_block);

COMMENT ON TABLE user_availability IS 'Stores user availability time blocks for date scheduling';
COMMENT ON COLUMN user_availability.time_block IS 'Time range in format [start, end) - use tstzrange for overlap detection';

--------------------------------------------------------------------------------
-- AGENT LOG
--------------------------------------------------------------------------------

CREATE TABLE agent_log
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_ref_id UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    log         TEXT        NOT NULL,
    display_by  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    seen_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ
);

CREATE INDEX idx_agent_log_user_display ON agent_log (user_ref_id, display_by DESC);
CREATE INDEX idx_agent_log_unseen ON agent_log (user_ref_id) WHERE seen_at IS NULL;
CREATE INDEX idx_agent_log_displayable ON agent_log (display_by) WHERE seen_at IS NULL;
