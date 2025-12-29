-- Migration 01: Foundation
-- Extensions and sys_param (category tables removed - using string enums)

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS btree_gist;

--------------------------------------------------------------------------------
-- SYS PARAM
--------------------------------------------------------------------------------

CREATE TABLE sys_param
(
    key  VARCHAR(255) UNIQUE PRIMARY KEY,
    val  TEXT NOT NULL,
    misc JSONB DEFAULT NULL
);

INSERT INTO sys_param (key, val)
VALUES ('SETTING_UP_AGENT_TRANSCRIPT_MAX_LOOPS', '6'),
       ('SETTING_UP_AGENT_TRANSCRIPT_LOOP_INTERVAL_SECS', '5'),
       ('SETTING_UP_AGENT_AUDIO_FILES_MAX_LOOPS', '6'),
       ('SETTING_UP_AGENT_AUDIO_FILES_LOOP_INTERVAL_SECS', '5'),
       ('USER_INVITE_CODE_MAX_USAGE', '20'),
       ('BROKEN_AUDIO_DURATION_THRESHOLD_SECS', '5'),
       ('SETTING_UP_AGENT_IDLE_WAITING_TIME_SECS', '15'),
       ('INVITE_EXPIRY_DAYS', '20'),
       ('WINGS_ECON_INCREMENT_THRESH_SEND_MESSAGE', '5');

--------------------------------------------------------------------------------
-- JOBS (async background processing)
--------------------------------------------------------------------------------

CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payload JSONB,
    result JSONB,
    error TEXT,
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT jobs_status_check CHECK (status IN ('pending', 'running', 'completed', 'failed'))
);

-- Index for worker polling (pending jobs, oldest first)
CREATE INDEX idx_jobs_pending ON jobs (created_at) WHERE status = 'pending';

-- Index for status queries
CREATE INDEX idx_jobs_status ON jobs (status);

-- Index for type queries
CREATE INDEX idx_jobs_type ON jobs (type);
