-- Migration 04: Wings Economy System

--------------------------------------------------------------------------------
-- SUBSCRIPTION PLANS
--------------------------------------------------------------------------------

CREATE TABLE wings_ecn_subscription_plan
(
    id               UUID PRIMARY KEY   DEFAULT uuid_generate_v4(),
    subscription_type VARCHAR(64)       NOT NULL
        CHECK (subscription_type IN ('Winged+', 'WingedX')),
    name             VARCHAR(255)   NOT NULL,
    price            NUMERIC(10, 2) NOT NULL,
    wings            INTEGER        NOT NULL,
    is_active        INTEGER            DEFAULT 1,
    created_by       INTEGER            DEFAULT NULL,
    created_date     TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    last_updated     TIMESTAMP          DEFAULT NULL,
    updated_by       INTEGER            DEFAULT NULL,
    CONSTRAINT wings_ecn_subscription_plan_unique UNIQUE (subscription_type, name)
);

-- Winged+ Plans
INSERT INTO wings_ecn_subscription_plan (subscription_type, name, price, wings)
VALUES ('Winged+', 'Weekly', 14.95, 25),
       ('Winged+', 'Monthly', 24.96, 55),
       ('Winged+', '3 Months', 55.32, 180),
       ('Winged+', '6 Months', 92.16, 360);

-- WingedX Plans
INSERT INTO wings_ecn_subscription_plan (subscription_type, name, price, wings)
VALUES ('WingedX', 'Weekly', 14.95, 25),
       ('WingedX', 'Monthly', 24.96, 55),
       ('WingedX', '3 Months', 55.32, 180),
       ('WingedX', '6 Months', 92.16, 360);

--------------------------------------------------------------------------------
-- USER SUBSCRIPTION PLANS
--------------------------------------------------------------------------------

CREATE TABLE wings_ecn_user_subscription_plan
(
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id              UUID      NOT NULL REFERENCES users (id),
    subscription_plan_id UUID      NOT NULL REFERENCES wings_ecn_subscription_plan (id),
    start_date           TIMESTAMP NOT NULL,
    end_date             TIMESTAMP NOT NULL,
    is_active            INTEGER          DEFAULT 1,
    created_date         TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    last_updated         TIMESTAMP        DEFAULT NULL
);

--------------------------------------------------------------------------------
-- ACTION LOG
--------------------------------------------------------------------------------

CREATE TABLE wings_ecn_action_log
(
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_ref_id       UUID        NOT NULL REFERENCES users (id),
    action_log_type   VARCHAR(64) NOT NULL
        CHECK (action_log_type IN (
            'Daily Check-In', 'Send Message',
            'WingedX - Weekly Payment', 'WingedX - Monthly Payment',
            'Winged+ - Weekly Payment', 'Winged+ - Monthly Payment',
            'Winged+ - 3 Month Payment', 'Winged+ - 6 Month Payment',
            'Referral - Friend Signup', 'Referral - Friend Complete',
            'Streak - 7 Day Milestone', 'Streak - 30 Day Milestone',
            'Attend a Date'
        )),
    ext_domain_ref_id UUID        NOT NULL,
    is_credit         BOOLEAN     NOT NULL,
    extra_info        JSONB            DEFAULT NULL,
    is_active         INTEGER          DEFAULT 1,
    created_by        INTEGER          DEFAULT NULL,
    created_date      TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    last_updated      TIMESTAMP        DEFAULT NULL,
    updated_by        INTEGER          DEFAULT NULL
);

CREATE INDEX idx_wings_ecn_action_log_ext_domain_ref_id ON wings_ecn_action_log (ext_domain_ref_id);

--------------------------------------------------------------------------------
-- TRANSACTION LEDGER
--------------------------------------------------------------------------------

CREATE TABLE wings_ecn_transaction
(
    id                UUID PRIMARY KEY     DEFAULT uuid_generate_v4(),
    action_log_type   VARCHAR(64) NOT NULL
        CHECK (action_log_type IN (
            'Daily Check-In', 'Send Message',
            'WingedX - Weekly Payment', 'WingedX - Monthly Payment',
            'Winged+ - Weekly Payment', 'Winged+ - Monthly Payment',
            'Winged+ - 3 Month Payment', 'Winged+ - 6 Month Payment',
            'Referral - Friend Signup', 'Referral - Friend Complete',
            'Streak - 7 Day Milestone', 'Streak - 30 Day Milestone',
            'Attend a Date'
        )),
    user_ref_id       UUID        NOT NULL,
    action_log_ref_id UUID        NOT NULL REFERENCES wings_ecn_action_log (id),
    is_credit         BOOLEAN     NOT NULL,
    claimed           BOOLEAN     NOT NULL DEFAULT FALSE,
    amount            INTEGER     NOT NULL,
    expires_at        TIMESTAMPTZ          DEFAULT NULL,
    is_expired        BOOLEAN     NOT NULL DEFAULT FALSE,
    extra_info        JSONB                DEFAULT NULL,
    is_active         INTEGER              DEFAULT 1,
    created_by        INTEGER              DEFAULT NULL,
    created_date      TIMESTAMP            DEFAULT CURRENT_TIMESTAMP,
    last_updated      TIMESTAMP            DEFAULT NULL,
    updated_by        INTEGER              DEFAULT NULL
);

-- Index for efficient expiry queries (cron job)
CREATE INDEX idx_wings_ecn_transaction_expires_at
    ON wings_ecn_transaction (expires_at)
    WHERE expires_at IS NOT NULL AND is_expired = FALSE AND is_active = 1;

--------------------------------------------------------------------------------
-- USER TOTALS
--------------------------------------------------------------------------------

CREATE TABLE wings_ecn_user_totals
(
    id                     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_ref_id            UUID    NOT NULL UNIQUE REFERENCES users (id),
    total_wings            INTEGER NOT NULL DEFAULT 0,
    counter_sent_messages  INTEGER NOT NULL DEFAULT 0,
    counter_daily_check_in INTEGER NOT NULL DEFAULT 0,
    premium_expires_in     TIMESTAMPTZ      DEFAULT NULL,
    is_active              INTEGER          DEFAULT 1,
    created_by             INTEGER          DEFAULT NULL,
    created_date           TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    last_updated           TIMESTAMP        DEFAULT NULL,
    updated_by             INTEGER          DEFAULT NULL
);
