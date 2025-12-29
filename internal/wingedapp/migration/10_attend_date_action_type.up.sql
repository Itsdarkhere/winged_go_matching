-- Migration 10: Add 'Attend a Date' action type for wings economy

-- Add 'Attend a Date' to action_log_type CHECK constraint in wings_ecn_action_log
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

-- Add 'Attend a Date' to action_log_type CHECK constraint in wings_ecn_transaction
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
