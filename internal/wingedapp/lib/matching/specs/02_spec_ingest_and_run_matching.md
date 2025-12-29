# Spec: IngestAll + Run Matching Algorithm

## Overview

Extend the batch ingestion flow to not only create a MatchSet with all user pairings, but also run the matching algorithm (`ProcessMatchResult`) on each pairing. The lib layer provides separate functions, and the business layer assembles them.

## Endpoint

```
POST /admin/user/matching/batch
```

(No change - same endpoint, enhanced functionality)

## Architecture

### Separation of Concerns

**Lib Layer (`lib/matching/`):**
- `IngestAll()` - Creates MatchSet and MatchResults (unprocessed)
- `RunIngestionSet()` - Runs matching algorithm on all pairs in a MatchSet

**Business Layer (`business/domain/matching/`):**
- `AdminIngestAll()` - Assembles both operations: ingest + run matching

This separation allows:
1. Testing each operation independently
2. Reusing `RunIngestionSet` for other scenarios
3. Future flexibility (e.g., async processing)

## Implementation

### Files Modified

| File | Changes |
|------|---------|
| `lib/matching/ingestion.go` | Implemented `RunIngestionSet()` |
| `business/domain/matching/adapters.go` | Added `matchRunner` interface |
| `business/domain/matching/matching.go` | Added `matchRunner` dependency to Business struct |
| `business/domain/matching/matching_admin.go` | Updated `AdminIngestAll()` to call both operations |
| `di/deps_biz.go` | Updated DI to pass `matchRunner` (7th parameter) |
| `lib/matching/run_ingestion_set_test.go` | New test file with 7 comprehensive tests |

### RunIngestionSet Implementation

```go
func (l *Logic) RunIngestionSet(ctx context.Context, exec boil.ContextExecutor, matchSetID uuid.UUID) error {
    // 1. Fetch all MatchResults for this MatchSet
    results, err := l.matchResultStorer.MatchResults(ctx, exec, &QueryFilterMatchResult{
        MatchSetID: null.StringFrom(matchSetID.String()),
    })

    // 2. Process each MatchResult through the matching algorithm
    for i := range results.Data {
        if _, err := l.ProcessMatchResult(ctx, exec, &results.Data[i]); err != nil {
            // Log error but continue - one bad pair shouldn't stop the batch
            continue
        }
    }

    return nil
}
```

### AdminIngestAll (Business Layer Assembly)

```go
func (b *Business) AdminIngestAll(ctx context.Context) (*matchLib.MatchSet, error) {
    // Step 1: Create MatchSet and MatchResults for all active users
    matchSet, err := b.ingester.IngestAll(ctx)
    if err != nil {
        return nil, fmt.Errorf("ingest all: %w", err)
    }

    // Step 2: Run matching algorithm on all pairs in the MatchSet
    if err = b.matchRunner.RunIngestionSet(ctx, b.transactor.DB(), matchSet.ID); err != nil {
        return nil, fmt.Errorf("run ingestion set: %w", err)
    }

    return matchSet, nil
}
```

## Flow Diagram

```
POST /admin/user/matching/batch
         │
         ▼
    API Handler (adminIngestAll)
         │
         ▼
    Business Layer (AdminIngestAll)
         │
         ├──► Step 1: ingester.IngestAll()
         │         │
         │         ├─ Fetch active users
         │         ├─ Create MatchSet
         │         └─ Create MatchResult for each unique pair
         │
         └──► Step 2: matchRunner.RunIngestionSet()
                   │
                   ├─ Fetch all MatchResults for MatchSet
                   └─ For each MatchResult:
                          │
                          └─ ProcessMatchResult()
                                  │
                                  ├─ Run hard qualifiers (age, prefs, height, distance)
                                  ├─ Save QualifierResults JSON
                                  └─ If hard pass → run qualitative matching
```

## Tests

### Lib Layer Tests (`run_ingestion_set_test.go`)

| Test | Description |
|------|-------------|
| `success-processes-all-match-results` | Verifies all MatchResults get QualifierResults populated |
| `success-stores-qualifier-results-json` | Verifies JSON contains qualifier names |
| `success-hard-qualifier-failures-stored-not-thrown` | Verifies errors stored in JSON, not thrown |
| `success-qualitative-match-runs-when-hard-pass` | Verifies MatchedQualitatively=true when all pass |
| `success-empty-match-set-no-error` | Handles empty MatchSet gracefully |
| `success-updates-matched-qualitatively-flag-false-on-hard-fail` | Verifies flag is false when hard qualifiers fail |
| `success-multiple-pairs-all-processed` | Verifies all 15 pairs processed for 6 users |

### API Layer Tests (`api_matching_test.go`)

| Test | Description |
|------|-------------|
| `fail-user-not-admin` | Returns 403 for non-admin users |
| `success-basic-response-structure` | Validates response has ID, Name, NumberOfParticipants |
| `success-creates-match-set-in-db` | Verifies MatchSet persisted |
| `success-creates-correct-match-results-count` | Verifies n*(n-1)/2 pairs using `UserPairsUniqPerm` |
| `success-returns-valid-uuid` | Verifies UUID is valid |

## Error Handling

- **Soft failures:** Hard qualifier failures stored in `QualifierResults` JSON - processing continues
- **Hard failures:** Database errors logged - processing continues to next pair
- **Design principle:** One bad pair shouldn't stop the entire batch

## Response

**Success (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "2024-01-15T10:30:00Z",
  "number_of_participants": 42
}
```

After the response, all MatchResults in the MatchSet will have:
- `QualifierResults` JSON populated with hard qualifier results
- `MatchedQualitatively` set (true if passed all qualifiers, false otherwise)
- `IsPossibleMatch` set (true if viable match)

## Usage Example

```bash
curl -X POST \
  -H "Authorization: Bearer <admin_token>" \
  /api/v1/admin/user/matching/batch
```
