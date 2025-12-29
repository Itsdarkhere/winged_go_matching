# Phase 3: Referral System Reconciliation

## TL;DR
**Current implementation is PARTIALLY CORRECT** for "two-phase" model. MVP spec says only ONE trigger: both get 5 wings when invitee completes onboarding. Need to simplify to single `ActionReferralComplete` and remove unused `ActionReferralSignup`.

---

## Spec vs Current Implementation

| Aspect | MVP Spec (refer_a_friend.md) | Current Implementation | Status |
|--------|------------------------------|------------------------|--------|
| **Trigger Event** | Deploy agent (single trigger) | Phase 1: Signup, Phase 2: Deploy | ⚠️ OVER-ENGINEERED |
| **Wings Amount** | 5 each (inviter + invitee) | `ReferralBonusWings = 5` | ✅ CORRECT |
| **Inviter Credit** | When invitee deploys | Phase 1: signup, Phase 2: deploy | ⚠️ DOUBLE CREDITS POSSIBLE |
| **Invitee Credit** | When invitee deploys | Phase 2 only | ✅ CORRECT |
| **Idempotency** | Check action_log | Implemented | ✅ CORRECT |
| **Expiry** | 30 days (earned wings) | `EarnedWingsExpiryDays = 30` | ✅ CORRECT |
| **Transaction** | `claimed: true` (auto-credit) | `Claimed: true` | ✅ CORRECT |
| **Atomic Update** | `total_wings = total_wings + 5` | Uses computed value | ⚠️ RACE CONDITION |

---

## Current Code Analysis

### Constants (`consts.go`)
```go
ActionReferralSignup   ActionType = "Referral - Friend Signup"   // Phase 1 - NOT IN MVP SPEC
ActionReferralComplete ActionType = "Referral - Friend Complete" // Phase 2 - THIS IS MVP
ReferralBonusWings = 5  // ✅ Correct
```

### Two-Phase Model (`referral.go`)

**Phase 1 - `processReferralSignup()`:**
- Triggered when invitee enters invite code
- Credits ONLY referrer (5 wings)
- **Problem:** MVP spec doesn't have this phase

**Phase 2 - `processReferralBonus()`:**
- Triggered when invitee deploys agent
- Credits BOTH referrer (5 wings) AND invitee (5 wings)
- **Problem:** If Phase 1 ran, referrer gets DOUBLE (10 wings total)

### The Bug
```
User A invites User B

Phase 1 (signup):
  - User B enters invite code
  - User A (referrer) gets +5 wings  ← NOT IN MVP SPEC

Phase 2 (deploy agent):
  - User B deploys agent
  - User A (referrer) gets +5 wings  ← MVP SPEC
  - User B (invitee) gets +5 wings   ← MVP SPEC

RESULT: User A has 10 wings (should have 5)
```

---

## Required Changes

### Option A: Remove Phase 1 Entirely (Recommended)

**Simplest fix - aligns with MVP spec:**

1. **Remove `ActionReferralSignup`** from consts
2. **Remove `processReferralSignup()`** function
3. **Remove routing** for `ActionReferralSignup` in entrypoints
4. **Keep `processReferralBonus()`** as sole implementation
5. **Verify hook location** - only in `DeployUserAgent()`

### Option B: Keep Both But Fix Double-Credit

**If product wants early referrer notification:**

1. Phase 1: Create action_log (tracking only), NO wings
2. Phase 2: Credit both users 5 wings each
3. Use different ActionType for tracking vs crediting

---

## Implementation Steps (Option A)

### Step 1: Update Constants
```go
// lib/economy/consts.go

// REMOVE:
// ActionReferralSignup ActionType = "Referral - Friend Signup"

// KEEP:
ActionReferralComplete ActionType = "Referral - Friend Complete"
ReferralBonusWings = 5
```

### Step 2: Update Referral Logic
```go
// lib/economy/referral.go

// DELETE: processReferralSignup() function entirely

// KEEP: processReferralBonus() - this is the MVP implementation
// But ADD idempotency check for referrer (currently missing)
```

### Step 3: Fix Race Condition
Current code:
```go
// ❌ WRONG - read then write allows race condition
referrerTotals, err := a.userTotalsStorer.Totals(ctx, exec, referrer.ID)
newWings := referrerTotals.Wings + ReferralBonusWings
err := a.userTotalsStorer.Update(ctx, exec, &UpdateUserTotals{
    ID:    referrerTotals.ID,
    Wings: null.IntFrom(newWings),
})
```

Should be:
```go
// ✅ CORRECT - atomic increment
err := a.userTotalsStorer.IncrementWings(ctx, exec, referrer.ID, ReferralBonusWings)
// SQL: UPDATE wings_ecn_user_totals SET total_wings = total_wings + $1 WHERE user_ref_id = $2
```

### Step 4: Update Entrypoints
```go
// lib/economy/entrypoints.go

actionLoggerHandlers := map[ActionType]actLoggerHandlerFn{
    // REMOVE: ActionReferralSignup: a.processReferralSignup,
    ActionReferralComplete: a.processReferralBonus,  // KEEP
}
```

### Step 5: Verify Hook Location
```go
// business/domain/registration/agent.go - DeployUserAgent()
// This is the ONLY place referral should be triggered

if user.UserInviteCodeRefID.Valid {
    err := a.actionLogger.CreateActionLog(ctx, tx, &economy.InsertActionLog{
        UserID: user.ID,
        RefID:  user.UserInviteCodeRefID.String,
        Type:   economy.ActionReferralComplete,  // ✅ Correct type
    })
    // ...
}
```

### Step 6: Remove Signup Hook

**CONFIRMED: Phase 1 IS hooked in production code!**

Location: `business/domain/registration/registration.go:448-455`
```go
// Award referrer when user enters invite code (Phase 1 of referral rewards)
if err := b.actionLogger.CreateActionLog(ctx, v.tx, &economy.InsertActionLog{
    UserID: user.ID,
    RefID:  v.userInviteCode.ID,
    Type:   economy.ActionReferralSignup,  // ← REMOVE THIS BLOCK
}); err != nil {
    return fmt.Errorf("award referrer on signup: %w", err)
}
```

**Action:** Remove this entire block from `registration.go`. Phase 2 in `agent.go` is the only trigger we need.

---

## Database Impact

### No Schema Changes Required
Existing tables support the MVP:
- `wings_ecn_action_log` - tracks referral events
- `wings_ecn_transaction` - records wing credits
- `wings_ecn_user_totals` - holds balance

### Data Cleanup (Optional)
If Phase 1 already credited wings in production:
```sql
-- Find referrers who got double-credited
SELECT user_ref_id, COUNT(*) as credit_count
FROM wings_ecn_action_log
WHERE action_log_type IN ('Referral - Friend Signup', 'Referral - Friend Complete')
GROUP BY user_ref_id
HAVING COUNT(*) > 1;

-- Manual cleanup required if found
```

---

## Test Updates

### Remove Phase 1 Tests
- Any test for `ActionReferralSignup`
- Any test for `processReferralSignup`

### Keep/Update Phase 2 Tests
Current tests in `referral_test.go` are good:
- `success-idempotency-invitee-only-credited-once` ✅
- `success-idempotency-referrer-only-credited-once` ✅
- `success-both-users-credited-with-existing-wings` ✅

### Add Missing Tests
| Test Case | Setup | Assert |
|-----------|-------|--------|
| `error-no-double-credit-on-retry` | Call 5 times | Each user still has 5 wings |
| `success-concurrent-calls` | Parallel goroutines | No race condition, correct totals |
| `success-referrer-deleted` | Delete referrer user | Invitee still gets 5, no error |

---

## Files to Change

| File | Change |
|------|--------|
| `lib/economy/consts.go` | Remove `ActionReferralSignup` |
| `lib/economy/referral.go` | Delete `processReferralSignup()`, keep `processReferralBonus()` |
| `lib/economy/entrypoints.go` | Remove `ActionReferralSignup` from handler map |
| `lib/economy/store/user_totals.go` | Add `IncrementWings()` for atomic updates |
| `lib/economy/referral_test.go` | Update tests if any test Phase 1 |
| `business/domain/registration/registration.go` | **Remove lines 448-455** (Phase 1 hook) |

---

## Edge Cases

| Case | Handling |
|------|----------|
| User invites themselves | Prevented at invite code level (not economy concern) |
| Referrer account deleted | Credit invitee only, log warning, don't fail |
| Invitee never deploys | No wings credited (correct per MVP) |
| Deploy agent fails | TX rollback, no wings credited |
| User changes invite code | Uses code at deployment time |

---

## Summary

**MVP Requirement:** Single trigger (deploy agent) → both get 5 wings.

**Current State:** Two triggers exist (signup + deploy) → referrer gets 10 wings (BUG).

**Fix:** Remove Phase 1 (`ActionReferralSignup`), keep only Phase 2 (`ActionReferralComplete`).

**Risk:** CONFIRMED BUG - Phase 1 IS hooked in `registration.go:448`. Must remove to prevent double-credit.

---

## Quick Reference - Delete Checklist

```bash
# 1. consts.go - remove ActionReferralSignup constant
# 2. referral.go - delete processReferralSignup() function (~50 lines)
# 3. entrypoints.go - remove ActionReferralSignup from handler map
# 4. registration.go - delete lines 448-455 (the CreateActionLog call)
# 5. Run tests: go test -v ./internal/wingedapp/lib/economy/...
```
