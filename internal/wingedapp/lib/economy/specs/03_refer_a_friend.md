#### Requirements:
1. everytime you successfully invite and refer a friend, you and that friend gets 5 wings

#### System design considerations:
1. architectural decisions here should stand as foundation going forward.

#### Politics:
1. CEO vibe coded her own version, need to tell her some of those elements weren't correct
   - for example, actions are triggered as hook logic and not as exposed manual endpoints 
   - this is very important to understand while doing the feature

#### Spec for Claude:
- We have an existing user_invite_system in place where I think the user_table has an invite code, and it can be traced?
- And it can also be traced who Owns that invite code?
- Success criteria: when a new user signs up with an invite code from another person and successfully onboards and hits deploy agent..
   - The invited gets 5 wings, the inviter gets 5 wings

---

## Final Spec (Approved)

### Summary
Hook into `DeployUserAgent()` to credit 5 wings each to inviter and invitee. All DB operations wrapped in single transaction.

---

### 1. New Constants
```go
// lib/economy/consts.go
ActionReferralComplete ActionType = "Referral - Friend Signup"
ReferralBonusWings = 5
```

**No migration needed** - `action_log_type` is varchar, not enum.

---

### 2. DB Flow (Single Transaction)
```
BEGIN TX
  ├─ Check idempotency (action_log exists for user+type+ref?)
  │
  ├─ INVITER (if exists):
  │   ├─ INSERT → wings_ecn_action_log
  │   ├─ INSERT → wings_ecn_transaction (amount: 5, claimed: true)
  │   └─ UPDATE → wings_ecn_user_totals SET total_wings = total_wings + 5
  │
  ├─ INVITEE:
  │   ├─ INSERT → wings_ecn_action_log
  │   ├─ INSERT → wings_ecn_transaction (amount: 5, claimed: true)
  │   └─ UPDATE → wings_ecn_user_totals SET total_wings = total_wings + 5
  │
COMMIT
```

**Key:** Atomic update `total_wings = total_wings + 5` prevents race conditions.

---

### 3. API Surface

**Public API** (existing):
```go
// lib/economy/entrypoints.go
func (a *ActionLogger) CreateActionLog(
    ctx context.Context,
    exec boil.ContextExecutor,
    inserter *InsertActionLog,
) error
```

**InsertActionLog struct** (existing):
```go
type InsertActionLog struct {
    UserID      string     `json:"user_id" validate:"required"`      // invitee
    RefID       string     `json:"ref_id" validate:"required"`       // invite_code.id
    Type        ActionType `json:"category" validate:"required"`     // ActionReferralComplete
    JSONDetails null.JSON  `json:"json_details"`                     // custom payload
}
```

**Handler signature** (existing pattern):
```go
type actLoggerHandlerFn func(
    ctx context.Context,
    exec boil.ContextExecutor,
    userTotals *UserTotals,      // invitee's totals (pre-fetched)
    inserter *InsertActionLog,
) error
```

**Routing** (in entrypoints.go - add to existing map):
```go
actionLoggerHandlers := map[ActionType]actLoggerHandlerFn{
    // ... existing handlers
    ActionReferralComplete: a.processReferralBonus,  // add this
}
```

**Internal handler** (new file):
```go
// lib/economy/referral.go
func (a *ActionLogger) processReferralBonus(
    ctx context.Context,
    exec boil.ContextExecutor,
    userTotals *UserTotals,      // invitee's totals
    inserter *InsertActionLog,
) error
```

**Logic in processReferralBonus:**
1. Check idempotency: query action_log for `user_ref_id + action_log_type + ext_domain_ref_id`
2. Load `user_invite_code` by `inserter.RefID` → get `referrer_number_hash`
3. Look up referrer user by phone hash
4. If referrer exists:
   - Fetch referrer's `userTotals`
   - Credit 5 wings to referrer (action_log + transaction + atomic update totals)
5. Credit 5 wings to invitee using passed `userTotals` (action_log + transaction + atomic update)

---

### 4. Hook Location
```go
// business/domain/registration/agent.go - DeployUserAgent()

// Inside existing transaction, after setting agents_deployed = true:
if user.UserInviteCodeRefID.Valid {
    err := a.actionLogger.CreateActionLog(ctx, tx, &economy.InsertActionLog{
        UserID: user.ID,
        RefID:  user.UserInviteCodeRefID.String,
        Type:   economy.ActionReferralComplete,
        // JSONDetails: optional - handler looks up referrer from RefID
    })
    if err != nil {
        a.logger.Error("referral bonus failed", "error", err, "user_id", user.ID)
        // Don't fail deployment - log and continue
    }
}
```

**Why here:** Deploy agent = final onboarding step. Same TX = rollback if deploy fails.

---

### 5. DI Changes
- Inject `ActionLogger` into registration domain
- Add user store method: `UserByPhoneHash(ctx, exec, hash) (*User, error)`

---

### 6. Edge Cases

| Case | Handling |
|------|----------|
| No invite code | Skip referral logic |
| Referrer not found (deleted) | Credit invitee only, log warning |
| Already processed (idempotency) | Return early, no error |
| Invitee has no user_totals row | Create row with 5 wings |
| Referrer has existing balance | Atomic add: `total_wings + 5` |

---

### 7. Files to Change

| File | Change |
|------|--------|
| `lib/economy/consts.go` | Add `ActionReferralComplete` + `ReferralBonusWings = 5` |
| `lib/economy/referral.go` | New - `processReferralBonus()` handler (unexported) |
| `lib/economy/entrypoints.go` | Add to `actionLoggerHandlers` map |
| `lib/economy/store/user_action_log.go` | Idempotency check query |
| `lib/economy/store/user_totals.go` | Atomic increment: `SET total_wings = total_wings + $1` |
| `business/domain/registration/agent.go` | Hook call via `CreateActionLog()` |
| `di/deps_biz.go` | Inject ActionLogger into registration |

---

### 8. Tests

**Unit tests** (`lib/economy/referral_test.go`):
| Test Case | Setup | Assert |
|-----------|-------|--------|
| `success-both-credited` | invitee with invite code, referrer exists | Both get 5 wings |
| `success-adds-to-existing-balance` | referrer has 10 wings | referrer: 15, invitee: 5 |
| `success-invitee-only-referrer-deleted` | referrer not found | invitee gets 5, no error |
| `success-idempotent-no-double-credit` | call twice | still 5 wings each |
| `error-no-invite-code` | invitee without code | skip, no wings |

**Integration test** (`api/*_test.go`):
- Full flow: signup with invite code → deploy agent → verify DB state

---

### 9. Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| ActionType storage | VARCHAR | No migration for new types, indexes perform same at scale |
| Transaction scope | Same TX as deploy | Atomic - rollback if deploy fails |
| Idempotency | Check action_log | Prevent double-credit on retry |
| Referrer lookup | By phone hash | Existing pattern in invite system |
| Failed referral | Log, don't fail deploy | User experience > bonus accuracy | 