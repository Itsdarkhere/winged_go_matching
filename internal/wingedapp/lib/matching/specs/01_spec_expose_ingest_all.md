# Spec: Admin Batch Ingestion

## Overview

Admin-only POST endpoint to trigger the matching ingestion process for all active users, creating a new match set with unique pairings.

## Endpoint

```
POST /admin/user/matching/batch
```

## Request Body

None required.

## Response

Returns the created `MatchSet` object on success (200 OK).

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "2024-01-15T10:30:00Z",
  "number_of_participants": 42
}
```

## Error Responses

| Status | Description |
|--------|-------------|
| 403 Forbidden | User is not an admin |
| 500 Internal Server Error | Ingestion failed |

## Architecture

### Files Modified

1. **`business/domain/matching/adapters.go`** - Added `ingester` interface
2. **`business/domain/matching/matching.go`** - Added `ingester` to Business struct and constructor
3. **`business/domain/matching/matching_admin.go`** - Added `AdminIngestAll()` method
4. **`api/route_paths.go`** - Added `PathAdminMatchBatch` constant
5. **`api/api_matching.go`** - Added `adminIngestAll` handler
6. **`api/route_matching.go`** - Added POST route for batch ingestion
7. **`di/deps_biz.go`** - Updated DI to pass `ingester` to matching business

### Tests

Added to `api/api_matching_test.go`:
- `fail-user-not-admin` - Returns 403 for non-admin users
- `success-basic-response-structure` - Validates response has ID, Name, NumberOfParticipants
- `success-creates-match-set-in-db` - Verifies match set is persisted
- `success-creates-correct-match-results-count` - Verifies n*(n-1)/2 pairs created
- `success-returns-valid-uuid` - Verifies UUID is valid

## Usage Example

```bash
curl -X POST \
  -H "Authorization: Bearer <admin_token>" \
  /api/v1/admin/user/matching/batch
```
