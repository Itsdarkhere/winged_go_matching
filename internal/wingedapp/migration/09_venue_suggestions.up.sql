-- Migration 09: Venue Suggestions
-- Stores venue suggestions from receiver ("I want to suggest a place myself")
-- See: lib/scheduling/specs/product/06_endpoints_v1_build_order.md (Tier 4)

--------------------------------------------------------------------------------
-- VENUE SUGGESTION
-- Stores venue suggestions from the receiver during venue negotiation.
-- When receiver uses "I want to suggest a place myself", they can submit a
-- Google Maps link or venue name + area, which the agent resolves to a venue.
--------------------------------------------------------------------------------

CREATE TABLE venue_suggestion
(
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_instance_ref_id  UUID        NOT NULL REFERENCES date_instance (id) ON DELETE CASCADE,
    suggested_by_ref_id   UUID        NOT NULL REFERENCES users (id),

    -- Input from user (one of these will be populated)
    venue_link            TEXT,        -- Google Maps link or similar
    venue_name            TEXT,        -- Venue name typed by user
    venue_area            TEXT,        -- Area/neighborhood typed by user

    -- Resolved venue (populated after agent processing)
    resolved_venue_ref_id UUID        REFERENCES venue (id),
    resolution_status     TEXT        NOT NULL DEFAULT 'pending'
        CHECK (resolution_status IN ('pending', 'resolved', 'failed')),
    resolution_error      TEXT,        -- Error message if resolution failed

    -- Response from initiator
    initiator_response    TEXT        -- 'accepted', 'show_alternatives', NULL (pending)
        CHECK (initiator_response IS NULL OR initiator_response IN ('accepted', 'show_alternatives')),
    responded_at          TIMESTAMPTZ,

    -- Meta
    created_at            TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMPTZ
);

-- Index for querying suggestions by date instance
CREATE INDEX idx_venue_suggestion_date_instance
    ON venue_suggestion (date_instance_ref_id);

-- Index for pending suggestions awaiting response
CREATE INDEX idx_venue_suggestion_pending_response
    ON venue_suggestion (date_instance_ref_id)
    WHERE initiator_response IS NULL AND resolution_status = 'resolved';

COMMENT ON TABLE venue_suggestion IS
    'Stores venue suggestions from the receiver during venue negotiation flow';

COMMENT ON COLUMN venue_suggestion.venue_link IS
    'Google Maps link or similar URL provided by receiver';

COMMENT ON COLUMN venue_suggestion.venue_name IS
    'Venue name typed by receiver (used with venue_area for resolution)';

COMMENT ON COLUMN venue_suggestion.resolution_status IS
    'pending=not yet processed, resolved=venue found, failed=could not resolve';

COMMENT ON COLUMN venue_suggestion.initiator_response IS
    'accepted=initiator accepted suggestion, show_alternatives=initiator wants similar venues';
