-- Migration 08: Date Instance Proposals
-- Stores proposed times from receiver for date scheduling negotiation
-- See: lib/scheduling/specs/product/06d_tier3_time_flow.md

--------------------------------------------------------------------------------
-- DATE INSTANCE PROPOSAL
-- Stores proposed scheduled times suggested by the receiver.
-- Each proposal can be: pending, accepted, rejected, or superseded.
--------------------------------------------------------------------------------

CREATE TABLE date_instance_proposal
(
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_instance_ref_id  UUID        NOT NULL REFERENCES date_instance (id) ON DELETE CASCADE,
    suggested_by_ref_id   UUID        NOT NULL REFERENCES users (id),
    proposed_time         TIMESTAMPTZ NOT NULL,
    status                TEXT        NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'rejected', 'superseded')),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMPTZ,

    -- Prevent duplicate times for the same date instance
    UNIQUE (date_instance_ref_id, proposed_time)
);

-- Index for querying proposals by date instance
CREATE INDEX idx_date_instance_proposal_date_instance
    ON date_instance_proposal (date_instance_ref_id);

-- Partial index for pending proposals (most common query)
CREATE INDEX idx_date_instance_proposal_pending
    ON date_instance_proposal (date_instance_ref_id, status)
    WHERE status = 'pending';

-- Index for querying by suggester
CREATE INDEX idx_date_instance_proposal_suggester
    ON date_instance_proposal (suggested_by_ref_id);

COMMENT ON TABLE date_instance_proposal IS
    'Stores proposed scheduled times suggested by the receiver for a date instance';

COMMENT ON COLUMN date_instance_proposal.status IS
    'pending=awaiting selection, accepted=chosen by initiator, rejected=not chosen, superseded=replaced by new proposals';

COMMENT ON COLUMN date_instance_proposal.proposed_time IS
    'The proposed date/time in UTC. Must be in the future when created.';
