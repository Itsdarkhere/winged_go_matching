-- Migration 03: Users and User-Related Tables

--------------------------------------------------------------------------------
-- USERS
--------------------------------------------------------------------------------

CREATE TABLE users
(
    id                          UUID PRIMARY KEY             DEFAULT gen_random_uuid(),
    supabase_id                 UUID,

    -- basic fields
    first_name                  VARCHAR(255),
    last_name                   VARCHAR(255),
    email                       VARCHAR(255) UNIQUE NOT NULL,
    password                    VARCHAR(255),
    address                     VARCHAR(255),
    mobile_number               VARCHAR(255),
    birthday                    DATE,
    gender                      VARCHAR(255),
    height_cm                   INTEGER,
    location                    TEXT,
    dating_pref_age_range_start INTEGER,
    dating_pref_age_range_end   INTEGER,
    agent_dating                BOOLEAN,

    -- registration and authentication fields
    reset_token                 VARCHAR(255),
    registration_code           VARCHAR(255),
    registration_code_sent_at   TIMESTAMPTZ,
    last_checked_call_status    TIMESTAMPTZ,
    agent_deployed              BOOLEAN                      DEFAULT FALSE,
    selected_intro_id           UUID,
    registered_successfully     BOOLEAN                      DEFAULT FALSE,
    mobile_code                 VARCHAR(255),
    sha256_hash                 VARCHAR(256),
    mobile_confirmed            BOOLEAN                      DEFAULT FALSE,
    user_invite_code_ref_id     UUID REFERENCES user_invite_code (id) ON DELETE SET NULL,
    created_by                  UUID,

    -- transcript fields
    latest_transcript_id        UUID                         DEFAULT NULL,
    latest_transcript_ts        TIMESTAMPTZ                  DEFAULT NULL,
    has_transcript              BOOLEAN             NOT NULL DEFAULT FALSE,

    -- location fields
    latitude                    FLOAT,
    longitude                   FLOAT,

    -- Venue Intelligence: ideal date preferences (from onboarding)
    ideal_first_date_phrase     TEXT,
    date_type_bucket            VARCHAR(32),
    date_type_subtype           VARCHAR(64),

    -- type and sexuality (string enums instead of FK)
    user_type                   VARCHAR(64) NOT NULL DEFAULT 'Regular User'
        CHECK (user_type IN ('Admin', 'Regular User')),
    sexuality                   VARCHAR(64)
        CHECK (sexuality IS NULL OR sexuality IN (
            'Prefer not to say', 'Straight', 'Gay', 'Lesbian',
            'Bisexual', 'Asexual', 'Questioning', 'Other'
        )),
    sexuality_is_visible        BOOLEAN,

    -- Audit fields
    last_updated_by             UUID,
    created_at                  TIMESTAMPTZ,
    updated_at                  TIMESTAMPTZ,
    is_active                   BOOLEAN                      DEFAULT TRUE,

    -- Test user flag (for separating test data from real users in batch matching)
    is_test_user                BOOLEAN                      DEFAULT FALSE,

    CONSTRAINT users_created_by_fk FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT users_last_updated_by_fk FOREIGN KEY (last_updated_by) REFERENCES users (id) ON DELETE RESTRICT
);

--------------------------------------------------------------------------------
-- USER DATING PREFERENCES
--------------------------------------------------------------------------------

CREATE TYPE dating_preferences AS ENUM ('Male', 'Female', 'Non-Binary');

CREATE TABLE user_dating_preferences
(
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID               NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    dating_preference dating_preferences NOT NULL,
    UNIQUE (user_id, dating_preference)
);

--------------------------------------------------------------------------------
-- USER SCHEDULING PREFERENCES (Tier 0)
-- Using string enums instead of FK to category
--------------------------------------------------------------------------------

-- Dietary restrictions for venue filtering
CREATE TABLE user_dietary_restriction
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    dietary_restriction  VARCHAR(64) NOT NULL
        CHECK (dietary_restriction IN (
            'Vegetarian', 'Vegan', 'Dairy-Free', 'Gluten-Free',
            'Alcohol-Free', 'Halal', 'Kosher', 'No Restrictions'
        )),
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, dietary_restriction)
);

-- Preferred date types (derived from ideal_date, manually editable)
CREATE TABLE user_date_type_preference
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    date_type_core  VARCHAR(32) NOT NULL
        CHECK (date_type_core IN ('coffee', 'drinks', 'meal', 'walk', 'activity')),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, date_type_core)
);

-- Accessibility/mobility constraints for venue filtering
CREATE TABLE user_mobility_constraint
(
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    mobility_constraint VARCHAR(64) NOT NULL
        CHECK (mobility_constraint IN (
            'Wheelchair Accessible', 'Limited Walking', 'No Stairs',
            'Service Animal', 'None'
        )),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, mobility_constraint)
);

--------------------------------------------------------------------------------
-- USER ELEVEN LABS
--------------------------------------------------------------------------------

CREATE TABLE user_eleven_labs
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID  NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    conversation JSONB NOT NULL,
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    is_active    BOOLEAN          DEFAULT TRUE
);

--------------------------------------------------------------------------------
-- USER BLOCKED CONTACT
--------------------------------------------------------------------------------

CREATE TABLE user_blocked_contact
(
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    blocked_number TEXT
);

--------------------------------------------------------------------------------
-- USER PHOTOS
--------------------------------------------------------------------------------

CREATE TABLE user_photo
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID          NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    bucket     VARCHAR(255)  NOT NULL,
    key        VARCHAR(2048) NOT NULL UNIQUE,
    order_no   INT           NOT NULL,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    is_active  BOOLEAN          DEFAULT TRUE,
    UNIQUE (user_id, bucket, key),
    UNIQUE (user_id, order_no)
);

--------------------------------------------------------------------------------
-- USER AI CONTEXT
--------------------------------------------------------------------------------

CREATE TABLE user_ai_context
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    ai_context_type VARCHAR(64) NOT NULL DEFAULT 'Your Agent'
        CHECK (ai_context_type IN ('Your Agent')),
    context         TEXT        NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ,
    is_active       BOOLEAN     DEFAULT TRUE
);

--------------------------------------------------------------------------------
-- USER AI CONVO
--------------------------------------------------------------------------------

CREATE TABLE user_ai_convo
(
    id                 UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    user_id            UUID         NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    prompt_response_id VARCHAR(255) NOT NULL,
    ai_convo_type      VARCHAR(64)  NOT NULL DEFAULT 'AI'
        CHECK (ai_convo_type IN ('AI', 'Your Agent')),
    message            TEXT         NOT NULL,
    additional_context TEXT         NOT NULL DEFAULT '',
    response           TEXT         NOT NULL,
    created_at         TIMESTAMPTZ           DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMPTZ,
    is_active          BOOLEAN               DEFAULT TRUE
);

--------------------------------------------------------------------------------
-- GENERAL AI CONTEXT
--------------------------------------------------------------------------------

CREATE TABLE general_ai_context
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ai_context_type VARCHAR(64) NOT NULL DEFAULT 'Your Agent'
        CHECK (ai_context_type IN ('Your Agent')),
    context         TEXT        NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ,
    is_active       BOOLEAN     DEFAULT TRUE
);

-- add initial context
INSERT INTO general_ai_context (ai_context_type, context)
VALUES ('Your Agent',
       $CONTEXT$Context
      You are the user''s personal AI dating agent inside Winged. Every user has a persistent, private thread with you, and your entire purpose is to support them in their dating journey.
      You know their history, preferences, and prior conversations. You help them reflect on their dating experiences, prepare for upcoming dates, and think through their goals, anxieties, or confusion around matches.
      You only talk about things related to dating, romance, their own life in relation to dating, or scheduled dates. You do not wander into unrelated topics. You are the user''s supportive, thoughtful, slightly playful wingman/wingwoman who sticks with them through their dating journey.
      You also connect with the date threads when relevant: if a user has an upcoming date, you can help prepare conversation starters, outfit ideas, and emotional readiness.
      Your thread is limited to 2 free messages a day, after which additional interactions cost Wings. Your responses should feel valuable and premium — more reflective, more strategic, more emotionally intelligent than a regular chat.

      Objective
      Be the user''s trusted dating companion.
      Check in daily with personalized prompts.
      Help them process emotions, reflect on matches, and improve their dating confidence.
      Offer both strategic support (profile improvements, match suggestions, conversation tips) and emotional support (validating feelings, reducing anxiety, celebrating wins).
      Keep the focus on dating, their experiences, and their goals.
      Make them feel understood, supported, and motivated to continue their dating journey.

      Style
      Conversational and intimate, like a close friend who knows their dating life inside-out.
      Honest and reflective — not superficial pep-talks, but genuine support.
      Encouraging without being clingy or overbearing.
      Natural language: easy flow, not overly scripted.

      Tone
      Warm, empathetic, and supportive.
      Playful when appropriate, especially around date prep or match excitement.
      Grounded in emotional intelligence — always helping the user process, not just giving advice.

      Audience
      Single users actively navigating dating (sometimes excited, sometimes anxious, sometimes discouraged).
      They value authentic connection, thoughtful support, and personalized insights.
      They want an agent who remembers them, adapts to them, and doesn''t waste time on surface-level chat.

      Response
      Keep all responses centered on dating, their experiences, their profile, or their upcoming dates.
      Reference user history and context whenever possible.
      Responses can be longer than typical chat — reflective and high-value.
      Ask questions that invite the user to share (feelings, reflections, goals).
      Offer concrete support (e.g., date prep, conversation starters, self-reflection exercises).
      Celebrate progress and normalize setbacks.
      Daily Check-In Starters (ROTATE daily):
      "How are you feeling about dating today?"
      "Want to review your matches or just talk about what''s on your mind?"
      "Any updates you''d like to share — or are you prepping for an upcoming date?"
      "What''s been exciting or frustrating in your dating life this week?"
      Success Metric: User feels supported, stays engaged with the personal thread, and uses you to prepare for and reflect on their dating journey.$CONTEXT$);
