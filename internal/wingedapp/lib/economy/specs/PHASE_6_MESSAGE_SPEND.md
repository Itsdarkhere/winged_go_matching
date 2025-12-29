# Phase 6: Message Spend Reconciliation

## TL;DR
**Current implementation is CORRECT.** Message spend properly implements the "5 messages = 1 wing" spec with pre-flight check, threshold-based deduction, premium bypass, and idempotency. Excellent test coverage.

---

## Spec vs Current Implementation

| Aspect | MVP Spec | Current Implementation | Status |
|--------|----------|------------------------|--------|
| **Threshold** | 5 messages = 1 wing | `SendMessageThreshold = 5` | ✅ CORRECT |
| **Wing Cost** | 1 wing per threshold | `SendMessageWingsCost = 1` | ✅ CORRECT |
| **Pre-flight Check** | Check before action | `CanPerformAction()` | ✅ CORRECT |
| **Balance Guard** | Require wings >= 1 | `wings < SendMessageWingsCost → error` | ✅ CORRECT |
| **Counter** | Track sent messages | `counter_sent_messages` in user_totals | ✅ CORRECT |
| **Deduction Timing** | Only at threshold | `newSentMessages % 5 == 0` | ✅ CORRECT |
| **Transaction** | Debit entry at threshold | `IsCredit: false` | ✅ CORRECT |
| **Premium Bypass** | Skip for premium users | `isPremiumActive()` check | ✅ CORRECT |
| **Idempotency** | No duplicate counts | Checks `(user_id, category, ref_id)` | ✅ CORRECT |
| **Hook Location** | Before AI prompt | `youragent/prompt.go` | ✅ CORRECT |

---

## Current Code Analysis

### Pre-flight Check (`youragent/prompt.go:12-23`)
```go
// check economy balance BEFORE processing
canPerform, err := b.actionLogger.CanPerformAction(ctx, b.trans.DB(), &economy.CanPerformActionParams{
    UserID:     userID,
    ActionType: economy.ActionSendMessage,
})
if !canPerform {
    return nil, economy.ErrInsufficientWings  // ✅ Blocked early
}
```

### Post-action Logging (`youragent/prompt.go:55-63`)
```go
// log action for economy (deducts wings at threshold)
err = b.actionLogger.CreateActionLog(ctx, b.trans.DB(), &economy.InsertActionLog{
    UserID: userID,
    RefID:  promptResp.ID,  // ✅ Message ID for idempotency
    Type:   economy.ActionSendMessage,
})
```

### Handler (`send_message.go`)
```go
func (a *ActionLogger) processSendMessage(...) error {
    // 1. ✅ Premium bypass
    if isPremiumActive(userTotals) {
        return nil  // No records, no deduction
    }

    // 2. ✅ Idempotency check
    existingLogs, _ := a.actionLogStorer.ActionLogs(ctx, exec, filter)
    if len(existingLogs) > 0 { return nil }

    // 3. ✅ Balance guard
    if userTotals.Wings < SendMessageWingsCost {
        return ErrInsufficientWings
    }

    // 4. ✅ Insert action log
    actionLog, _ := a.actionLogStorer.Insert(...)

    // 5. ✅ Increment counter
    newSentMessages := userTotals.SentMessages + 1

    // 6. ✅ Conditional deduction at threshold
    shouldDeductWing := newSentMessages % SendMessageThreshold == 0
    if shouldDeductWing {
        // Create debit transaction
        a.transactionStorer.Insert(ctx, exec, &InsertTransaction{
            WingsAmount: SendMessageWingsCost,  // 1
            IsCredit:    false,                  // debit
        })
    }

    // 7. ✅ Update totals (counter always, wings conditionally)
    a.userTotalsStorer.Update(ctx, exec, &UpdateUserTotals{
        SentMessages: newSentMessages,
        Wings:        conditionally updated,
    })
}
```

### Guard (`entrypoints.go:129-136`)
```go
func (a *ActionLogger) canPerformWithBalance(actionType ActionType, wings int) (bool, error) {
    switch actionType {
    case ActionSendMessage:
        if wings < SendMessageWingsCost {
            return false, ErrInsufficientWings  // ✅ Blocked if < 1 wing
        }
        return true, nil
    default:
        return true, nil  // Earning actions always allowed
    }
}
```

---

## Test Coverage (Excellent)

### `TestEconomy_SendMessage`
| Test Case | Covered |
|-----------|---------|
| `success-message-counted-no-deduction-yet` | ✅ Counter increments, no wing deduction |
| `success-5th-message-deducts-1-wing` | ✅ Threshold triggers deduction |
| `success-idempotency-same-message-not-counted-twice` | ✅ Duplicate RefID ignored |
| `error-insufficient-wings` | ✅ Returns `ErrInsufficientWings` |
| `success-10th-message-deducts-2nd-wing` | ✅ Multiple threshold cycles |

### `TestEconomy_CanPerformAction`
| Test Case | Covered |
|-----------|---------|
| `success-can-send-message-with-wings` | ✅ |
| `error-cannot-send-message-without-wings` | ✅ |
| `success-earning-actions-always-allowed` | ✅ |
| `success-premium-subscriber-bypasses-wings-check` | ✅ |
| `error-expired-premium-does-not-bypass` | ✅ |

---

## Flow Diagram

```
User sends message
       │
       ▼
┌─────────────────────┐
│ CanPerformAction()  │
│ (pre-flight check)  │
└──────────┬──────────┘
           │
     ┌─────┴─────┐
     │ wings >= 1?│
     └─────┬─────┘
           │
    No     │     Yes
    ▼      │      ▼
┌─────────┐│┌─────────────────┐
│ BLOCKED ││ Process message │
│ Error   │││ (AI prompt)     │
└─────────┘│└────────┬────────┘
           │         │
           │         ▼
           │  ┌──────────────────┐
           │  │ CreateActionLog() │
           │  │ (post-action)    │
           │  └────────┬─────────┘
           │           │
           │     ┌─────┴─────┐
           │     │ Premium?  │
           │     └─────┬─────┘
           │           │
           │    Yes    │    No
           │     ▼     │     ▼
           │ ┌───────┐ │ ┌───────────────┐
           │ │ SKIP  │ │ │ Idempotency?  │
           │ │ (free)│ │ └───────┬───────┘
           │ └───────┘ │         │
           │           │  Already│  New
           │           │    ▼    │   ▼
           │           │ ┌─────┐ │ ┌─────────────┐
           │           │ │SKIP │ │ │Increment    │
           │           │ └─────┘ │ │sent_messages│
           │           │         │ └──────┬──────┘
           │           │         │        │
           │           │         │  ┌─────┴─────┐
           │           │         │  │ % 5 == 0? │
           │           │         │  └─────┬─────┘
           │           │         │        │
           │           │         │   No   │  Yes
           │           │         │   ▼    │   ▼
           │           │         │┌─────┐ │ ┌────────────┐
           │           │         ││Done │ │ │Deduct wing │
           │           │         │└─────┘ │ │-1 from bal │
           │           │         │        │ └────────────┘
```

---

## Files - No Changes Required

| File | Status |
|------|--------|
| `lib/economy/consts.go` | ✅ `SendMessageThreshold = 5`, `SendMessageWingsCost = 1` |
| `lib/economy/send_message.go` | ✅ Full implementation |
| `lib/economy/entrypoints.go` | ✅ Handler registered, `canPerformWithBalance()` |
| `business/domain/youragent/prompt.go` | ✅ Pre-flight + post-action hooks |
| `lib/economy/send_message_test.go` | ✅ 10 test cases |

---

## Edge Cases (All Handled)

| Case | Handling | Status |
|------|----------|--------|
| User has 0 wings | `CanPerformAction` returns error | ✅ |
| User has exactly 1 wing | Allowed until threshold hit | ✅ |
| Premium subscriber with 0 wings | Bypasses check entirely | ✅ |
| Expired premium with 0 wings | NOT bypassed, returns error | ✅ |
| Same message ID sent twice | Idempotency check, no double count | ✅ |
| 5th message when user has 1 wing | Deducts to 0, next blocked | ✅ |
| User totals don't exist | Creates totals | ✅ |

---

## Architecture Notes

### Two-Phase Economy Check
1. **Pre-flight** (`CanPerformAction`) - Guards the action before expensive work
2. **Post-action** (`CreateActionLog`) - Records and deducts after success

This pattern ensures:
- User isn't charged if AI prompt fails
- User can't exploit by retrying failed prompts
- Clean separation of concerns

### Premium Bypass
- Stored in `wings_ecn_user_totals.premium_expires_in`
- Checked both in `CanPerformAction` and `processSendMessage`
- Premium users: no records created, no deduction

---

## Summary

**Status: COMPLETE** ✅

Message Spend implementation is fully aligned with MVP spec:
- 5 messages = 1 wing threshold working
- Pre-flight balance check blocks insufficient wings
- Counter tracks all messages
- Transaction only at threshold (not every message)
- Premium subscribers bypass entirely
- Idempotent per message_id
- Excellent test coverage (10 cases)

**No changes required.**

---

## Quick Reference

```go
// Constants
SendMessageThreshold = 5   // messages per wing
SendMessageWingsCost = 1   // wing per threshold

// Pre-flight (in business layer)
canPerform, _ := e.CanPerformAction(ctx, db, &CanPerformActionParams{
    UserID:     userID,
    ActionType: ActionSendMessage,
})

// Post-action (after successful processing)
e.CreateActionLog(ctx, db, &InsertActionLog{
    UserID: userID,
    RefID:  messageID,  // for idempotency
    Type:   ActionSendMessage,
})
```
