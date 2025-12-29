-- Migration 11 DOWN: Remove streak tracking fields

--------------------------------------------------------------------------------
-- REMOVE STREAK FIELDS FROM USER TOTALS
--------------------------------------------------------------------------------

ALTER TABLE wings_ecn_user_totals
    DROP COLUMN IF EXISTS streak_last_date,
    DROP COLUMN IF EXISTS streak_current_days,
    DROP COLUMN IF EXISTS streak_longest_days;

--------------------------------------------------------------------------------
-- RESTORE ORIGINAL ACTION TYPE CONSTRAINTS (without streak milestones)
--------------------------------------------------------------------------------

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
        'Attend a Date'
    ));

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
        'Attend a Date'
    ));
