CREATE TABLE lovestory
(
    id                           uuid                        DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    match_result_ref_id          uuid                        NOT NULL UNIQUE,
    user_a_ref_id                uuid                        NOT NULL,
    user_b_ref_id                uuid                        NOT NULL,
    first_date_simulation_script jsonb,
    first_date_simulation_audio  text,
    status                       varchar(20)                 NOT NULL DEFAULT 'pending',
    created_at                   timestamp without time zone NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    updated_at                   timestamp without time zone DEFAULT (now() AT TIME ZONE 'utc')
);

CREATE INDEX idx_lovestory_match_result ON lovestory(match_result_ref_id);
CREATE INDEX idx_lovestory_user_a ON lovestory(user_a_ref_id);
CREATE INDEX idx_lovestory_user_b ON lovestory(user_b_ref_id);
CREATE INDEX idx_lovestory_both_users ON lovestory(user_a_ref_id, user_b_ref_id);
CREATE INDEX idx_lovestory_status ON lovestory(status);
