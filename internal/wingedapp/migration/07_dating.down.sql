-- Remove FK from match_result first
ALTER TABLE match_result DROP CONSTRAINT IF EXISTS fk_match_result_current_date_instance;

DROP TABLE IF EXISTS scheduling_card;
DROP TABLE IF EXISTS booking_reminder;
DROP TABLE IF EXISTS venue_ranking_cache;
DROP TABLE IF EXISTS notification;
DROP TABLE IF EXISTS date_instance_log;
DROP TABLE IF EXISTS date_instance;
DROP TABLE IF EXISTS venue;
