# User Invite Completion Hook

> **STATUS: DONE** - Implemented in `business/domain/registration/agent.go`

## Overview

Award wings when an invited user completes registration (deploys their agent).

## Implementation

**Hook Location:** `DeployUserAgent()` in `business/domain/registration/agent.go:42-50`

```go
// Process referral bonus if user was referred (same transaction)
if user.UserInviteCodeRefID.Valid {
    if err := b.actionLogger.CreateActionLog(ctx, tx, &economy.InsertActionLog{
        UserID: user.ID,
        RefID:  user.UserInviteCodeRefID.String,
        Type:   economy.ActionReferralComplete,
    }); err != nil {
        return fmt.Errorf("referral bonus: %w", err)
    }
}
```

## Registration Flow & Hooks

```
[OAuth Login] → [Enter Invite Code] → [Onboarding...] → [Deploy Agent]
      ↓               ↓                                       ↓
 AuthenticateOauthUser  EnterInviteCode                  DeployUserAgent
 (creates user)     (registered_successfully=true)    (agent_deployed=true)
                    (links user_invite_code_ref_id)   ✅ ActionReferralComplete
```

| Step | Function | What Happens | Economy Hook |
|------|----------|--------------|--------------|
| 1 | `AuthenticateOauthUser` | Creates user in DB (if new) | None |
| 2 | `EnterInviteCode` | Sets `registered_successfully=true`, links invite code | None |
| 3 | `DeployUserAgent` | Sets `agent_deployed=true` | ✅ `ActionReferralComplete` (5 wings) |

## Who Gets Wings?

**BOTH get 5 wings each:**
1. **Invitee** - directly credited via `creditReferralBonusWithTotals()`
2. **Referrer** - looked up via `user_invite_code.referrer_number_hash` → `users.sha256_hash`, credited via `creditReferralBonus()`

Full logic in `lib/economy/referral.go:processReferralBonus()` (lines 13-62)

## Related Files

| File | Purpose |
|------|---------|
| `business/domain/registration/agent.go` | Hook implementation |
| `lib/economy/referral.go` | `processReferralBonus()` handler |
| `lib/economy/consts.go:63` | `ActionReferralComplete` constant |