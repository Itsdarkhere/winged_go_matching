-- Migration 11: Add streak tracking fields to user_totals
-- Per spec: Daily check-in does NOT grant wings directly.
-- Wings are granted ONLY through streak milestones (7-day: +2, 30-day: +6).

--------------------------------------------------------------------------------
-- ADD STREAK FIELDS TO USER TOTALS
--------------------------------------------------------------------------------

ALTER TABLE wings_ecn_user_totals
    ADD COLUMN streak_last_date      DATE    DEFAULT NULL,
    ADD COLUMN streak_current_days   INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN streak_longest_days   INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN wings_ecn_user_totals.streak_last_date IS 'Date of last successful check-in (UTC date only)';
COMMENT ON COLUMN wings_ecn_user_totals.streak_current_days IS 'Current consecutive check-in streak';
COMMENT ON COLUMN wings_ecn_user_totals.streak_longest_days IS 'Longest streak ever achieved by user';

--------------------------------------------------------------------------------
-- ADD NEW ACTION TYPES FOR STREAK MILESTONES
--------------------------------------------------------------------------------

-- Add streak milestone action types to action_log constraint
ALTER TABLE wings_ecn_action_log
    DROP CONSTRAINT IF EXISTS wings_ecn_action_log_action_log_type_check;

ALTER TABLE wings_ecn_action_log
    ADD CONSTRAINT wings_ecn_action_log_action_log_type_check
    CHECK (action_log_type IN (
        'Daily Check-In', 'Send Message',
        'WingedX - Weekly Payment', 'WingedX - Monthly Payment',
        'Winged+ - Weekly Payment', 'Winged+ - Monthly Payment',
        'Winged+ - 3 Month Payment', 'Winged+ - 6 Month Payment',
        'Top Up - Mini', 'Top Up - Boost', 'Top Up - Premium',
        'Referral - Friend Signup', 'Referral - Friend Complete',
        'Attend a Date',
        'Streak - 7 Day Milestone', 'Streak - 30 Day Milestone'
    ));

-- Add streak milestone action types to transaction constraint
ALTER TABLE wings_ecn_transaction
    DROP CONSTRAINT IF EXISTS wings_ecn_transaction_action_log_type_check;

ALTER TABLE wings_ecn_transaction
    ADD CONSTRAINT wings_ecn_transaction_action_log_type_check
    CHECK (action_log_type IN (
        'Daily Check-In', 'Send Message',
        'WingedX - Weekly Payment', 'WingedX - Monthly Payment',
        'Winged+ - Weekly Payment', 'Winged+ - Monthly Payment',
        'Winged+ - 3 Month Payment', 'Winged+ - 6 Month Payment',
        'Top Up - Mini', 'Top Up - Boost', 'Top Up - Premium',
        'Referral - Friend Signup', 'Referral - Friend Complete',
        'Attend a Date',
        'Streak - 7 Day Milestone', 'Streak - 30 Day Milestone'
    ));
