# Phase 2: Subscription Wings (Winged+ vs WingedX)

## TL;DR
**Winged+** (1.1): Wings credited immediately, expire at END of billing cycle (NOT 30 days). NO rollover. **WingedX** (1.2): 0 wings, all actions free. Current DB tracks subscriptions but missing: billing cycle expiry logic, WingedX zero-wing tier handling, subscription snapshot table.

---

## Spec Breakdown: Winged+ (Section 1.1)

| Feature | Spec Requirement | Current Implementation | Status |
|---------|------------------|------------------------|--------|
| **Wings Credited** | Immediately on purchase | Unknown (not in code reviewed) | ❓ UNKNOWN |
| **Expiry Logic** | End of billing cycle (NOT 30 days) | All earned wings expire in 30 days | ❌ WRONG |
| **Rollover** | Do NOT roll over | No rollover logic found | ⚠️ UNCONFIRMED |
| **Tiers** | Weekly (15), Monthly (45), 3mo (135), 6mo (270) | Plans exist in consts but wings not specified | ⚠️ INCOMPLETE |
| **Billing Cycle Tracking** | Need period_start, period_end | `wings_ecn_user_subscription_plan` has start_date, end_date | ✅ EXISTS |
| **Snapshot Table** | `subscription_snapshots` per spec Section 5 | No table found | ❌ MISSING |

---

## Spec Breakdown: WingedX (Section 1.2)

| Feature | Spec Requirement | Current Implementation | Status |
|---------|------------------|------------------------|--------|
| **Wings Allocation** | 0 wings (all actions free) | No special handling for 0-wing plans | ❌ MISSING |
| **Free Actions** | All economy actions cost 0 | `SendMessage` deducts wings regardless of plan | ❌ WRONG |
| **Expiry** | No wings = no expiry | N/A | ✅ N/A |
| **Tiers** | Weekly, Monthly | Consts exist but no 0-wing enforcement | ❌ MISSING |

---

## Current Database Schema

### `wings_ecn_subscription_plan` (exists)
```go
type WingsEcnSubscriptionPlan struct {
    ID               string
    SubscriptionType string        // "Winged+", "WingedX", "Top Up"
    Name             string
    Price            types.Decimal
    Wings            int            // ❌ PROBLEM: WingedX should have 0, but logic doesn't check
    IsActive         null.Int
}
```

### `wings_ecn_user_subscription_plan` (exists)
```go
type WingsEcnUserSubscriptionPlan struct {
    ID                 string
    UserID             string
    SubscriptionPlanID string
    StartDate          time.Time  // ✅ Billing cycle start
    EndDate            time.Time  // ✅ Billing cycle end
    IsActive           null.Int
}
```

### `wings_ecn_transaction` (exists)
```go
type WingsEcnTransaction struct {
    ID             string
    ActionLogType  string
    UserRefID      string
    ActionLogRefID string
    IsCredit       bool
    Claimed        bool
    Amount         int
    ExpiresAt      null.Time  // ❌ PROBLEM: Set to +30 days for earned wings
                                // ❌ MISSING: Should be set to subscription.EndDate for Winged+ wings
    IsExpired      bool
}
```

### `subscription_snapshots` (MISSING - per spec Section 5)
**Spec requirement:**
```sql
CREATE TABLE subscription_snapshots (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    tier VARCHAR(50) NOT NULL,        -- "Winged+ Weekly", "Winged+ Monthly", etc.
    wings_granted INTEGER NOT NULL,   -- Wings allocated for this period
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Purpose:** Historical record of subscription periods & wings granted (for auditing, analytics, billing disputes).

---

## Required Changes

### 1. Database Schema Changes

#### Add `subscription_snapshots` table
```sql
CREATE TABLE subscription_snapshots (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES users(id),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    tier VARCHAR(100) NOT NULL,
    wings_granted INTEGER NOT NULL,
    is_active INTEGER DEFAULT 1,
    created_by INTEGER,
    created_date TIMESTAMP DEFAULT NOW(),
    last_updated TIMESTAMP DEFAULT NOW(),
    updated_by INTEGER
);

CREATE INDEX idx_subscription_snapshots_user ON subscription_snapshots(user_id);
CREATE INDEX idx_subscription_snapshots_period ON subscription_snapshots(period_start, period_end);
```

#### Add subscription plan tier configuration
Update `wings_ecn_subscription_plan` seed data or add constant mappings:

```sql
-- Winged+ Weekly: 15 wings
-- Winged+ Monthly: 45 wings
-- Winged+ 3 Months: 135 wings (45 * 3)
-- Winged+ 6 Months: 270 wings (45 * 6)
-- WingedX Weekly: 0 wings
-- WingedX Monthly: 0 wings
```

### 2. Code Changes

#### `lib/economy/consts.go` - Add Wing Allocations
```go
// Winged+ Wing Allocations (per spec Section 1.1)
const WingedPlusWeeklyWings = 15
const WingedPlusMonthlyWings = 45
const WingedPlusThreeMonthWings = 135
const WingedPlusSixMonthWings = 270

// WingedX Wing Allocations (per spec Section 1.2)
const WingedXWeeklyWings = 0
const WingedXMonthlyWings = 0
```

#### `lib/economy/models.go` - Add Snapshot Model
```go
// SubscriptionSnapshot for auditing subscription periods
type SubscriptionSnapshot struct {
    ID           string    `boil:"id"`
    UserID       string    `boil:"user_id"`
    PeriodStart  time.Time `boil:"period_start"`
    PeriodEnd    time.Time `boil:"period_end"`
    Tier         string    `boil:"tier"`
    WingsGranted int       `boil:"wings_granted"`
}

// InsertSubscriptionSnapshot for creating snapshots
type InsertSubscriptionSnapshot struct {
    UserID       string
    PeriodStart  time.Time
    PeriodEnd    time.Time
    Tier         string
    WingsGranted int
}

// QueryFilterSubscriptionSnapshot for filtering snapshots
type QueryFilterSubscriptionSnapshot struct {
    UserID null.String
    Tier   null.String
}
```

#### NEW: `lib/economy/subscription.go` - Subscription Logic
```go
package economy

import (
    "context"
    "fmt"
    "time"
    "github.com/aarondl/null/v8"
    "github.com/aarondl/sqlboiler/v4/boil"
)

type SubscriptionLogic struct {
    logger          applog.Logger
    subscriptionStore subscriptionStorer
    snapshotStore   snapshotStorer
    transactionRepo transactionInserter
    userTotalsStore userTotalsStorer
}

// ProcessSubscriptionPurchase handles Winged+ purchase
func (s *SubscriptionLogic) ProcessSubscriptionPurchase(
    ctx context.Context,
    exec boil.ContextExecutor,
    userID string,
    planID string,
) error {
    // 1. Get plan details
    plan, err := s.subscriptionStore.Plan(ctx, exec, planID)
    if err != nil {
        return fmt.Errorf("get plan: %w", err)
    }

    // 2. Calculate billing period
    var periodStart, periodEnd time.Time
    switch plan.Name {
    case SubscriptionPaymentWeekly:
        periodStart = time.Now()
        periodEnd = periodStart.AddDate(0, 0, 7)
    case SubscriptionPaymentMonthly:
        periodStart = time.Now()
        periodEnd = periodStart.AddDate(0, 1, 0)
    case SubscriptionPaymentThreeMonth:
        periodStart = time.Now()
        periodEnd = periodStart.AddDate(0, 3, 0)
    case SubscriptionPaymentSixMonth:
        periodStart = time.Now()
        periodEnd = periodStart.AddDate(0, 6, 0)
    default:
        return fmt.Errorf("unknown subscription period: %s", plan.Name)
    }

    // 3. Create user subscription record
    _, err = s.subscriptionStore.Insert(ctx, exec, &InsertUserSubscription{
        UserID:             userID,
        SubscriptionPlanID: planID,
        StartDate:          periodStart,
        EndDate:            periodEnd,
    })
    if err != nil {
        return fmt.Errorf("insert user subscription: %w", err)
    }

    // 4. Create subscription snapshot (audit trail)
    _, err = s.snapshotStore.Insert(ctx, exec, &InsertSubscriptionSnapshot{
        UserID:       userID,
        PeriodStart:  periodStart,
        PeriodEnd:    periodEnd,
        Tier:         fmt.Sprintf("%s - %s", plan.SubscriptionType, plan.Name),
        WingsGranted: plan.Wings,
    })
    if err != nil {
        return fmt.Errorf("insert snapshot: %w", err)
    }

    // 5. Grant wings IMMEDIATELY (if Winged+)
    if plan.SubscriptionType == SubscriptionTypeWingedPlus && plan.Wings > 0 {
        // Create transaction with expiry = end of billing cycle (NOT +30 days)
        actionLog, err := s.actionLogStore.Insert(ctx, exec, string(planActionType), &InsertActionLog{
            UserID: userID,
            RefID:  planID,
            Type:   planActionType, // e.g., ActionWingedPlusMonthlyPayment
        })
        if err != nil {
            return fmt.Errorf("insert action log: %w", err)
        }

        _, err = s.transactionRepo.Insert(ctx, exec, &InsertTransaction{
            UserID:       userID,
            ActionTypeID: string(planActionType),
            ActionRefID:  actionLog.ID,
            WingsAmount:  plan.Wings,
            IsCredit:     true,
            Claimed:      true, // Auto-claimed for subscriptions
            ExpiresAt:    null.TimeFrom(periodEnd), // ✅ SPEC: Expires at END of billing cycle
        })
        if err != nil {
            return fmt.Errorf("insert transaction: %w", err)
        }

        // Credit to balance immediately
        totals, err := s.userTotalsStore.Totals(ctx, exec, userID)
        if err != nil {
            return fmt.Errorf("get totals: %w", err)
        }
        err = s.userTotalsStore.Update(ctx, exec, &UpdateUserTotals{
            ID:    totals.ID,
            Wings: null.IntFrom(totals.Wings + plan.Wings),
        })
        if err != nil {
            return fmt.Errorf("update totals: %w", err)
        }
    }

    // 6. For WingedX: No wings granted, just activate subscription
    // All actions will check subscription and bypass wing costs

    return nil
}
```

#### `lib/economy/send_message.go` - Add WingedX Check
```go
// SendMessage deducts wings for sending messages
func (s *SendMessageLogic) SendMessage(
    ctx context.Context,
    exec boil.ContextExecutor,
    userID string,
) error {
    // 1. Check if user has WingedX subscription
    hasWingedX, err := s.subscriptionStore.HasActiveSubscription(ctx, exec, userID, SubscriptionTypeWingedX)
    if err != nil {
        return fmt.Errorf("check wingedx: %w", err)
    }

    // 2. WingedX users: all actions free, skip wing deduction
    if hasWingedX {
        // Still track action in logs, but no transaction
        _, err := s.actionLogStore.Insert(ctx, exec, string(ActionSendMessage), &InsertActionLog{
            UserID: userID,
            RefID:  "free_wingedx",
            Type:   ActionSendMessage,
        })
        return err
    }

    // 3. Non-WingedX: deduct wings as normal
    // ... existing logic ...
}
```

### 3. Store Layer Changes

#### NEW: `lib/economy/store/subscription_snapshot.go`
```go
package store

type SubscriptionSnapshotStore struct {
    logger applog.Logger
    repo   snapshotInserter
}

func (s *SubscriptionSnapshotStore) Insert(
    ctx context.Context,
    exec boil.ContextExecutor,
    input *InsertSubscriptionSnapshot,
) (*SubscriptionSnapshot, error) {
    // Call repo layer for insert
    // Return domain type
}

func (s *SubscriptionSnapshotStore) Snapshots(
    ctx context.Context,
    exec boil.ContextExecutor,
    filter *QueryFilterSubscriptionSnapshot,
) ([]SubscriptionSnapshot, error) {
    // Query with filters
    // Return domain types
}
```

#### Update: `lib/economy/store/subscription_plan.go`
Add method:
```go
// HasActiveSubscription checks if user has active subscription of given type
func (s *SubscriptionPlanStore) HasActiveSubscription(
    ctx context.Context,
    exec boil.ContextExecutor,
    userID string,
    subscriptionType string,
) (bool, error) {
    qMods := []qm.QueryMod{
        qm.InnerJoin("wings_ecn_user_subscription_plan usp ON usp.subscription_plan_id = wings_ecn_subscription_plan.id"),
        pgmodel.WingsEcnUserSubscriptionPlanWhere.UserID.EQ(userID),
        pgmodel.WingsEcnSubscriptionPlanWhere.SubscriptionType.EQ(subscriptionType),
        pgmodel.WingsEcnUserSubscriptionPlanWhere.EndDate.GT(time.Now()), // Still active
        pgmodel.WingsEcnUserSubscriptionPlanWhere.IsActive.EQ(null.IntFrom(1)),
    }
    count, err := pgmodel.NewQuery(qMods...).Count(ctx, exec)
    return count > 0, err
}
```

---

## Winged+ vs WingedX Comparison Table

| Feature | Winged+ (1.1) | WingedX (1.2) |
|---------|---------------|---------------|
| **Wings Granted** | Weekly: 15, Monthly: 45, 3mo: 135, 6mo: 270 | 0 (none) |
| **Wings Credited** | Immediately on purchase | N/A |
| **Wings Expiry** | End of billing cycle (period_end) | N/A |
| **Rollover** | No rollover | N/A |
| **Action Costs** | Normal (5 messages = 1 wing, etc.) | All actions FREE |
| **Transaction Created** | Yes (with expires_at = period_end) | No (or 0-amount for tracking) |
| **Snapshot Required** | Yes | Yes (for audit trail) |

---

## Implementation Steps

### Step 1: DB Migration
1. Create `subscription_snapshots` table
2. Seed `wings_ecn_subscription_plan` with correct wing amounts
3. Regenerate pgmodel: `make wingedapp-sqlboiler`

### Step 2: Add Models & Constants
1. Add `WingedPlus*Wings` and `WingedX*Wings` constants
2. Add `SubscriptionSnapshot` models
3. Add `InsertSubscriptionSnapshot` model

### Step 3: Create Snapshot Store
1. `lib/economy/store/subscription_snapshot.go`
2. Implement `Insert()`, `Snapshots()` methods
3. Add interface to `lib/economy/adapters.go`

### Step 4: Create Subscription Logic
1. `lib/economy/subscription.go`
2. Implement `ProcessSubscriptionPurchase()`
3. Handle Winged+ wing granting with billing cycle expiry
4. Handle WingedX zero-wing activation

### Step 5: Update Spending Logic
1. Add `HasActiveSubscription()` to subscription store
2. Update `SendMessage()` to check WingedX and skip costs
3. Update any other action costs to check WingedX

### Step 6: Fix Expiry Logic
1. **Earned Wings:** `expires_at = +30 days` (already correct)
2. **Subscription Wings:** `expires_at = subscription.end_date` (NEW)
3. Ensure expiry cron job respects both types

### Step 7: Testing
1. Test Winged+ purchase → wings credited immediately
2. Test wings expire at end_date (not +30 days)
3. Test no rollover (wings lost at period end)
4. Test WingedX purchase → 0 wings, free actions
5. Test snapshot creation for audit trail

---

## Pattern Recognition for Codebase Learning

### Lesson 1: Dual Expiry Logic Required
- **Earned wings:** Fixed 30-day expiry (global rule)
- **Subscription wings:** Variable expiry tied to billing cycle (subscription-specific)
- Current implementation only supports fixed expiry

### Lesson 2: Zero-Cost Actions Require Guards
- Current: all actions deduct wings blindly
- WingedX: requires checking subscription BEFORE deducting
- Pattern: `hasWingedX ? skip_cost : deduct_wings`

### Lesson 3: Audit Trail via Snapshots
- Spec requires historical record (snapshots) separate from active state (user_subscription_plan)
- Why: Billing disputes, analytics, subscription history
- Pattern: Active state + immutable audit log

### Lesson 4: Immediate Credit vs Claim Mechanic
- Daily check-in: accumulate → claim
- Subscription: immediate credit (no claim step)
- Different crediting patterns for different sources

### Lesson 5: Subscription Type as Behavior Flag
- `SubscriptionType` field is NOT just metadata
- It determines: wing allocation, expiry logic, action cost behavior
- Must be checked in business logic, not just stored

---

## Questions for Product

1. **Rollover clarification:** Spec says "do NOT roll over". Does this mean:
   - Wings expire exactly at `period_end` (even if unused)?
   - OR wings expire but user gets fresh allocation on renewal?

2. **Renewal behavior:** When Winged+ renews (e.g., monthly → next month):
   - Create new snapshot?
   - Create new transaction with fresh wings?
   - Expire old wings + grant new wings?

3. **Downgrade/Upgrade:** User switches from Winged+ to WingedX mid-cycle:
   - Do unused Winged+ wings persist until period_end?
   - OR immediately invalidated?

4. **WingedX tracking:** Should we still create 0-amount transactions for audit? Or skip entirely?

5. **Mixed subscriptions:** Can user have both Winged+ and WingedX? Or mutually exclusive?

6. **Top-Up plans:** Spec mentions "Top Up - Mini/Boost/Premium" in consts. How do these differ from Winged+?
   - Do they have expiry?
   - Do they roll over?
   - Clarify expiry rules for Top-Up wings.
