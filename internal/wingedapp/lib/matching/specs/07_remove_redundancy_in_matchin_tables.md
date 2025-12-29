#### Tech Specs: remove redundant in matching tables

## Problem
The `user_match` table is redundant - it stores `(user_id, match_result_id)` pairs, but this relationship can be derived directly from `match_result` via:
```sql
SELECT * FROM match_result WHERE user_a_ref_id = ? OR user_b_ref_id = ?
```

## Solution - COMPLETED

### 1. Migration Created
- `26_drop_user_match_table.up.sql` - Drops the `user_match` table
- `26_drop_user_match_table.down.sql` - Recreates it (for rollback)

### 2. Code Already Using Derived Query
The `UserMatchActionsStore` already queries directly from `match_result`:
```go
// UserMatches returns dropped matches for a user from their perspective.
qMods := []qm.QueryMod{
    qm.From(matchResultTbl+" mr"),
    qm.Where("(mr.user_a_ref_id = ? OR mr.user_b_ref_id = ?)", f.UserID.String(), f.UserID.String()),
    // ...
}
```
This is the correct pattern - no need for a separate `user_match` table.

### 3. Cleanup Done
- Deleted factory: `db/factory/user_match.go`
- Removed from `apprepo.go`: `toDeleteBackendApp` no longer references `user_match`
- Updated tests: Removed `user_match` creation/verification from:
  - `apprepo_test.go`
  - `api_matching_ingest_test.go`

### 4. Post-Deploy Steps
After running migration `26`, regenerate SQLBoiler models:
```bash
make wingedapp-pgmodel-gen
```
This will remove `db/pgmodel/user_match.go` and update relationship code in `match_result.go` and `users.go`.