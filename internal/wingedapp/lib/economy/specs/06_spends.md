#### Tech specs

#### Spends:
- send 5 messages
  - you see your chat? if you send like 5 messages every x seconds, youre substracted 1 wing..
  - but we also need to extend can perform action right? I think you need to follow the api entrypoint pattern so far
- guards:
  - I think we need to extnd can perform action (rememer this is the pattern entrypoint like CreatActionLog)
  - ofc.. return proper error taxonomy..
  - see previous PRs to get a better sense of the hooks patterns, and idempotetncy.. and checkout the PR descriptions

---

## Execution Summary

### Completed ✅

| Task | File(s) |
|------|---------|
| Add `ActionSendMessage` const | `consts.go` |
| Add `SendMessageThreshold` (5) and `SendMessageWingsCost` (1) | `consts.go` |
| Extend `UpdateUserTotals` with `SentMessages` field | `models.go`, `store/user_totals.go` |
| Update repo layer to support `SentMessages` update | `db/repo/wings_ecn_user_totals.go` |
| Implement `CanPerformAction` with balance check | `entrypoints.go` |
| Create `processSendMessage` handler | `send_message.go` |
| Register handler in action logger map | `entrypoints.go` |
| Write comprehensive tests | `send_message_test.go` |
| Remove debug `fmt.Println` from codebase | `entrypoints.go` |

### Logic Flow

1. **CanPerformAction** - Guard check before action
   - `ActionSendMessage` requires `wings >= 1`
   - Earning actions (payments, referrals) always allowed

2. **processSendMessage** - Handler for message sends
   - Idempotency: checks existing action log by `(user_id, category, ref_id)`
   - Increments `sent_messages` counter on every call
   - Deducts 1 wing when `sent_messages % 5 == 0`
   - Creates debit transaction only at threshold

### Test Coverage

- `success-message-counted-no-deduction-yet` - counter increments, no wing deduction
- `success-5th-message-deducts-1-wing` - threshold triggers deduction
- `success-idempotency-same-message-not-counted-twice` - duplicate message IDs ignored
- `error-insufficient-wings` - returns `ErrInsufficientWings`
- `success-10th-message-deducts-2nd-wing` - multiple threshold cycles work
- `CanPerformAction` tests for balance checks
