# User-Facing Matching Flow Specification

> **Author:** Claude (AI Analysis)
> **Based on:** Codebase analysis of placeholder endpoints, domain models, and business logic
> **Status:** Extrapolated spec for implementation guidance

---

## Executive Summary

The matching system operates as a **curated dating experience** where:
1. Admin ingests population data and runs batch matching algorithm
2. Algorithm generates scored pairs using hard qualifiers + LLM compatibility
3. Admin reviews and approves high-quality matches
4. Approved matches are **dropped** to users daily at configured times
5. Users see one match at a time, can **propose** (accept) or **pass** (decline)
6. Mutual proposals unlock **chat**

This is NOT a swipe-based app. It's a **quality-over-quantity** matchmaking service with human curation.

---

## User Journey Flow

```
                                    ADMIN FLOW
    ┌──────────────────────────────────────────────────────────────────┐
    │                                                                  │
    │  [CSV Upload] → [Batch Ingest] → [Algorithm Scoring] → [Review]  │
    │                                                                  │
    │                              ↓                                   │
    │                       [Approve Match]                            │
    │                              ↓                                   │
    └──────────────────────────────┼───────────────────────────────────┘
                                   │
                        ┌──────────┴──────────┐
                        │   CRON: Match Drop   │
                        │  (daily at DropHours) │
                        └──────────┬──────────┘
                                   │
                                   ▼
    ┌──────────────────────────────────────────────────────────────────┐
    │                         USER FLOW                                │
    │                                                                  │
    │  [Get Matches] → [View Match Profile] → [Propose] or [Pass]     │
    │        │                                      │                  │
    │        │                                      ▼                  │
    │        │                              ┌──────────────┐           │
    │        │                              │ Mutual       │           │
    │        │                              │ Proposal?    │           │
    │        │                              └──────┬───────┘           │
    │        │                                     │                   │
    │        │                          YES ───────┴────── NO          │
    │        │                           │                 │           │
    │        │                           ▼                 ▼           │
    │        │                    [Chat Unlocked]   [Wait for other]   │
    │        │                           │                             │
    │        ▼                           ▼                             │
    │  [Chat List] ◄───────────── [Send Messages]                      │
    │                                                                  │
    └──────────────────────────────────────────────────────────────────┘
```

---

## API Endpoints (User-Facing)

### Match Discovery

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/user/matching` | GET | List all dropped matches for current user |
| `/user/matching/:match_id` | GET | Get single match details with profile |
| `/user/matching/:match_id/propose` | POST | Accept/like this match |
| `/user/matching/:match_id/pass` | POST | Decline/skip this match |
| `/user/matching/seen` | PATCH | Mark matches as viewed |

### Chat

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/user/matching/chat-list` | GET | List all unlocked conversations |
| `/user/matching/chat-list/:match_id` | GET | Get conversation history |
| `/user/matching/chat-list/:match_id` | POST | Send message in conversation |

---

## Domain Model Specifications

### `Match` - User-Facing Match Card

The match object shown to users when viewing potential partners.

```go
type Match struct {
    ID           string   // UUID of the match result
    Name         string   // Partner's display name (first name only for privacy)
    Age          int      // Partner's age in years
    Gender       string   // "Male", "Female", "Non-Binary"
    Sexuality    string   // Sexual orientation for compatibility display
    IntroURL     string   // Audio introduction URL (voice memo from partner)
    Images       []Image  // Ordered profile photos (max 4-6)
    DistanceInKM float64  // Geographic distance from current user
    MatchPercent float64  // Algorithm compatibility score (0.0-1.0)
    YourStatus   string   // Current user's action: "Pending", "Proposed", "Passed"
    MatchStatus  string   // Partner's action: "Pending", "Proposed", "Passed"
    Seen         bool     // Whether user has viewed this match
}
```

#### Field Semantics

| Field | Domain Purpose |
|-------|----------------|
| `ID` | Unique identifier for API operations (propose, pass, chat) |
| `Name` | First name only - full identity revealed after mutual match |
| `Age` | Derived from birthday; key qualifier in matching algorithm |
| `Gender` | Used in dating preference matching |
| `Sexuality` | Helps users understand compatibility context |
| `IntroURL` | Voice introduction - differentiator from photo-only apps |
| `Images` | Ordered array; first image is primary profile photo |
| `DistanceInKM` | Calculated from lat/long; respects LocationRadiusKM config |
| `MatchPercent` | Composite score from qualitative LLM analysis (personality, lifestyle, values) |
| `YourStatus` | State machine: Pending → Proposed/Passed |
| `MatchStatus` | Partner's state; "Proposed" + "Proposed" = Chat unlocked |
| `Seen` | Analytics tracking; prevents re-showing viewed matches prominently |

---

### `Image` - Profile Photo

```go
type Image struct {
    URL     string // CDN URL to photo (Supabase Storage)
    OrderNo int    // Display order (1 = primary/avatar)
}
```

#### Field Semantics

| Field | Domain Purpose |
|-------|----------------|
| `URL` | Direct link to image; frontend caches aggressively |
| `OrderNo` | User-controlled ordering; 1 appears in cards, rest in detail view |

---

### `ChatBlock` - Conversation Preview

Summary object for the chat list screen.

```go
type ChatBlock struct {
    ID          string // Match result UUID (same as Match.ID)
    Name        string // Partner's name
    Seen        bool   // Whether there are unread messages
    YourStatus  string // For UI state indication
    MatchStatus string // For UI state indication
}
```

#### Field Semantics

| Field | Domain Purpose |
|-------|----------------|
| `ID` | Links to full conversation; same UUID as match |
| `Name` | Partner name for chat header |
| `Seen` | Unread indicator; false = new messages exist |
| `YourStatus` | Allows showing "You proposed" context |
| `MatchStatus` | Allows showing "They proposed" context |

---

### `ChatMessage` - Individual Message

```go
type ChatMessage struct {
    ID        string    // Message UUID
    Role      string    // "current_user" or "current_match"
    Message   string    // Text content
    CreatedAt time.Time // Timestamp for ordering
}
```

#### Field Semantics

| Field | Domain Purpose |
|-------|----------------|
| `ID` | Unique per message; enables edit/delete if implemented |
| `Role` | Determines message alignment (left/right in UI) |
| `Message` | Plain text; may support markdown/emoji in future |
| `CreatedAt` | Chronological ordering; enables "time since" display |

---

## State Machine: Match Lifecycle

```
                    ┌─────────────────────────────────────────┐
                    │           MatchResult States            │
                    └─────────────────────────────────────────┘

                           ADMIN DOMAIN
    ┌───────────────────────────────────────────────────────────┐
    │                                                           │
    │   [Created]                                               │
    │       │                                                   │
    │       ▼                                                   │
    │   [Algorithm Scored]                                      │
    │       │                                                   │
    │       ├──── is_possible_match: false ──► [Rejected]       │
    │       │                                                   │
    │       ├──── is_possible_match: true                       │
    │       │           │                                       │
    │       │           ▼                                       │
    │       │   [Awaiting Admin Review]                         │
    │       │           │                                       │
    │       │           ├── Approve ──► is_approved: true       │
    │       │           │                                       │
    │       │           └── Reject ──► (stays false)            │
    │       │                                                   │
    └───────┼───────────────────────────────────────────────────┘
            │
            ▼ (Cron: DropOneMatchPerUser)

                           USER DOMAIN
    ┌───────────────────────────────────────────────────────────┐
    │                                                           │
    │   [Dropped to User]  (is_dropped: true)                   │
    │       │                                                   │
    │       ├──── User A: Propose ──► YourStatus: "Proposed"    │
    │       │                                                   │
    │       ├──── User A: Pass ──► YourStatus: "Passed"         │
    │       │                                                   │
    │       └──── No action ──► (expires after MatchExpiration) │
    │                                                           │
    │   [Mutual Proposal] (both YourStatus = "Proposed")        │
    │       │                                                   │
    │       ▼                                                   │
    │   [Chat Unlocked]                                         │
    │       │                                                   │
    │       ├──── Active conversation                           │
    │       │                                                   │
    │       └──── Stale ──► StaleChatNudge triggers AI nudge    │
    │                                                           │
    └───────────────────────────────────────────────────────────┘
```

---

## Predicted Implementation: User Endpoints

### `GET /user/matching` - List Dropped Matches

**What it needs to do:**
1. Get authenticated user ID from JWT
2. Query match_results where:
   - `is_approved = true`
   - `is_dropped = true`
   - `(user_a_id = current_user OR user_b_id = current_user)`
   - `is_expired = false`
3. Transform to `[]Match` with partner details
4. Calculate `MatchPercent` from `qualifier_results.total_score`
5. Compute `DistanceInKM` from current user location

**Expected Query Filter:**
```go
&QueryFilterMatchResult{
    IsApproved:  null.BoolFrom(true),
    IsDropped:   null.BoolFrom(true),
    IsExpired:   null.BoolFrom(false),
    // Custom: involves_user filter needed
}
```

**Returns:** Array of matches not yet acted upon (Pending status)

---

### `GET /user/matching/:match_id` - Single Match Detail

**What it needs to do:**
1. Verify match exists and involves current user
2. Fetch full profile data for partner:
   - Basic: age, gender, height
   - Photos: ordered image URLs
   - Audio: intro_url
   - Profile: qualitative/quantitative sections (for "About" display)
3. Calculate live distance
4. Return enriched `Match` object

**Permissions:** Only accessible if user is user_a or user_b in match

---

### `POST /user/matching/:match_id/propose` - Accept Match

**What it needs to do:**
1. Verify match involves current user
2. Update user's status to "Proposed"
3. Check if mutual proposal:
   - If both proposed → unlock chat capability
   - Send push notification to partner
4. Log action for analytics

**State Change:**
```
YourStatus: "Pending" → "Proposed"
```

**Side Effects:**
- If mutual: Create chat room / enable messaging
- Send notification: "You have a new match!"

---

### `POST /user/matching/:match_id/pass` - Decline Match

**What it needs to do:**
1. Verify match involves current user
2. Update user's status to "Passed"
3. Apply `MatchBlockDeclined` cooldown (prevent re-matching for N hours)
4. Log action for analytics

**State Change:**
```
YourStatus: "Pending" → "Passed"
```

**Config Reference:**
- `MatchBlockDeclined`: Hours before this pair can be re-matched

---

### `GET /user/matching/chat-list` - List Conversations

**What it needs to do:**
1. Find all matches where:
   - Both users have `YourStatus = "Proposed"`
   - Chat is active (not closed)
2. For each, get latest message preview
3. Calculate unread count
4. Order by last message timestamp (recent first)

**Returns:** `[]ChatBlock` with conversation summaries

---

### `GET /user/matching/chat-list/:match_id` - Conversation History

**What it needs to do:**
1. Verify chat is unlocked (mutual proposal)
2. Fetch messages ordered by `created_at ASC`
3. Mark messages as read for current user
4. Check for stale chat (no messages in `StaleChatNudge` hours)

**Returns:** `[]ChatMessage` in chronological order

---

### `POST /user/matching/chat-list/:match_id` - Send Message

**What it needs to do:**
1. Verify chat is unlocked
2. Validate message content (non-empty, within length limits)
3. Insert message with `role = "current_user"`
4. Send push notification to partner
5. Reset stale chat timer

**Request Body:**
```json
{
    "message": "Hey! Your intro audio was really funny..."
}
```

---

## Configuration Dependencies

These `Config` fields directly impact user-facing behavior:

| Config Field | User Impact |
|--------------|-------------|
| `DropHours` | When matches appear (e.g., "19:00" = 7 PM daily) |
| `DropHoursUTC` | Timezone for drop schedule |
| `MatchExpirationHours` | How long to act before match expires |
| `StaleChatNudge` | Hours before AI prompts "keep chatting" |
| `StaleChatAgentSetup` | Hours before AI agent intervenes |
| `MatchBlockDeclined` | Cooldown after passing on someone |
| `MatchBlockIgnored` | Cooldown after ignoring (no action) |
| `MatchBlockClosed` | Cooldown after closing chat |
| `LocationRadiusKM` | Max distance shown in match cards |

---

## Future Considerations

### Missing in Current Stubs

1. **Notification System** - Push notifications for:
   - New match dropped
   - Partner proposed
   - New message received
   - Stale chat nudge

2. **Read Receipts** - Track message seen status

3. **Typing Indicators** - Real-time chat UX

4. **Block/Report** - Safety features for chat

5. **Match Expiration Handling** - What happens when timer runs out

6. **Re-match Logic** - After cooldown period, can pair be re-evaluated?

### Suggested Additions

```go
// Extend ChatMessage
type ChatMessage struct {
    // ... existing fields
    SeenAt    *time.Time `json:"seen_at,omitempty"`    // Read receipt
    EditedAt  *time.Time `json:"edited_at,omitempty"`  // Edit tracking
}

// Extend Match
type Match struct {
    // ... existing fields
    ExpiresAt    time.Time `json:"expires_at"`         // Countdown timer
    UnlockedAt   *time.Time `json:"unlocked_at"`       // When chat opened
    LastActivity *time.Time `json:"last_activity"`     // For stale detection
}
```

---

## Appendix: Lib Layer Models Referenced

### `MatchResult` - Backend Representation

The underlying database model that powers user-facing `Match`:

```go
type MatchResult struct {
    ID                   uuid.UUID  // Primary key
    MatchSetID           uuid.UUID  // Batch reference
    UserAID              uuid.UUID  // First user in pair
    UserBID              uuid.UUID  // Second user in pair
    QualifierResults     null.JSON  // Algorithm scoring breakdown
    MatchedQualitatively null.Bool  // Passed LLM scoring
    IsPossibleMatch      bool       // Passed all hard qualifiers
    IsApproved           bool       // Admin approved for delivery
    IsExpired            bool       // Timed out without action
    UserLifeCycleStatus  string     // State machine position
}
```

### `Config` - Matching Parameters

```go
type Config struct {
    // Age Rules
    AgeRangeStart             int       // Min age (e.g., 18)
    AgeRangeEnd               int       // Max age (e.g., 65)
    AgeRangeWomanOlderBy      int       // How much older woman can be
    AgeRangeManOlderBy        int       // How much older man can be

    // Physical
    HeightMaleGreaterByCM     float64   // Preferred height difference

    // Location
    LocationRadiusKM          float64   // Max match distance
    LocationAdaptiveExpansion []int64   // Progressive radius expansion

    // Scheduling
    DropHours                 []string  // When to drop matches ("19:00")
    DropHoursUTC              []string  // Timezone references

    // Engagement
    StaleChatNudge            int       // Hours before nudge
    StaleChatAgentSetup       int       // Hours before AI agent
    MatchExpirationHours      int       // Time to act on match

    // Cooldowns
    MatchBlockDeclined        int       // Hours after decline
    MatchBlockIgnored         int       // Hours after ignore
    MatchBlockClosed          int       // Hours after chat close

    // Scoring
    ScoreRangeStart           float64   // Min compatibility (0.0)
    ScoreRangeEnd             float64   // Max compatibility (1.0)
}
```

### `PersonProfile` - Compatibility Data

What the LLM uses to calculate `MatchPercent`:

```go
type PersonProfile struct {
    Qualitative  QualitativeSection   // Text personality data
    Quantitative QuantitativeSection  // 1-10 scale scores
    Categorical  CategoricalSection   // Predefined options
}
```

**Qualitative** (text-based, LLM interprets):
- Self Portrait, Interests, Wellbeing Habits
- Money Management, Moral Frameworks, Life Goals
- Partnership Values, Family Planning, Ideal Date

**Quantitative** (numeric, algorithm compares):
- Extroversion (1-10)
- Routine vs Spontaneity (1-10)
- Agreeableness, Conscientiousness, Neuroticism
- Dominance Level, Emotional Expressiveness
- Sex Drive, Geographical Mobility

**Categorical** (enum matching):
- Conflict Resolution Style
- Sexuality Preferences
- Religion

---

## Implementation Priority

For MVP user-facing launch:

1. **P0 - Must Have**
   - `GET /user/matching` (list matches)
   - `GET /user/matching/:id` (view profile)
   - `POST /user/matching/:id/propose`
   - `POST /user/matching/:id/pass`

2. **P1 - Needed for Engagement**
   - `GET /user/matching/chat-list`
   - `GET /user/matching/chat-list/:id`
   - `POST /user/matching/chat-list/:id`

3. **P2 - Polish**
   - `PATCH /user/matching/seen`
   - Push notifications
   - Stale chat nudges
   - Match expiration handling

---

## Schema Gap Analysis

> **Based on migrations 1-21 analysis**

### Current Schema Summary

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           EXISTING TABLES (Migrations 1-21)                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  CORE USER                          MATCHING                                    │
│  ───────────                        ────────                                    │
│  ├── users                          ├── match_set                               │
│  │   ├── id, email, first_name      │   ├── id, name                            │
│  │   ├── birthday, gender           │   ├── number_of_participants              │
│  │   ├── height_cm                  │   └── match_configuration (JSONB)         │
│  │   ├── latitude, longitude        │                                           │
│  │   ├── sexuality_category_ref_id  ├── match_result                            │
│  │   └── is_active                  │   ├── id, match_set_ref_id                │
│  │                                  │   ├── user_a_ref_id, user_b_ref_id        │
│  ├── user_dating_preferences        │   ├── qualifier_results (JSONB)           │
│  │   └── user_id → dating_pref      │   ├── is_approved, is_dropped             │
│  │                                  │   ├── is_possible_match, is_expired       │
│  ├── user_photo                     │   ├── user_lifecycle_status_ref_id ← ONE  │
│  │   └── user_id → bucket/key       │   └── dropped_ts                          │
│  │                                  │                                           │
│  └── user_match (UNUSED?)           ├── match_config (singleton)                │
│      └── user_ref_id → match_result │   └── all algorithm parameters            │
│                                     │                                           │
│  ECONOMY                            CATEGORIES                                  │
│  ───────                            ──────────                                  │
│  ├── wings_ecn_subscription_plan    ├── category_type                           │
│  ├── wings_ecn_user_subscription    │   └── 'Match Status', 'Sexuality', etc.   │
│  ├── wings_ecn_action_log           │                                           │
│  ├── wings_ecn_transaction          └── category                                │
│  └── wings_ecn_user_totals              └── 'Active', 'Expired', 'Proposed'...  │
│                                                                                 │
│  AI/MISC                                                                        │
│  ───────                                                                        │
│  ├── user_ai_convo (YourAgent)                                                  │
│  ├── sys_param (config KV)                                                      │
│  └── anonymized_contact                                                         │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Critical Schema Gap: Per-User Match Status

**Current Problem:**

```sql
-- match_result has ONE status for the WHOLE pair
user_lifecycle_status_ref_id UUID  -- Points to category 'Active'/'Expired'
```

**Required:** Each user in a match needs INDEPENDENT status tracking.

```
User A: Proposed    ←── needs separate column
User B: Pending     ←── needs separate column
```

Without this, we cannot:
- Know if User A proposed but User B hasn't responded
- Detect mutual proposal (both proposed = chat unlocked)
- Show correct `YourStatus` vs `MatchStatus` in UI

---

### Required Schema Changes

#### Migration 22: Add Per-User Match Status

```sql
-- 22_add_user_match_status.up.sql

-- Add category types for user-level match actions
INSERT INTO category_type (name)
VALUES ('Match User Action');

-- Add action categories
INSERT INTO category (category_type_ref_id, name)
SELECT id, 'Pending'
FROM category_type WHERE name = 'Match User Action'
UNION ALL
SELECT id, 'Proposed'
FROM category_type WHERE name = 'Match User Action'
UNION ALL
SELECT id, 'Passed'
FROM category_type WHERE name = 'Match User Action';

-- Add per-user status columns to match_result
ALTER TABLE match_result
    -- User A's independent action
    ADD COLUMN user_a_action_ref_id UUID,
    ADD COLUMN user_a_action_at     TIMESTAMPTZ,
    ADD COLUMN user_a_seen_at       TIMESTAMPTZ,

    -- User B's independent action
    ADD COLUMN user_b_action_ref_id UUID,
    ADD COLUMN user_b_action_at     TIMESTAMPTZ,
    ADD COLUMN user_b_seen_at       TIMESTAMPTZ,

    -- Chat unlock tracking
    ADD COLUMN chat_unlocked_at     TIMESTAMPTZ,
    ADD COLUMN chat_closed_at       TIMESTAMPTZ,
    ADD COLUMN chat_closed_by       UUID,

    -- Expiration tracking
    ADD COLUMN expires_at           TIMESTAMPTZ,

    -- Constraints
    ADD CONSTRAINT user_a_action_fk FOREIGN KEY (user_a_action_ref_id)
        REFERENCES category (id),
    ADD CONSTRAINT user_b_action_fk FOREIGN KEY (user_b_action_ref_id)
        REFERENCES category (id),
    ADD CONSTRAINT chat_closed_by_fk FOREIGN KEY (chat_closed_by)
        REFERENCES users (id);

-- Set default status to 'Pending' for existing rows
UPDATE match_result
SET user_a_action_ref_id = (
    SELECT c.id FROM category c
    JOIN category_type ct ON c.category_type_ref_id = ct.id
    WHERE ct.name = 'Match User Action' AND c.name = 'Pending'
),
user_b_action_ref_id = (
    SELECT c.id FROM category c
    JOIN category_type ct ON c.category_type_ref_id = ct.id
    WHERE ct.name = 'Match User Action' AND c.name = 'Pending'
)
WHERE user_a_action_ref_id IS NULL;

-- Make columns NOT NULL after backfill
ALTER TABLE match_result
    ALTER COLUMN user_a_action_ref_id SET NOT NULL,
    ALTER COLUMN user_b_action_ref_id SET NOT NULL;

-- Index for querying user's matches efficiently
CREATE INDEX idx_match_result_user_a_dropped
    ON match_result (user_a_ref_id, is_dropped, is_expired);
CREATE INDEX idx_match_result_user_b_dropped
    ON match_result (user_b_ref_id, is_dropped, is_expired);
```

#### Column Semantics

| Column | Type | Purpose |
|--------|------|---------|
| `user_a_action_ref_id` | UUID FK | User A's action: Pending/Proposed/Passed |
| `user_a_action_at` | TIMESTAMPTZ | When User A took action |
| `user_a_seen_at` | TIMESTAMPTZ | When User A first viewed this match |
| `user_b_action_ref_id` | UUID FK | User B's action: Pending/Proposed/Passed |
| `user_b_action_at` | TIMESTAMPTZ | When User B took action |
| `user_b_seen_at` | TIMESTAMPTZ | When User B first viewed this match |
| `chat_unlocked_at` | TIMESTAMPTZ | When mutual proposal occurred (both Proposed) |
| `chat_closed_at` | TIMESTAMPTZ | When chat was ended by either user |
| `chat_closed_by` | UUID FK | Who closed the chat |
| `expires_at` | TIMESTAMPTZ | Calculated: `dropped_ts + match_expiration_hours` |

---

#### Migration 23: Create Chat Messages Table

```sql
-- 23_create_match_chat_message.up.sql

CREATE TABLE match_chat_message
(
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_result_id  UUID         NOT NULL,
    sender_id        UUID         NOT NULL,
    message          TEXT         NOT NULL,

    -- Read tracking
    seen_at          TIMESTAMPTZ,           -- When recipient saw it

    -- Edit/Delete support
    edited_at        TIMESTAMPTZ,
    deleted_at       TIMESTAMPTZ,

    -- Meta
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT chat_msg_match_fk FOREIGN KEY (match_result_id)
        REFERENCES match_result (id) ON DELETE CASCADE,
    CONSTRAINT chat_msg_sender_fk FOREIGN KEY (sender_id)
        REFERENCES users (id) ON DELETE RESTRICT
);

-- Index for fetching conversation
CREATE INDEX idx_chat_message_match_created
    ON match_chat_message (match_result_id, created_at);

-- Index for unread count
CREATE INDEX idx_chat_message_unseen
    ON match_chat_message (match_result_id, seen_at)
    WHERE seen_at IS NULL;
```

#### Column Semantics

| Column | Type | Purpose |
|--------|------|---------|
| `id` | UUID PK | Message identifier |
| `match_result_id` | UUID FK | Which match conversation this belongs to |
| `sender_id` | UUID FK | Who sent the message |
| `message` | TEXT | Message content (plain text, may add markdown later) |
| `seen_at` | TIMESTAMPTZ | When recipient viewed (null = unread) |
| `edited_at` | TIMESTAMPTZ | If message was edited |
| `deleted_at` | TIMESTAMPTZ | Soft delete for "deleted message" placeholder |
| `created_at` | TIMESTAMPTZ | Timestamp for ordering |

---

#### Migration 24: Create User Notifications Table

```sql
-- 24_create_user_notification.up.sql

-- Notification types
INSERT INTO category_type (name)
VALUES ('Notification Type');

INSERT INTO category (category_type_ref_id, name)
SELECT id, 'New Match Dropped'
FROM category_type WHERE name = 'Notification Type'
UNION ALL
SELECT id, 'Partner Proposed'
FROM category_type WHERE name = 'Notification Type'
UNION ALL
SELECT id, 'Chat Unlocked'
FROM category_type WHERE name = 'Notification Type'
UNION ALL
SELECT id, 'New Message'
FROM category_type WHERE name = 'Notification Type'
UNION ALL
SELECT id, 'Match Expiring Soon'
FROM category_type WHERE name = 'Notification Type'
UNION ALL
SELECT id, 'Chat Stale Nudge'
FROM category_type WHERE name = 'Notification Type';

CREATE TABLE user_notification
(
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID         NOT NULL,
    notification_type   UUID         NOT NULL,

    -- Reference to related entity
    ref_match_result_id UUID,                  -- For match-related notifications
    ref_message_id      UUID,                  -- For message notifications

    -- Content
    title               VARCHAR(255) NOT NULL,
    body                TEXT         NOT NULL,

    -- State
    seen_at             TIMESTAMPTZ,           -- When user tapped/viewed
    push_sent_at        TIMESTAMPTZ,           -- When push was sent
    push_failed         BOOLEAN      DEFAULT FALSE,

    -- Meta
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT notif_user_fk FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT notif_type_fk FOREIGN KEY (notification_type)
        REFERENCES category (id),
    CONSTRAINT notif_match_fk FOREIGN KEY (ref_match_result_id)
        REFERENCES match_result (id) ON DELETE SET NULL,
    CONSTRAINT notif_msg_fk FOREIGN KEY (ref_message_id)
        REFERENCES match_chat_message (id) ON DELETE SET NULL
);

-- Index for user's notification inbox
CREATE INDEX idx_notification_user_unseen
    ON user_notification (user_id, created_at DESC)
    WHERE seen_at IS NULL;
```

---

#### Migration 25: Add Push Token to Users

```sql
-- 25_add_user_push_token.up.sql

ALTER TABLE users
    ADD COLUMN push_token          VARCHAR(512),  -- FCM/APNS token
    ADD COLUMN push_token_platform VARCHAR(32),   -- 'ios', 'android', 'web'
    ADD COLUMN push_token_updated  TIMESTAMPTZ;
```

---

### Summary: New Tables & Columns

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          NEW SCHEMA (Migrations 22-25)                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ALTER match_result                 NEW: match_chat_message                     │
│  ────────────────────               ───────────────────────                     │
│  + user_a_action_ref_id             ├── id                                      │
│  + user_a_action_at                 ├── match_result_id (FK)                    │
│  + user_a_seen_at                   ├── sender_id (FK)                          │
│  + user_b_action_ref_id             ├── message                                 │
│  + user_b_action_at                 ├── seen_at                                 │
│  + user_b_seen_at                   ├── edited_at                               │
│  + chat_unlocked_at                 ├── deleted_at                              │
│  + chat_closed_at                   └── created_at                              │
│  + chat_closed_by                                                               │
│  + expires_at                       NEW: user_notification                      │
│                                     ────────────────────                        │
│  ALTER users                        ├── id                                      │
│  ────────────                       ├── user_id (FK)                            │
│  + push_token                       ├── notification_type (FK)                  │
│  + push_token_platform              ├── ref_match_result_id                     │
│  + push_token_updated               ├── ref_message_id                          │
│                                     ├── title, body                             │
│  NEW category_type                  ├── seen_at                                 │
│  ─────────────────                  ├── push_sent_at                            │
│  + 'Match User Action'              └── created_at                              │
│  + 'Notification Type'                                                          │
│                                                                                 │
│  NEW categories                                                                 │
│  ──────────────                                                                 │
│  Match User Action:                 Notification Type:                          │
│  + Pending                          + New Match Dropped                         │
│  + Proposed                         + Partner Proposed                          │
│  + Passed                           + Chat Unlocked                             │
│                                     + New Message                               │
│                                     + Match Expiring Soon                       │
│                                     + Chat Stale Nudge                          │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

### Query Examples After Schema Changes

#### Get User's Dropped Matches

```sql
SELECT
    mr.id,
    mr.qualifier_results->>'total_score' as match_percent,
    mr.expires_at,
    CASE
        WHEN mr.user_a_ref_id = :current_user_id THEN mr.user_a_action_ref_id
        ELSE mr.user_b_action_ref_id
    END as your_action_ref_id,
    CASE
        WHEN mr.user_a_ref_id = :current_user_id THEN mr.user_b_action_ref_id
        ELSE mr.user_a_action_ref_id
    END as their_action_ref_id,
    -- Partner details
    CASE
        WHEN mr.user_a_ref_id = :current_user_id THEN u_b.id
        ELSE u_a.id
    END as partner_id,
    CASE
        WHEN mr.user_a_ref_id = :current_user_id THEN u_b.first_name
        ELSE u_a.first_name
    END as partner_name
FROM match_result mr
JOIN users u_a ON mr.user_a_ref_id = u_a.id
JOIN users u_b ON mr.user_b_ref_id = u_b.id
WHERE
    (mr.user_a_ref_id = :current_user_id OR mr.user_b_ref_id = :current_user_id)
    AND mr.is_approved = true
    AND mr.is_dropped = true
    AND mr.is_expired = false
ORDER BY mr.dropped_ts DESC;
```

#### Check for Mutual Proposal (Chat Unlock)

```sql
-- When user proposes, check if chat should unlock
UPDATE match_result
SET
    user_a_action_ref_id = :proposed_category_id,
    user_a_action_at = NOW(),
    chat_unlocked_at = CASE
        WHEN user_b_action_ref_id = :proposed_category_id THEN NOW()
        ELSE NULL
    END
WHERE id = :match_id
  AND user_a_ref_id = :current_user_id
RETURNING chat_unlocked_at IS NOT NULL as is_mutual;
```

#### Get Chat Messages with Unread Count

```sql
SELECT
    m.*,
    (SELECT COUNT(*)
     FROM match_chat_message
     WHERE match_result_id = :match_id
       AND sender_id != :current_user_id
       AND seen_at IS NULL
    ) as unread_count
FROM match_chat_message m
WHERE m.match_result_id = :match_id
  AND m.deleted_at IS NULL
ORDER BY m.created_at ASC;
```

---

### Implementation Order

| Priority | Migration | Enables |
|----------|-----------|---------|
| **P0** | 22: Per-user match status | Propose/Pass endpoints, mutual detection |
| **P0** | 23: Chat messages | All chat endpoints |
| **P1** | 24: Notifications | Push notifications |
| **P1** | 25: Push tokens | Sending push to devices |

---

### Notes on Existing Tables

#### `user_match` Table - Clarification Needed

Migration 21 created `user_match`:
```sql
CREATE TABLE user_match (
    user_ref_id         UUID NOT NULL,
    match_result_ref_id UUID NOT NULL,
    ...
)
```

**Question:** What was the intended purpose?
- Current codebase doesn't seem to use it
- Might have been intended for per-user match tracking
- Consider dropping if redundant with new `user_a_action`/`user_b_action` columns

**Recommendation:** Keep for now, but document as deprecated. The new columns on `match_result` are cleaner because:
1. No need for separate queries to join
2. Atomic updates (both users' data in one row)
3. Easier to check mutual proposal status
