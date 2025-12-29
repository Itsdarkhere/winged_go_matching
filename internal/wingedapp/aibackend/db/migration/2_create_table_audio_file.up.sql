create table audio_files
(
    id               uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    user_id          uuid                           not null,
    webhook_type     varchar(100)                   not null,
    event_timestamp  integer                        not null,
    created_at       timestamp                      not null,
    updated_at       timestamp,
    agent_id         varchar(255),
    conversation_id  varchar(255),
    audio_format     varchar(10)                    not null,
    audio_size_bytes bigint,
    duration_seconds double precision,
    audio_data       bytea,
    raw_payload      jsonb                          not null,
    voice_id         varchar(255),
    category         varchar(100),
    storage_path     varchar(500),
    storage_url      varchar(1000),
    storage_bucket   varchar(100)
);