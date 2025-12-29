-- Migration 02: Invite Codes and Anonymized Contacts

--------------------------------------------------------------------------------
-- USER INVITE CODE
--------------------------------------------------------------------------------

CREATE TABLE user_invite_code
(
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invite_code         VARCHAR(6)   NOT NULL UNIQUE,
    usage_count         INT          NOT NULL DEFAULT 0,
    referral_source     VARCHAR(255) NOT NULL,
    invite_code_type    VARCHAR(64)  NOT NULL DEFAULT 'Referral'
        CHECK (invite_code_type IN ('Referral', 'Event')),
    for_number          VARCHAR(256),
    for_number_hash     VARCHAR(256),
    referrer_number_hash VARCHAR(256),
    last_used           TIMESTAMP,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--------------------------------------------------------------------------------
-- ANONYMIZED CONTACT
--------------------------------------------------------------------------------

CREATE TABLE anonymized_contact
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_hash   VARCHAR(255) NOT NULL,
    contact_hash VARCHAR(255) NOT NULL,
    created_at   TIMESTAMPTZ       DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (owner_hash, contact_hash)
);

CREATE INDEX idx_anonymized_contact_owner_hash ON anonymized_contact (owner_hash);
CREATE INDEX idx_anonymized_contact_contact_hash ON anonymized_contact (contact_hash);
