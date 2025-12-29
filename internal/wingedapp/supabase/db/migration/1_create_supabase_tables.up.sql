-- Supabase Auth: Users table only
-- This is the minimal schema needed for our app to interact with Supabase auth

CREATE TABLE users
(
    instance_id                 uuid,
    id                          uuid                                 NOT NULL,
    aud                         character varying(255),
    role                        character varying(255),
    email                       character varying(255),
    encrypted_password          character varying(255),
    email_confirmed_at          timestamp with time zone,
    invited_at                  timestamp with time zone,
    confirmation_token          character varying(255),
    confirmation_sent_at        timestamp with time zone,
    recovery_token              character varying(255),
    recovery_sent_at            timestamp with time zone,
    email_change_token_new      character varying(255),
    email_change                character varying(255),
    email_change_sent_at        timestamp with time zone,
    last_sign_in_at             timestamp with time zone,
    raw_app_meta_data           jsonb,
    raw_user_meta_data          jsonb,
    is_super_admin              boolean,
    created_at                  timestamp with time zone,
    updated_at                  timestamp with time zone,
    phone                       text                   DEFAULT NULL::character varying,
    phone_confirmed_at          timestamp with time zone,
    phone_change                text                   DEFAULT ''::character varying,
    phone_change_token          character varying(255) DEFAULT ''::character varying,
    phone_change_sent_at        timestamp with time zone,
    confirmed_at                timestamp with time zone GENERATED ALWAYS AS (LEAST(email_confirmed_at, phone_confirmed_at)) STORED,
    email_change_token_current  character varying(255) DEFAULT ''::character varying,
    email_change_confirm_status smallint               DEFAULT 0,
    banned_until                timestamp with time zone,
    reauthentication_token      character varying(255) DEFAULT ''::character varying,
    reauthentication_sent_at    timestamp with time zone,
    is_sso_user                 boolean                DEFAULT false NOT NULL,
    deleted_at                  timestamp with time zone,
    is_anonymous                boolean                DEFAULT false NOT NULL,
    CONSTRAINT users_email_change_confirm_status_check CHECK (((email_change_confirm_status >= 0) AND (email_change_confirm_status <= 2)))
);

-- Primary Key
ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

-- Unique constraints
ALTER TABLE ONLY users
    ADD CONSTRAINT users_phone_key UNIQUE (phone);

-- Indexes
CREATE UNIQUE INDEX confirmation_token_idx ON users USING btree (confirmation_token) WHERE ((confirmation_token)::text !~ '^[0-9 ]*$'::text);
CREATE UNIQUE INDEX email_change_token_current_idx ON users USING btree (email_change_token_current) WHERE ((email_change_token_current)::text !~ '^[0-9 ]*$'::text);
CREATE UNIQUE INDEX email_change_token_new_idx ON users USING btree (email_change_token_new) WHERE ((email_change_token_new)::text !~ '^[0-9 ]*$'::text);
CREATE UNIQUE INDEX reauthentication_token_idx ON users USING btree (reauthentication_token) WHERE ((reauthentication_token)::text !~ '^[0-9 ]*$'::text);
CREATE UNIQUE INDEX recovery_token_idx ON users USING btree (recovery_token) WHERE ((recovery_token)::text !~ '^[0-9 ]*$'::text);
CREATE UNIQUE INDEX users_email_partial_key ON users USING btree (email) WHERE (is_sso_user = false);
CREATE INDEX users_instance_id_email_idx ON users USING btree (instance_id, lower((email)::text));
CREATE INDEX users_instance_id_idx ON users USING btree (instance_id);
CREATE INDEX users_is_anonymous_idx ON users USING btree (is_anonymous);
