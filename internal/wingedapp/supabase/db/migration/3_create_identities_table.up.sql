-- Supabase Auth: Identities table
-- Required for Supabase auth login (OTP/email) to work properly

CREATE TABLE identities
(
    id               uuid                     DEFAULT gen_random_uuid() NOT NULL,
    user_id          uuid                                               NOT NULL,
    identity_data    jsonb                                              NOT NULL,
    provider         text                                               NOT NULL,
    provider_id      text                                               NOT NULL,
    last_sign_in_at  timestamp with time zone,
    created_at       timestamp with time zone,
    updated_at       timestamp with time zone,
    email            text GENERATED ALWAYS AS (lower(identity_data->>'email')) STORED,
    CONSTRAINT identities_pkey PRIMARY KEY (id),
    CONSTRAINT identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX identities_user_id_idx ON identities USING btree (user_id);
CREATE UNIQUE INDEX identities_provider_id_provider_unique ON identities USING btree (provider_id, provider);

COMMENT ON TABLE identities IS 'Auth: Stores identities associated to a user.';
