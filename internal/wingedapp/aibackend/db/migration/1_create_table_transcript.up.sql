CREATE TABLE transcripts
(
    id                   uuid                     DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    user_id              uuid                                               NOT NULL,
    webhook_type         character varying(100)                             NOT NULL,
    event_timestamp      integer                                            NOT NULL,
    created_at           timestamp with time zone DEFAULT now()             NOT NULL,
    updated_at           timestamp with time zone DEFAULT now(),
    agent_id             character varying(255),
    conversation_id      character varying(255),
    status               character varying(50),
    call_duration_secs   integer,
    cost                 integer,
    main_language        character varying(10),
    termination_reason   character varying(255),
    call_successful      character varying(50),
    extroversion_score   double precision,
    interview_smoothness character varying(50),
    raw_payload          jsonb                                              NOT NULL,
    transcript_data      jsonb,
    call_metadata        jsonb,
    analysis             jsonb
);