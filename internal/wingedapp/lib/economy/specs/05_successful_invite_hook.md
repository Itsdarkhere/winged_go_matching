# Successful Invite Hook (Phase 1)

> **STATUS: DONE**

## Overview

Award wings to **referrer** when an invited user **enters their invite code** (registered_successfully=true).

This is separate from the "Agent Deployed" hook (Phase 2) which fires later when onboarding completes.

## Two-Phase Referral Rewards

| Phase | Trigger | Hook Location | Who Gets Wings | Status |
|-------|---------|---------------|----------------|--------|
| 1. **Successful Invite** | User enters valid invite code | `EnterInviteCode` | Referrer only (5 wings) | ✅ DONE |
| 2. **Agent Deployed** | User completes onboarding | `DeployUserAgent` | Both (5 wings each) | ✅ DONE |

## Registration Flow

```
[OAuth Login] → [Enter Invite Code] → [Onboarding...] → [Deploy Agent]
      ↓                   ↓                                   ↓
AuthenticateOauthUser  EnterInviteCode                  DeployUserAgent
(creates user)       (registered_successfully=true)    (agent_deployed=true)
                     ✅ ActionReferralSignup           ✅ ActionReferralComplete
                     (referrer gets 5 wings)           (both get 5 wings)
```

## Implementation

### Hook Location

**File:** `business/domain/registration/registration.go` (in `EnterInviteCode`)

```go
// Award referrer when user enters invite code (Phase 1 of referral rewards)
if b.actionLogger != nil {
    if err := b.actionLogger.CreateActionLog(ctx, v.tx, &economy.InsertActionLog{
        UserID: user.ID,
        RefID:  v.userInviteCode.ID,
        Type:   economy.ActionReferralSignup,
    }); err != nil {
        b.logger.Warn(ctx, fmt.Sprintf("award referrer on signup: %v", err))
        // Non-blocking: don't fail registration
    }
}
```

### Lib Handler

**File:** `lib/economy/referral.go:15-103`

```go
// processReferralSignup handles crediting wings to referrer only
// when an invited user enters their invite code - Phase 1.
func (a *ActionLogger) processReferralSignup(ctx context.Context,
    exec boil.ContextExecutor,
    _ *UserTotals, // userTotals for invitee - not used, we credit referrer
    actionInserter *InsertActionLog,
) error
```

Flow:
1. Load invite code by ID (actionInserter.RefID)
2. Check if has referrer (referrer_number_hash)
3. Look up referrer by mobile hash
4. Check idempotency
5. Get/create referrer totals
6. Insert action log
7. Insert transaction
8. Update referrer totals

### Action Types

**File:** `lib/economy/consts.go:63-66`

```go
// Phase 1: When invited user enters invite code (referrer only)
ActionReferralSignup ActionType = "Referral - Friend Signup"
// Phase 2: When invited user completes onboarding/deploys agent (both get wings)
ActionReferralComplete ActionType = "Referral - Friend Complete"
```

### Handler Map

**File:** `lib/economy/entrypoints.go:77-78`

```go
ActionReferralSignup:   a.processReferralSignup,  // Phase 1: referrer only
ActionReferralComplete: a.processReferralBonus,   // Phase 2: both
```

## Edge Cases Handled

1. **Invite code not found** - Returns nil (non-blocking)
2. **Referrer not found** - Returns nil (non-blocking)
3. **No referrer_number_hash** - Returns nil (system-generated codes)
4. **Idempotency** - Uses invite_code_id as RefID, checks existing logs

## Files Changed

| File | Change |
|------|--------|
| `lib/economy/consts.go` | Added `ActionReferralSignup` |
| `lib/economy/referral.go` | Added `processReferralSignup()` handler |
| `lib/economy/entrypoints.go` | Added handler to map |
| `business/domain/registration/registration.go` | Added hook in `EnterInviteCode` |
| `migration/10_referral_action_types.up.sql` | Added new action type to DB constraints |
