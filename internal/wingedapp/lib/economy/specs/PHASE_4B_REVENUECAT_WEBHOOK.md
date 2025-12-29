# Phase 4B: RevenueCat Webhook Integration

## Status: PENDING

---

## TL;DR

Expose a webhook endpoint to receive RevenueCat subscription events. On `INITIAL_PURCHASE` or `RENEWAL`, credit wings to user with 30-day expiry. Idempotent via event ID.

---

## RevenueCat Webhook Payload

```json
{
  "api_version": "1.0",
  "event": {
    "id": "unique-event-id",
    "type": "INITIAL_PURCHASE",
    "event_timestamp_ms": 1234567890000,
    "app_user_id": "user_123",
    "original_app_user_id": "user_123",
    "aliases": ["alias1", "alias2"],
    "product_id": "com.app.wingedplus_weekly",
    "period_type": "NORMAL",
    "purchased_at_ms": 1234567890000,
    "expiration_at_ms": 1234567890000,
    "environment": "PRODUCTION",
    "store": "APP_STORE",
    "country_code": "US",
    "currency": "USD",
    "price": 14.95,
    "entitlement_ids": ["wingedplus"],
    "offer_code": null,
    "is_family_share": false
  }
}
```

---

## Event Types We Handle

| Event Type | Action |
|------------|--------|
| `INITIAL_PURCHASE` | Credit wings based on product_id |
| `RENEWAL` | Credit wings based on product_id |
| `CANCELLATION` | Log only (wings already credited) |
| `EXPIRATION` | Log only (expiry cron handles balance) |
| `BILLING_ISSUE` | Log only |
| Others | Ignore |

---

## Product ID Mapping

| RevenueCat product_id | Our ActionType | Wings |
|-----------------------|----------------|-------|
| `com.app.wingedplus_weekly` | `ActionWingedPlusWeeklyPayment` | 25 |
| `com.app.wingedplus_monthly` | `ActionWingedPlusMonthlyPayment` | 55 |
| `com.app.wingedplus_3month` | `ActionWingedPlusThreeMonthPayment` | 180 |
| `com.app.wingedplus_6month` | `ActionWingedPlusSixMonthPayment` | 360 |
| `com.app.wingedx_*` | `ActionWingedX*` | 0 (no wings) |

---

## API Endpoint

### POST `/webhooks/revenuecat`

**Auth:** Webhook secret header validation (not user auth)

**Request:**
```json
{
  "api_version": "1.0",
  "event": { ... }
}
```

**Response (200):**
```json
{
  "success": true,
  "event_id": "unique-event-id",
  "action": "wings_credited",
  "wings_awarded": 25
}
```

**Response (200 - already processed):**
```json
{
  "success": true,
  "event_id": "unique-event-id",
  "action": "already_processed"
}
```

**Response (200 - ignored event):**
```json
{
  "success": true,
  "event_id": "unique-event-id",
  "action": "ignored",
  "reason": "event_type_not_handled"
}
```

**Response (400):**
```json
{
  "success": false,
  "error": "invalid_payload"
}
```

**Response (401):**
```json
{
  "success": false,
  "error": "invalid_webhook_secret"
}
```

---

## Implementation

### 1. Request/Response Models

```go
// RevenueCatWebhookRequest is the webhook payload from RevenueCat.
type RevenueCatWebhookRequest struct {
    APIVersion string              `json:"api_version"`
    Event      RevenueCatEvent     `json:"event"`
}

type RevenueCatEvent struct {
    ID                string   `json:"id"`
    Type              string   `json:"type"`
    EventTimestampMs  int64    `json:"event_timestamp_ms"`
    AppUserID         string   `json:"app_user_id"`
    OriginalAppUserID string   `json:"original_app_user_id"`
    Aliases           []string `json:"aliases"`
    ProductID         string   `json:"product_id"`
    PeriodType        string   `json:"period_type"`
    PurchasedAtMs     int64    `json:"purchased_at_ms"`
    ExpirationAtMs    int64    `json:"expiration_at_ms"`
    Environment       string   `json:"environment"`
    Store             string   `json:"store"`
    CountryCode       string   `json:"country_code"`
    Currency          string   `json:"currency"`
    Price             float64  `json:"price"`
    EntitlementIDs    []string `json:"entitlement_ids"`
    OfferCode         *string  `json:"offer_code"`
    IsFamilyShare     bool     `json:"is_family_share"`
}

type RevenueCatWebhookResponse struct {
    Success      bool   `json:"success"`
    EventID      string `json:"event_id"`
    Action       string `json:"action"`
    WingsAwarded int    `json:"wings_awarded,omitempty"`
    Reason       string `json:"reason,omitempty"`
    Error        string `json:"error,omitempty"`
}
```

### 2. Product ID to Action Type Mapping

```go
// productIDToActionType maps RevenueCat product IDs to our action types.
var productIDToActionType = map[string]economy.ActionType{
    "com.app.wingedplus_weekly":  economy.ActionWingedPlusWeeklyPayment,
    "com.app.wingedplus_monthly": economy.ActionWingedPlusMonthlyPayment,
    "com.app.wingedplus_3month":  economy.ActionWingedPlusThreeMonthPayment,
    "com.app.wingedplus_6month":  economy.ActionWingedPlusSixMonthPayment,
    // WingedX products - no wings, but we still track
    "com.app.wingedx_weekly":  economy.ActionWingedXWeeklyPayment,
    "com.app.wingedx_monthly": economy.ActionWingedXMonthlyPayment,
}
```

### 3. Handler Flow

```go
func (a *Mux) webhookRevenueCat(ctx fiber.Ctx) error {
    // 1. Validate webhook secret header
    secret := ctx.Get("X-RevenueCat-Webhook-Secret")
    if secret != a.cfg.RevenueCatWebhookSecret {
        return ctx.Status(401).JSON(RevenueCatWebhookResponse{
            Success: false,
            Error:   "invalid_webhook_secret",
        })
    }

    // 2. Parse payload
    var req RevenueCatWebhookRequest
    if err := ctx.Bind().JSON(&req); err != nil {
        return ctx.Status(400).JSON(RevenueCatWebhookResponse{
            Success: false,
            Error:   "invalid_payload",
        })
    }

    // 3. Process event
    result, err := a.bizEconomy.ProcessRevenueCatEvent(ctx, &req)
    if err != nil {
        return a.respond(ctx, nil, err)
    }

    return ctx.JSON(result)
}
```

### 4. Business Layer

```go
func (b *Business) ProcessRevenueCatEvent(ctx context.Context, req *RevenueCatWebhookRequest) (*RevenueCatWebhookResponse, error) {
    event := req.Event

    // Only handle purchase/renewal events
    if event.Type != "INITIAL_PURCHASE" && event.Type != "RENEWAL" {
        return &RevenueCatWebhookResponse{
            Success: true,
            EventID: event.ID,
            Action:  "ignored",
            Reason:  "event_type_not_handled",
        }, nil
    }

    // Map product_id to action type
    actionType, ok := productIDToActionType[event.ProductID]
    if !ok {
        return &RevenueCatWebhookResponse{
            Success: true,
            EventID: event.ID,
            Action:  "ignored",
            Reason:  "unknown_product_id",
        }, nil
    }

    // Process payment (idempotent via event.ID as RefID)
    tx, err := b.transactor.TX()
    if err != nil {
        return nil, fmt.Errorf("begin tx: %w", err)
    }
    defer b.transactor.Rollback(tx)

    result, err := b.subscriptionProcessor.ProcessPayment(ctx, tx, &ProcessPaymentParams{
        UserID:       event.AppUserID,
        ActionType:   actionType,
        RefID:        event.ID,  // RevenueCat event ID for idempotency
        ProductID:    event.ProductID,
        Price:        event.Price,
        Currency:     event.Currency,
        Store:        event.Store,
        PurchasedAt:  time.UnixMilli(event.PurchasedAtMs),
        ExpirationAt: time.UnixMilli(event.ExpirationAtMs),
    })
    if err != nil {
        if errors.Is(err, economy.ErrAlreadyProcessed) {
            return &RevenueCatWebhookResponse{
                Success: true,
                EventID: event.ID,
                Action:  "already_processed",
            }, nil
        }
        return nil, fmt.Errorf("process payment: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit tx: %w", err)
    }

    return &RevenueCatWebhookResponse{
        Success:      true,
        EventID:      event.ID,
        Action:       "wings_credited",
        WingsAwarded: result.WingsAwarded,
    }, nil
}
```

---

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `api/api_webhook.go` | CREATE | RevenueCat webhook handler |
| `api/api_webhook_test.go` | CREATE | API tests |
| `api/route_webhook.go` | CREATE | Webhook routes (no auth middleware) |
| `api/route_paths.go` | MODIFY | Add `PathWebhookRevenueCat` |
| `business/domain/economy/adapters.go` | MODIFY | Add `subscriptionProcessor` interface |
| `business/domain/economy/economy.go` | MODIFY | Add `ProcessRevenueCatEvent` |
| `business/domain/economy/models.go` | MODIFY | Add request/response types |
| `lib/economy/subscription.go` | CREATE | `ProcessPayment` logic |
| `lib/economy/errors.go` | MODIFY | Add `ErrAlreadyProcessed` |
| `config.go` | MODIFY | Add `RevenueCatWebhookSecret` |

---

## API Test Cases

| Test | Setup | Assert |
|------|-------|--------|
| `success-initial-purchase-credits-wings` | Valid payload, INITIAL_PURCHASE | 200, wings_credited, DB has transaction |
| `success-renewal-credits-wings` | Valid payload, RENEWAL | 200, wings_credited, DB has transaction |
| `success-idempotent-same-event-id` | Same event ID twice | 200, already_processed |
| `success-ignored-cancellation` | CANCELLATION event | 200, ignored |
| `success-ignored-unknown-product` | Unknown product_id | 200, ignored |
| `error-invalid-webhook-secret` | Wrong secret header | 401 |
| `error-invalid-payload` | Malformed JSON | 400 |
| `error-user-not-found` | Non-existent user ID | 404 or create user totals |
| `success-wingedx-no-wings` | WingedX product | 200, wings_credited, 0 wings |
| `success-sets-expiry` | Any purchase | Transaction has expires_at |

---

## Security

1. **Webhook Secret Validation**
   - Header: `X-RevenueCat-Webhook-Secret`
   - Compare against `cfg.RevenueCatWebhookSecret`
   - Return 401 if mismatch

2. **No User Auth Required**
   - Webhook endpoint is server-to-server
   - Protected by webhook secret only

3. **Idempotency**
   - Use `event.id` as `RefID` in action log
   - Check for existing action log with same RefID before processing

---

## Post-Implementation Checklist

- [ ] Add `RevenueCatWebhookSecret` to config
- [ ] Create webhook handler in `api/api_webhook.go`
- [ ] Create routes in `api/route_webhook.go`
- [ ] Add `ProcessRevenueCatEvent` to business layer
- [ ] Add `ProcessPayment` to lib layer (or reuse existing)
- [ ] Add `ErrAlreadyProcessed` error
- [ ] Write API tests
- [ ] Test with RevenueCat sandbox
- [ ] Verify wings credited with expiry

---

## Example Test

```go
func TestWebhookRevenueCat_InitialPurchase(t *testing.T) {
    t.Parallel()
    tSuite := testsuite.New(t)
    defer tSuite.UseBackendDB()()

    user := tSuite.PersistRegisteredUser()

    payload := RevenueCatWebhookRequest{
        APIVersion: "1.0",
        Event: RevenueCatEvent{
            ID:            "evt_123",
            Type:          "INITIAL_PURCHASE",
            AppUserID:     user.ID,
            ProductID:     "com.app.wingedplus_weekly",
            PurchasedAtMs: time.Now().UnixMilli(),
            Price:         14.95,
            Currency:      "USD",
            Store:         "APP_STORE",
        },
    }

    resp := tSuite.Request().
        Post(api.PathWebhookRevenueCat).
        Header("X-RevenueCat-Webhook-Secret", "test-secret").
        JSON(payload).
        Expect().
        Status(200).
        JSON().Object()

    resp.Value("success").Boolean().IsTrue()
    resp.Value("action").String().IsEqual("wings_credited")
    resp.Value("wings_awarded").Number().IsEqual(25)

    // Verify DB
    txn := tSuite.GetTransactionByUserID(user.ID)
    assert.Equal(t, 25, txn.Amount)
    assert.True(t, txn.ExpiresAt.Valid)
}
```
