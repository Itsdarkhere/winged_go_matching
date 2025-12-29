# Phase 4: Remove Top-Up Wings & Add Expiry to Subscription Payments

## Status: ✅ COMPLETE

---

## TL;DR

**Top-Up wings are being removed from MVP.** Users earn wings through:
1. Streak milestones (7-day, 30-day)
2. Referrals
3. Attending dates
4. Winged+ subscription (existing)

**Additionally:** Subscription wing payments need `expires_at` set (like earned wings).

---

## What's Changing

### Removing Top-Up

| Remove | Reason |
|--------|--------|
| `ActionWingedMiniPayment` | Not in Dec Final MVP |
| `ActionWingedBoostPayment` | Not in Dec Final MVP |
| `ActionWingedPremiumPayment` | Not in Dec Final MVP |
| `addWingedTopUpMiniPayment()` | Not needed |
| `addWingedTopUpBoostPayment()` | Not needed |
| `addWingedTopUpPremiumPayment()` | Not needed |
| `SubscriptionTypeTopUp` | Not needed |
| `SubscriptionPaymentMini/Boost/Premium` | Not needed |
| DB seed data for Top-Up plans | Not needed |

### Adding Expiry to Subscription Payments

The `addWingedPayment()` function needs to set `ExpiresAt` for Winged+ wings:

```go
// BEFORE
if err = a.transactionStorer.Insert(ctx, exec, &InsertTransaction{
    UserID:       actionInserter.UserID,
    ActionTypeID: string(actionInserter.Type),
    ActionRefID:  actionLog.ID,
    WingsAmount:  subscriptionPlan.Wings,
    Claimed:      true,
    IsCredit:     true,
    ExtraInfo:    actionInserter.JSONDetails,
}); err != nil {

// AFTER
if err = a.transactionStorer.Insert(ctx, exec, &InsertTransaction{
    UserID:       actionInserter.UserID,
    ActionTypeID: string(actionInserter.Type),
    ActionRefID:  actionLog.ID,
    WingsAmount:  subscriptionPlan.Wings,
    Claimed:      true,
    IsCredit:     true,
    ExtraInfo:    actionInserter.JSONDetails,
    ExpiresAt:    null.TimeFrom(time.Now().AddDate(0, 0, EarnedWingsExpiryDays)), // 30 days
}); err != nil {
```

---

## Files to Change

| File | Action | Change |
|------|--------|--------|
| `lib/economy/consts.go` | DELETE | `SubscriptionTypeTopUp`, `SubscriptionPaymentMini/Boost/Premium`, `ActionWingedMini/Boost/PremiumPayment` |
| `lib/economy/winged_payments.go` | DELETE | `addWingedTopUpMiniPayment()`, `addWingedTopUpBoostPayment()`, `addWingedTopUpPremiumPayment()` |
| `lib/economy/winged_payments.go` | MODIFY | Add `ExpiresAt` to `addWingedPayment()` |
| `lib/economy/entrypoints.go` | DELETE | Top-up handlers from `actionLoggerHandlers` map |
| `lib/economy/winged_payments_test.go` | DELETE | Top-up test cases |
| `lib/economy/winged_payments_test.go` | ADD | Test for `ExpiresAt` on subscription payment |
| `migration/04_wings_economy.up.sql` | MODIFY | Remove Top-Up seed data (or leave - won't hurt) |

---

## Code Changes

### 1. consts.go - Remove Top-Up Constants

```go
// DELETE these lines:
SubscriptionTypeTopUp         = "Top Up"
SubscriptionPaymentMini       = "Mini"
SubscriptionPaymentBoost      = "Boost"
SubscriptionPaymentPremium    = "Premium"

ActionWingedMiniPayment    ActionType = "Top Up - Mini"
ActionWingedBoostPayment   ActionType = "Top Up - Boost"
ActionWingedPremiumPayment ActionType = "Top Up - Premium"
```

### 2. winged_payments.go - Remove Top-Up Functions

Delete these functions entirely:
- `addWingedTopUpMiniPayment()`
- `addWingedTopUpBoostPayment()`
- `addWingedTopUpPremiumPayment()`

### 3. winged_payments.go - Add Expiry to addWingedPayment()

```go
func (a *ActionLogger) addWingedPayment(ctx context.Context,
    exec boil.ContextExecutor,
    userTotals *UserTotals,
    actionInserter *InsertActionLog,
    subPayment *SubscriptionPayment,
) error {
    subscriptionPlan, err := a.subscriptionStorer.SubscriptionPlan(ctx, exec,
        subPayment.Type,
        subPayment.Name,
    )
    if err != nil {
        return fmt.Errorf("fetch subscription plan: %w", err)
    }

    actionLog, err := a.actionLogStorer.Insert(ctx, exec, string(actionInserter.Type), actionInserter)
    if err != nil {
        return fmt.Errorf("insert action log: %w", err)
    }

    // Subscription wings expire in 30 days (same as earned wings)
    expiresAt := null.TimeFrom(time.Now().AddDate(0, 0, EarnedWingsExpiryDays))

    if err = a.transactionStorer.Insert(ctx, exec, &InsertTransaction{
        UserID:       actionInserter.UserID,
        ActionTypeID: string(actionInserter.Type),
        ActionRefID:  actionLog.ID,
        WingsAmount:  subscriptionPlan.Wings,
        Claimed:      true,
        IsCredit:     true,
        ExtraInfo:    actionInserter.JSONDetails,
        ExpiresAt:    expiresAt, // <-- ADD THIS
    }); err != nil {
        return fmt.Errorf("insert transaction: %w", err)
    }

    if err = a.userTotalsStorer.Update(ctx, exec, &UpdateUserTotals{
        ID:    userTotals.ID,
        Wings: null.IntFrom(userTotals.Wings + subscriptionPlan.Wings),
    }); err != nil {
        return fmt.Errorf("update user totals: %w", err)
    }

    return nil
}
```

### 4. entrypoints.go - Remove Top-Up from Handler Map

```go
actionLoggerHandlers := map[ActionType]actLoggerHandlerFn{
    ActionWingedXWeeklyPayment:        a.addWingedXWeeklyPayment,
    ActionWingedXMonthlyPayment:       a.addWingedXMonthlyPayment,
    ActionWingedPlusWeeklyPayment:     a.addWingedPlusWeeklyPayment,
    ActionWingedPlusMonthlyPayment:    a.addWingedPlusMonthlyPayment,
    ActionWingedPlusThreeMonthPayment: a.addWingedPlusThreeMonthlyPayment,
    ActionWingedPlusSixMonthPayment:   a.addWingedPlusSixMonthlyPayment,
    // DELETE: ActionWingedMiniPayment, ActionWingedBoostPayment, ActionWingedPremiumPayment
    ActionReferralSignup:              a.processReferralSignup,
    ActionReferralComplete:            a.processReferralBonus,
    ActionAttendDate:                  a.processAttendDate,
    ActionSendMessage:                 a.processSendMessage,
}
```

---

## Test Changes

### Delete Top-Up Tests

Remove all test cases in `winged_payments_test.go` that test Top-Up functionality.

### Add Subscription Expiry Test

```go
{
    name: "success-winged-plus-weekly-sets-expiry",
    setup: func(th *testsuite.Helper) (userID string) {
        user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
        createUserTotals(th, user.Subject.ID, 0)
        return user.Subject.ID
    },
    extraAssertions: func(th *testsuite.Helper, userID string, err error) {
        require.NoError(th.T, err)

        // Verify transaction has expires_at set
        txns := getTestTransactionsByUser(th, userID)
        require.Len(th.T, txns, 1)
        assert.True(th.T, txns[0].ExpiresAt.Valid)

        // Should be ~30 days from now
        expectedExpiry := time.Now().AddDate(0, 0, 30)
        diff := txns[0].ExpiresAt.Time.Sub(expectedExpiry)
        assert.True(th.T, diff < time.Minute && diff > -time.Minute)
    },
},
```

---

## Migration Cleanup (Optional)

The Top-Up seed data in migration 04 can stay - it's just unused rows:

```sql
-- These rows exist but are never used (safe to leave):
INSERT INTO wings_ecn_subscription_plan (subscription_type, name, price, wings)
VALUES ('Top Up', 'Mini', 2.00, 5),
       ('Top Up', 'Boost', 1.00, 30),
       ('Top Up', 'Premium', 0.53, 150);
```

---

## Post-Implementation Checklist

- [x] Remove Top-Up constants from `consts.go`
- [x] Remove Top-Up functions from `winged_payments.go`
- [x] Add `ExpiresAt` to `addWingedPayment()`
- [x] Remove Top-Up handlers from `entrypoints.go`
- [x] Delete Top-Up tests
- [x] Add subscription expiry test
- [x] Run: `go test -v ./internal/wingedapp/lib/economy/...`
- [x] Verify build: `go build ./...`

---

## Summary

| Action | What |
|--------|------|
| REMOVE | Top-Up Mini/Boost/Premium (not in MVP) |
| ADD | `ExpiresAt` to subscription wing payments |
| KEEP | WingedX (free tier, no wings) |
| KEEP | Winged+ (subscription with wings + now expires) |
