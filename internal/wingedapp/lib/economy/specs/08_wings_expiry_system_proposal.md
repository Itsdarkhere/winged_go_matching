# Wings Expiry System Proposal

**STATUS: IMPLEMENTED**

## TL;DR

The spec requires **expiry timestamps** on all wings. Current implementation has **no expiry** - wings never expire. This is a **medium-sized change** affecting DB schema, stores, and spend logic.

---

## Implementation Summary

### Files Changed

| File | Change |
|------|--------|
| `migration/04_wings_economy.up.sql` | Added `expires_at TIMESTAMPTZ`, `is_expired BOOLEAN DEFAULT FALSE`, index |
| `db/repo/wings_ecn_transaction.go` | Added `ExpiresAt null.Time` to insert struct |
| `lib/economy/models.go` | Added `ExpiresAt` to `InsertTransaction`, `InsertCheckinTransaction` |
| `lib/economy/consts.go` | Added `EarnedWingsExpiryDays = 30` |
| `lib/economy/store/transaction.go` | Pass `ExpiresAt` through to repo |
| `lib/economy/store/daily_checkin.go` | Pass `ExpiresAt` through to repo |
| `lib/economy/daily_checkin.go` | Set `ExpiresAt` to 30 days on insert |
| `lib/economy/referral.go` | Set `ExpiresAt` to 30 days on all 3 inserts |
| `lib/economy/attend_date.go` | Set `ExpiresAt` to 30 days on insert |
| `lib/economy/expiry.go` | **NEW** - `ExpiryLogic.ExpireWings()` |
| `lib/economy/store/expiry.go` | **NEW** - `ExpiryStore` with queries |
| `lib/economy/store/store.go` | Added `ExpiryStore` to `EconomyStores` |
| `di/deps_lib.go` | Added `libEconomyExpiry` DI definition |
| `cmd/wingedapp/api/main.go` | Added hourly cron job for expiry |

### How It Works

1. **On wing credit**: `expires_at = NOW() + 30 days` (for earned wings)
2. **Cron job (hourly)**:
   - Query transactions where `expires_at < NOW() AND is_expired = FALSE`
   - Sum amounts per user
   - Decrement `user_totals.total_wings`
   - Set `is_expired = TRUE`
3. **Atomic**: All wrapped in single transaction

### Post-Deploy Steps

```bash
# 1. Regenerate pgmodel (new columns)
make wingedapp-pgmodel-gen

# 2. Regenerate DI (new ExpiryLogic)
make wingedapp-di-generate

# 3. Run tests
go test -v ./internal/wingedapp/lib/economy/...
```

---

## Gap Analysis: Spec vs Current Implementation

### What the Spec Says

| Wing Source | Expiry Rule |
|-------------|-------------|
| Subscription (Winged+) | Expires at billing cycle end |
| New user bonus | 30 days |
| Streak milestones | 30 days |
| Referrals | 30 days |
| Date feedback | 30 days |
| WingedX | No wings, no expiry (0 cost) |

### What's Currently Implemented

| Component | Status | Gap |
|-----------|--------|-----|
| `wings_ecn_transaction` | No `expires_at` column | **MISSING** |
| `wings_ecn_user_totals.total_wings` | Single aggregate counter | **MISSING expiry-aware balance** |
| Spend logic | Debits from counter directly | **MISSING FIFO expiry logic** |
| Earn logic | No expiry set on insert | **MISSING** |
| Background job | None | **MISSING expiry cleanup cron** |

---

## Required Changes

### 1. Database Migration (~50 lines)

```sql
-- Migration XX: Add expiry support to wings economy

-- Add expires_at to transaction table
ALTER TABLE wings_ecn_transaction
    ADD COLUMN expires_at TIMESTAMPTZ DEFAULT NULL;

-- Add index for efficient expiry queries
CREATE INDEX idx_wings_ecn_transaction_expires_at
    ON wings_ecn_transaction (expires_at)
    WHERE expires_at IS NOT NULL AND is_active = 1;

-- Add subscription period tracking
ALTER TABLE wings_ecn_user_subscription_plan
    ADD COLUMN period_start TIMESTAMPTZ,
    ADD COLUMN period_end TIMESTAMPTZ;

-- View: Active (non-expired) balance per user
CREATE OR REPLACE VIEW view_wings_active_balance AS
SELECT
    user_ref_id AS user_id,
    COALESCE(SUM(
        CASE
            WHEN is_credit THEN amount
            ELSE -amount
        END
    ), 0) AS balance
FROM wings_ecn_transaction
WHERE is_active = 1
  AND claimed = TRUE
  AND (expires_at IS NULL OR expires_at > NOW())
GROUP BY user_ref_id;
```

**Effort: ~2 hours**

---

### 2. Model Changes (`lib/economy/models.go`)

```go
// Add to InsertTransaction
type InsertTransaction struct {
    // ... existing fields ...
    ExpiresAt null.Time // NEW: expiry timestamp
}

// Add to Transaction
type Transaction struct {
    // ... existing fields ...
    ExpiresAt null.Time `boil:"expires_at"`
}

// New: expiry calculation helpers
type ExpirySource string

const (
    ExpirySourceSubscription  ExpirySource = "subscription"
    ExpirySourceEarned        ExpirySource = "earned"     // 30 days
    ExpirySourceNewUserBonus  ExpirySource = "new_user"   // 30 days
)

const EarnedWingsExpiryDays = 30
```

**Effort: ~1 hour**

---

### 3. Store Changes (`lib/economy/store/transaction.go`)

Add `expires_at` to insert/select queries. Update filters to exclude expired wings.

```go
// In Transactions() - add expiry filter option
type QueryFilterTransactions struct {
    // ... existing ...
    ExcludeExpired null.Bool // if true, WHERE expires_at IS NULL OR expires_at > NOW()
}

// In Insert() - handle expires_at column
```

**Effort: ~2 hours**

---

### 4. Balance Calculation Change

**Current:** `wings_ecn_user_totals.total_wings` (single counter)

**New:** Query `view_wings_active_balance` OR compute from transactions with expiry filter

**Option A: Keep counter + sync job**
- Pro: Fast reads
- Con: Counter drift risk, more complexity

**Option B: Always compute from transactions (RECOMMENDED)**
- Pro: Single source of truth, no drift
- Con: Slightly slower reads (mitigated by index)

```go
// GetActiveBalance replaces reading from user_totals
func (s *Store) GetActiveBalance(ctx context.Context, exec boil.ContextExecutor, userID string) (int, error) {
    // Use view_wings_active_balance or inline query
}
```

**Effort: ~3 hours**

---

### 5. Spend Logic Changes (`spend_wings`)

**Current:** Debit from counter

**New:** FIFO debit from oldest non-expired transactions

```go
// SpendWings debits from oldest expiring wings first (FIFO)
func (s *SpendLogic) SpendWings(ctx context.Context, exec boil.ContextExecutor, userID string, amount int, action ActionType) error {
    // 1. Check tier - WingedX skips spending
    // 2. Get active balance (expiry-aware)
    // 3. If insufficient, return ErrInsufficientWings
    // 4. Insert negative transaction (debit)
    // 5. Log action
}
```

**Effort: ~4 hours**

---

### 6. Earn Logic Changes

Each earn source must set appropriate `expires_at`:

```go
func (e *EarnLogic) EarnWings(ctx context.Context, exec boil.ContextExecutor, params EarnParams) error {
    expiresAt := calculateExpiry(params.Source, params.SubscriptionPeriodEnd)
    // Insert with expires_at
}

func calculateExpiry(source ExpirySource, subscriptionEnd *time.Time) null.Time {
    switch source {
    case ExpirySourceSubscription:
        if subscriptionEnd == nil {
            return null.Time{} // shouldn't happen
        }
        return null.TimeFrom(*subscriptionEnd)
    case ExpirySourceEarned, ExpirySourceNewUserBonus:
        return null.TimeFrom(time.Now().AddDate(0, 0, EarnedWingsExpiryDays))
    default:
        return null.Time{} // no expiry
    }
}
```

**Effort: ~2 hours**

---

### 7. Background Job: Expiry Cleanup

Soft-delete or mark expired transactions:

```go
// Cron: runs hourly or daily
func (j *ExpiryJob) CleanupExpired(ctx context.Context) error {
    // UPDATE wings_ecn_transaction
    // SET is_active = 0
    // WHERE expires_at < NOW() AND is_active = 1
}
```

**Note:** Not strictly required if balance calculation filters by expiry. But useful for DB hygiene.

**Effort: ~2 hours**

---

### 8. Streak System (NEW - not currently implemented)

Spec requires streak tracking on `profiles` table:
- `streak_last_date`
- `streak_current_days`
- `streak_longest_days`

**Current:** Daily check-in exists but no streak tracking.

**Gap:** Need to add streak columns and milestone reward logic.

**Effort: ~4 hours (if implementing)**

---

## Summary: Effort Estimate

| Component | Hours | Priority |
|-----------|-------|----------|
| Migration (expires_at column) | 2 | P0 |
| Model changes | 1 | P0 |
| Store changes | 2 | P0 |
| Balance calculation | 3 | P0 |
| Spend logic (FIFO) | 4 | P0 |
| Earn logic (set expiry) | 2 | P0 |
| Expiry cleanup job | 2 | P1 |
| Streak system | 4 | P2 |
| Tests | 6 | P0 |
| **Total** | **~26 hours** | |

---

## Recommended Implementation Order

1. **Migration** - Add `expires_at` column, create view
2. **Store** - Update insert/select to handle expiry
3. **Earn** - Set expiry on all wing grants
4. **Balance** - Switch to expiry-aware balance calculation
5. **Spend** - No FIFO needed if using transactions (just check balance)
6. **Tests** - Table-driven tests for expiry scenarios
7. **Cleanup job** - Optional, add later

---

## Questions for Product

1. **Existing wings:** What happens to wings already in the system with no expiry?
   - Option A: Grandfather them (never expire)
   - Option B: Set expiry to 30 days from migration date
   - Option C: Set expiry to 30 days from original created_date

2. **Balance display:** Should UI show "X wings expiring in Y days"?

3. **Streak system priority:** MVP or post-MVP?

---

## Files to Change

```
internal/wingedapp/migration/XX_wings_expiry.up.sql        (NEW)
internal/wingedapp/migration/XX_wings_expiry.down.sql      (NEW)
internal/wingedapp/lib/economy/models.go                   (MODIFY)
internal/wingedapp/lib/economy/consts.go                   (MODIFY)
internal/wingedapp/lib/economy/store/transaction.go        (MODIFY)
internal/wingedapp/lib/economy/store/user_totals.go        (MODIFY or DEPRECATE)
internal/wingedapp/lib/economy/spend_wings.go              (NEW or MODIFY existing)
internal/wingedapp/lib/economy/earn_wings.go               (NEW or MODIFY existing)
internal/wingedapp/db/pgmodel/*                            (REGENERATE after migration)
```

---

## Decision Points

| Decision | Options | Recommendation |
|----------|---------|----------------|
| Balance source | Counter vs Computed | **Computed** (single source of truth) |
| FIFO spend | Track per-transaction vs Sum only | **Sum only** (simpler, spec doesn't require FIFO) |
| Existing wings | Grandfather vs Expire | **Ask product** |
| Streak | MVP vs Post-MVP | **Post-MVP** (daily check-in works without it) |
