# Spec: Admin Update Match Config

## Overview

Admin-only PATCH endpoint to update match configuration settings. Supports partial updates (PATCH semantics - only provided fields are updated).

## Endpoint

```
PATCH /admin/user/matching/config
```

## Request Body

All fields are optional. Only provided fields are updated.

```json
{
  "age_range_start": 21,
  "age_range_end": 45,
  "age_range_woman_older_by": 3,
  "age_range_man_older_by": 7,
  "height_male_greater_by_cm": 10.5,
  "location_radius_km": 50.0,
  "location_adaptive_expansion": [10, 20, 30],
  "drop_hours": ["19:00", "20:00"],
  "drop_hours_utc": ["GMT+8"],
  "stale_chat_nudge": 24,
  "stale_chat_agent_setup": 48,
  "match_expiration_hours": 72,
  "match_block_declined": 7,
  "match_block_ignored": 14,
  "match_block_closed": 30,
  "score_range_start": 0.5,
  "score_range_end": 1.0
}
```

## Response

Returns the full updated `Config` object on success (200 OK).

## Validations

| Field | Rule | Error |
|-------|------|-------|
| `location_adaptive_expansion` | Must be strictly incrementing array (e.g., [10, 20, 30]) | `ErrAdaptiveExpansionNotIncrementing` |
| `age_range_start`, `age_range_end` | start <= end (when both provided) | `ErrAgeRangeStartGreaterThanEnd` |
| `score_range_start`, `score_range_end` | start <= end (when both provided) | `ErrScoreRangeStartGreaterThanEnd` |
| `drop_hours` | Valid HH:MM format | `ErrDropHoursInvalidFormat` |
| `drop_hours_utc` | Valid timezone format (GMT+N or UTC+N) | `ErrDropHoursUTCInvalidFormat` |
| Numeric fields | Must be non-negative | `ErrNegativeValue` |

## Architecture

### Files Modified

1. **`lib/matching/models.go`** - Added `UpdateMatchConfig` struct with nullable fields
2. **`lib/matching/errors.go`** - Added validation error sentinels
3. **`lib/matching/adapters.go`** - Added `Update()` to `configStorer` interface
4. **`lib/matching/store/config.go`** - Implemented `ConfigStore.Update()` with validation
5. **`lib/matching/config.go`** - Implemented `Logic.UpdateConfig()` with transaction
6. **`business/domain/matching/adapters.go`** - Added `configUpdater` interface
7. **`business/domain/matching/matching.go`** - Added `configUpdater` dependency
8. **`business/domain/matching/matching_admin.go`** - Added `AdminUpdateConfig()` method
9. **`api/api_matching.go`** - Added `adminUpdateMatchConfig` handler
10. **`api/route_matching.go`** - Added route registration
11. **`api/route_paths.go`** - Added `PathAdminMatchConfig` constant
12. **`di/deps_biz.go`** - Updated DI to pass `configUpdater`

### Tests

Added to `api/api_matching_test.go`:
- `success-update-all-fields` - Updates all config fields
- `success-partial-update` - Partial update with subset of fields
- `error-adaptive-expansion-not-incrementing` - Validation: non-incrementing array
- `error-age-range-start-greater-than-end` - Validation: age range consistency
- `error-invalid-drop-hours-format` - Validation: HH:MM format
- `error-negative-value` - Validation: non-negative numeric values

## Usage Example

```bash
curl -X PATCH \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"location_radius_km": 75.0, "match_expiration_hours": 48}' \
  /api/v1/admin/user/matching/config
```
