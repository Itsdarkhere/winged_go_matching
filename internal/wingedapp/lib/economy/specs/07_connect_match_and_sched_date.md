#### Connect with match, and schedule a date

- we need to continue hooking into the action logs for when a user connects with a match, and when they schedule a date
- see previous prs.. and lets see what you'll do
- Additional context:
  - this is a dating app, people can date.. somewheere I think when you want to propose.. you need to spend wings, so hook into that.. i think you need to wrap it with cAn, and spend.. (i think this is when you propose)
  - then schedule a date... I think this is when you actually start to propos..
  - but I am a little confused. this app is vibe code.. not sure where those 2 actions are probably needed to be hooked, any ideas?\

---

## Claude Analysis

### TL;DR

Two economy hooks needed:
1. **Connect Proposal** - When a user proposes on a match (spend wings)
2. **Schedule Date Confirmed** - When mutual proposal auto-creates date instance (spend/earn wings)

Both hooks follow the existing `actionLogger.CreateActionLog()` pattern.

---

### 1. Current Flow Discovery

#### Match Proposal Flow (User "Connects")

```
POST /users/matching/matches/{id}/propose
  │
  ├── api/api_matching.go:313        → userMatchPropose()
  ├── business/matching/matching.go:285 → ProposeOnMatch()
  └── lib/matching/logic.go:320      → ProposeMatch()
        │
        ├── Update user's action → "Proposed"
        │
        └── If partner ALSO "Proposed" (mutual):
            └── createDateInstanceOnMutualProposal() (line 365)
                ├── Insert date_instance (status="Proposed")
                ├── Insert date_instance_log
                └── Update match_result.lifecycle_status="Scheduling"
```

**Key Insight**: "Connect with a match" = `ProposeMatch()`. Mutual detection happens automatically.

#### Date Scheduling Flow

Date instances are **auto-created** when both users propose. No separate "schedule" API.

```
Mutual Proposal Detected
  │
  └── createDateInstanceOnMutualProposal() (lib/matching/logic.go:365)
        │
        ├── Get match config (for expiration window)
        ├── Create date_instance with status="Proposed"
        └── Both users can now propose time slots
```

**Time Proposals** (separate flow - not for this spec):
- `lib/scheduling/store/proposal.go:76` → `InsertProposal()`
- `lib/scheduling/store/proposal.go:140` → `AcceptProposal()`

---

### 2. Existing Economy Pattern

#### Handler Registration (`lib/economy/action_logger.go:66-80`)

```go
var handlers = map[ActionType]func(...) error {
    ActionDailyCheckIn:       (*ActionLogger).processDailyCheckin,
    ActionReferralSignup:     (*ActionLogger).processReferralSignup,
    ActionReferralComplete:   (*ActionLogger).processReferralComplete,
    ActionAttendDate:         (*ActionLogger).processAttendDate,
    ActionSendMessage:        (*ActionLogger).processSendMessage,
    // NEW: Add our handlers here
}
```

#### Spending Example (`send_message.go`)

```go
func (a *ActionLogger) processSendMessage(ctx, exec, log) error {
    // 1. Check idempotency
    existing, _ := a.actionLogStore.ActionLog(ctx, exec, filter)
    if existing != nil { return nil } // already processed

    // 2. Check/create user totals
    totals := a.ensureTotalsExist(ctx, exec, log.UserID)

    // 3. Check balance (every 5 messages costs 1 wing)
    if shouldCharge && totals.Wings < SendMessageWingsCost {
        return ErrInsufficientWings
    }

    // 4. Deduct wings
    a.transactionStore.Insert(ctx, exec, &InsertTransaction{
        UserID:   log.UserID,
        Amount:   SendMessageWingsCost,
        IsCredit: false, // spending
    })

    // 5. Update user totals
    a.userTotalsStore.Update(ctx, exec, totals.ID, &UpdateUserTotals{
        Wings: null.From(totals.Wings - SendMessageWingsCost),
    })

    return nil
}
```

---

### 3. Implementation Plan

#### Phase 1: Add Constants

**File**: `lib/economy/consts.go`

```go
// New action types
ActionConnectProposal       ActionType = "Connect - Proposal"
ActionScheduleDateConfirmed ActionType = "Schedule - Date Confirmed"

// New costs (configurable - discuss with product)
ConnectProposalWingsCost   = 1  // Cost to propose on a match
ScheduleDateConfirmedWings = 0  // Reward when date instance created (or cost?)
```

**Questions for Product**:
- Is proposing free? Or costs wings?
- When mutual proposal creates date instance, who pays? Both? Neither?
- Is there a reward for mutual match?

#### Phase 2: Add Handler - Connect Proposal

**File**: `lib/economy/connect_proposal.go` (new)

```go
func (a *ActionLogger) processConnectProposal(ctx, exec, log) error {
    // Idempotency check by (UserID, RefID=match_result_id)
    existing, _ := a.actionLogStore.ActionLog(ctx, exec, &ActionLogFilter{
        UserID: null.From(log.UserID),
        RefID:  null.From(log.RefID),
        Type:   null.From(string(ActionConnectProposal)),
    })
    if existing != nil {
        return nil // Already processed
    }

    // Ensure totals exist
    totals, err := a.ensureTotalsExist(ctx, exec, log.UserID)
    if err != nil {
        return err
    }

    // Check balance (if proposal costs wings)
    if ConnectProposalWingsCost > 0 {
        if totals.Wings < ConnectProposalWingsCost {
            return ErrInsufficientWings
        }

        // Deduct wings
        _, err = a.transactionStore.Insert(ctx, exec, &InsertTransaction{
            UserID:   log.UserID,
            Amount:   ConnectProposalWingsCost,
            IsCredit: false,
        })
        if err != nil {
            return fmt.Errorf("insert transaction: %w", err)
        }

        // Update totals
        _, err = a.userTotalsStore.Update(ctx, exec, totals.ID, &UpdateUserTotals{
            Wings: null.From(totals.Wings - ConnectProposalWingsCost),
        })
        if err != nil {
            return fmt.Errorf("update totals: %w", err)
        }
    }

    // Insert action log
    _, err = a.actionLogStore.Insert(ctx, exec, log)
    return err
}
```

#### Phase 3: Add Handler - Schedule Date Confirmed

**File**: `lib/economy/schedule_date_confirmed.go` (new)

```go
func (a *ActionLogger) processScheduleDateConfirmed(ctx, exec, log) error {
    // Idempotency check by (UserID, RefID=date_instance_id)
    existing, _ := a.actionLogStore.ActionLog(ctx, exec, &ActionLogFilter{
        UserID: null.From(log.UserID),
        RefID:  null.From(log.RefID),
        Type:   null.From(string(ActionScheduleDateConfirmed)),
    })
    if existing != nil {
        return nil
    }

    // Ensure totals
    totals, err := a.ensureTotalsExist(ctx, exec, log.UserID)
    if err != nil {
        return err
    }

    // Credit or debit based on config
    if ScheduleDateConfirmedWings != 0 {
        isCredit := ScheduleDateConfirmedWings > 0
        amount := abs(ScheduleDateConfirmedWings)

        if !isCredit && totals.Wings < amount {
            return ErrInsufficientWings
        }

        _, err = a.transactionStore.Insert(ctx, exec, &InsertTransaction{
            UserID:   log.UserID,
            Amount:   amount,
            IsCredit: isCredit,
        })
        if err != nil {
            return err
        }

        newWings := totals.Wings
        if isCredit {
            newWings += amount
        } else {
            newWings -= amount
        }

        _, err = a.userTotalsStore.Update(ctx, exec, totals.ID, &UpdateUserTotals{
            Wings: null.From(newWings),
        })
        if err != nil {
            return err
        }
    }

    _, err = a.actionLogStore.Insert(ctx, exec, log)
    return err
}
```

#### Phase 4: Register Handlers

**File**: `lib/economy/action_logger.go`

```go
var handlers = map[ActionType]func(...) error {
    // ... existing
    ActionConnectProposal:       (*ActionLogger).processConnectProposal,
    ActionScheduleDateConfirmed: (*ActionLogger).processScheduleDateConfirmed,
}
```

#### Phase 5: Hook into Matching Logic

**File**: `lib/matching/logic.go`

**Option A**: Inject actionLogger into matching.Logic (cleaner)
**Option B**: Call from business layer after ProposeMatch returns (current pattern)

**Recommended**: Option B - matches existing architecture.

**File**: `business/domain/matching/matching.go`

```go
func (b *Business) ProposeOnMatch(ctx, userID, matchID string) (*ProposeResult, error) {
    var result *ProposeResult
    err := b.trans.Transaction(ctx, func(tx *sql.Tx) error {
        var err error
        result, err = b.userMatchActioner.ProposeMatch(ctx, tx, params)
        if err != nil {
            return err
        }

        // NEW: Log economy action for proposal
        if err := b.actionLogger.CreateActionLog(ctx, tx, &economy.InsertActionLog{
            UserID: userID,
            RefID:  matchID,
            Type:   economy.ActionConnectProposal,
        }); err != nil {
            // Log but continue - economy tracking is non-critical
            b.logger.Warn("failed to log connect proposal", "error", err)
        }

        // NEW: If mutual proposal, log for both users
        if result.MutualProposal && result.DateInstanceID != "" {
            // Log for proposing user
            _ = b.actionLogger.CreateActionLog(ctx, tx, &economy.InsertActionLog{
                UserID: userID,
                RefID:  result.DateInstanceID,
                Type:   economy.ActionScheduleDateConfirmed,
            })
            // Log for partner (need to get partner ID from match)
            _ = b.actionLogger.CreateActionLog(ctx, tx, &economy.InsertActionLog{
                UserID: partnerID, // Need to fetch this
                RefID:  result.DateInstanceID,
                Type:   economy.ActionScheduleDateConfirmed,
            })
        }

        return nil
    })
    return result, err
}
```

**DI Update Required**: Add actionLogger to matching.Business

---

### 4. Files to Modify

| File | Change |
|------|--------|
| `lib/economy/consts.go` | Add ActionConnectProposal, ActionScheduleDateConfirmed |
| `lib/economy/connect_proposal.go` | NEW - handler logic |
| `lib/economy/schedule_date_confirmed.go` | NEW - handler logic |
| `lib/economy/action_logger.go` | Register new handlers |
| `business/domain/matching/matching.go` | Add actionLogger calls |
| `business/domain/matching/business.go` | Add actionLogger dependency |
| `di/deps_biz.go` | Inject actionLogger into bizMatching |

---

### 5. Alternative: "Can + Spend" Pattern

The user mentioned "wrap it with Can, and Spend". This would be:

```go
// In business layer, BEFORE calling ProposeMatch:
canPropose, err := b.economyChecker.CanPerformAction(ctx, userID, economy.ActionConnectProposal)
if !canPropose {
    return nil, ErrInsufficientWings
}

// After ProposeMatch succeeds:
err = b.economySpender.SpendForAction(ctx, tx, userID, economy.ActionConnectProposal, matchID)
```

This requires:
1. New interface `economyChecker` with `CanPerformAction()`
2. New interface `economySpender` with `SpendForAction()`
3. Pre-flight balance check before action

**Current pattern** (CreateActionLog) handles check+spend atomically inside the handler. More flexible but less explicit.

---

### 6. Test Coverage Needed

```go
// connect_proposal_test.go
func TestActionLogger_ProcessConnectProposal(t *testing.T) {
    cases := []testCaseProcessConnectProposal{
        {name: "success-first-proposal"},
        {name: "success-idempotent-duplicate"},
        {name: "error-insufficient-wings"},
        {name: "success-free-proposal-zero-cost"},
    }
}

// schedule_date_confirmed_test.go
func TestActionLogger_ProcessScheduleDateConfirmed(t *testing.T) {
    cases := []testCaseProcessScheduleDateConfirmed{
        {name: "success-credits-wings"},
        {name: "success-idempotent"},
        {name: "success-both-users-logged"},
    }
}

// Integration test in matching
func TestProposeOnMatch_LogsEconomyAction(t *testing.T) {
    // Verify action_log created after proposal
}

func TestProposeOnMatch_MutualProposal_LogsBothUsers(t *testing.T) {
    // Verify both users get action_log on mutual
}
```

---

### 7. Open Questions for Product

1. **Proposal Cost**: Should proposing cost wings? How many?
2. **Mutual Reward**: When mutual proposal creates date, is there a reward?
3. **Who Pays**: If date creation costs wings, does the proposing user pay, or both?
4. **Failed Proposal**: If other user already passed, does proposer still pay?
5. **Subscription Override**: Do subscribers get free proposals?

---

### 8. Execution Checklist

- [ ] Add constants to `consts.go`
- [ ] Create `connect_proposal.go` handler
- [ ] Create `schedule_date_confirmed.go` handler
- [ ] Register handlers in `action_logger.go`
- [ ] Add actionLogger to matching.Business
- [ ] Update DI in `deps_biz.go`
- [ ] Hook CreateActionLog in ProposeOnMatch
- [ ] Write unit tests for handlers
- [ ] Write integration test for ProposeOnMatch
- [ ] Run `make wingedapp-di-generate`
- [ ] Run full test suite
- [ ] Manual QA: propose → check action_log table
