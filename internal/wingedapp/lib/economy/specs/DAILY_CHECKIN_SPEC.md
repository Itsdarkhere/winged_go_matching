# Daily Check-in Spec

## TL;DR
User checks in daily → creates unclaimed transaction → accumulates up to 7 max → user claims → wings credited to balance. No expiration, just cap at 7 unclaimed.

---

## De Facto Implementation Strategy (from `lib/matching/store/`)

```
Store Layer (lib/*/store/) - DATA ACCESS ONLY
├── Reads: pgmodel.NewQuery(qMods...).Bind() → domain types (hybrid type-safe)
├── Writes: ALWAYS call db/repo/ - NEVER insert/update directly
├── CRUD naming: Entity(), Entities(), EntityCount(), Insert(), Update(), Delete()
├── Interfaces defined in lib/adapters.go (NOT in store package)
└── NO business logic

Repo Layer (db/repo/) - MUTATIONS
├── InsertX(), UpdateX(), DeleteX() methods
├── Uses pgmodel directly
└── Returns *pgmodel.X

Logic Layer (lib/*/) - BUSINESS RULES + DATA COMPOSITION
├── Composes store calls
├── Business validation
├── Orchestrates multi-store operations
└── Where "ClaimWings", "PerformCheckin" etc. live
```

**Key Distinction:**
- Store = `Transactions()`, `Insert()` (data access)
- Logic = `PerformCheckin()`, `ClaimWings()` (business methods)

---

## Implementation Plan

### 1. NEW: DailyCheckinStore (store/daily_checkin.go)

Store = data access only. CRUD naming. Uses repo for mutations.

```go
package store

import (
    "context"
    "database/sql"
    "fmt"
    "wingedapp/pgtester/internal/wingedapp/db/pgmodel"
    "wingedapp/pgtester/internal/wingedapp/db/repo"
    "wingedapp/pgtester/internal/wingedapp/lib/applog"
    "wingedapp/pgtester/internal/wingedapp/lib/economy"
    "github.com/aarondl/sqlboiler/v4/boil"
    "github.com/aarondl/sqlboiler/v4/queries/qm"
)

type DailyCheckinStore struct {
    l    applog.Logger
    repo *repo.Store
}

func NewDailyCheckinStore(l applog.Logger) *DailyCheckinStore {
    return &DailyCheckinStore{l: l, repo: &repo.Store{}}
}
```

#### Select: Transactions (get unclaimed)

```go
// Transactions returns check-in transactions matching filter
func (s *DailyCheckinStore) Transactions(
    ctx context.Context,
    exec boil.ContextExecutor,
    f *economy.QueryFilterCheckinTransaction,
) ([]economy.CheckinTransaction, error) {
    var results []economy.CheckinTransaction

    tbl := pgmodel.TableNames.WingsEcnTransaction
    cols := pgmodel.WingsEcnTransactionColumns

    qMods := []qm.QueryMod{
        qm.Select(
            "t."+cols.ID+" AS id",
            "t."+cols.UserRefID+" AS user_id",
            "t."+cols.Amount+" AS amount",
            "t."+cols.Claimed+" AS claimed",
            "t."+cols.CreatedDate+" AS created_date",
        ),
        qm.From(tbl+" t"),
        qm.Where("t."+cols.ActionLogType+" = ?", string(economy.ActionDailyCheckIn)),
        qm.Where("t."+cols.IsActive+" = ?", 1),
    }

    if f.UserID.Valid {
        qMods = append(qMods, qm.Where("t."+cols.UserRefID+" = ?", f.UserID.String))
    }
    if f.Claimed.Valid {
        qMods = append(qMods, qm.Where("t."+cols.Claimed+" = ?", f.Claimed.Bool))
    }

    qMods = append(qMods, qm.OrderBy("t."+cols.CreatedDate+" ASC"))

    if err := pgmodel.NewQuery(qMods...).Bind(ctx, exec, &results); err != nil {
        return nil, fmt.Errorf("query checkin transactions: %w", err)
    }
    return results, nil
}
```

#### Select: Transaction (single)

```go
// Transaction returns single transaction matching filter
func (s *DailyCheckinStore) Transaction(
    ctx context.Context,
    exec boil.ContextExecutor,
    f *economy.QueryFilterCheckinTransaction,
) (*economy.CheckinTransaction, error) {
    txns, err := s.Transactions(ctx, exec, f)
    if err != nil {
        return nil, err
    }
    if len(txns) != 1 {
        return nil, fmt.Errorf("transaction count mismatch, have %d, want 1", len(txns))
    }
    return &txns[0], nil
}
```

#### Select: TransactionCount

```go
// TransactionCount returns count of transactions matching filter
func (s *DailyCheckinStore) TransactionCount(
    ctx context.Context,
    exec boil.ContextExecutor,
    f *economy.QueryFilterCheckinTransaction,
) (int, error) {
    cols := pgmodel.WingsEcnTransactionColumns

    qMods := []qm.QueryMod{
        qm.Where(cols.ActionLogType+" = ?", string(economy.ActionDailyCheckIn)),
        qm.Where(cols.IsActive+" = ?", 1),
    }

    if f.UserID.Valid {
        qMods = append(qMods, qm.Where(cols.UserRefID+" = ?", f.UserID.String))
    }
    if f.Claimed.Valid {
        qMods = append(qMods, qm.Where(cols.Claimed+" = ?", f.Claimed.Bool))
    }

    count, err := pgmodel.WingsEcnTransactions(qMods...).Count(ctx, exec)
    if err != nil {
        return 0, fmt.Errorf("count checkin transactions: %w", err)
    }
    return int(count), nil
}
```

#### Select: LastTransaction (most recent check-in)

```go
// LastTransaction returns most recent check-in transaction for user
func (s *DailyCheckinStore) LastTransaction(
    ctx context.Context,
    exec boil.ContextExecutor,
    userID string,
) (*economy.CheckinTransaction, error) {
    var result economy.CheckinTransaction

    cols := pgmodel.WingsEcnTransactionColumns

    err := pgmodel.WingsEcnTransactions(
        qm.Select(cols.ID, cols.UserRefID, cols.Amount, cols.Claimed, cols.CreatedDate),
        qm.Where(cols.UserRefID+" = ?", userID),
        qm.Where(cols.ActionLogType+" = ?", string(economy.ActionDailyCheckIn)),
        qm.Where(cols.IsActive+" = ?", 1),
        qm.OrderBy(cols.CreatedDate+" DESC"),
        qm.Limit(1),
    ).Bind(ctx, exec, &result)

    if err == sql.ErrNoRows {
        return nil, nil  // no check-in yet
    }
    if err != nil {
        return nil, fmt.Errorf("last checkin transaction: %w", err)
    }
    return &result, nil
}
```

#### Insert: calls repo

```go
// Insert creates a new check-in transaction via repo
func (s *DailyCheckinStore) Insert(
    ctx context.Context,
    exec boil.ContextExecutor,
    inserter *economy.InsertCheckinTransaction,
) (*economy.CheckinTransaction, error) {
    pgRes, err := s.repo.InsertWingsEcnTransaction(ctx, exec, &repo.InsertWingsEcnTransaction{
        UserRefID:      inserter.UserID,
        ActionLogType:  string(economy.ActionDailyCheckIn),
        ActionLogRefID: inserter.ActionLogID,
        IsCredit:       true,
        Claimed:        false,  // unclaimed by default
        Amount:         inserter.Amount,
    })
    if err != nil {
        return nil, fmt.Errorf("insert checkin transaction: %w", err)
    }

    return &economy.CheckinTransaction{
        ID:          pgRes.ID,
        UserID:      pgRes.UserRefID,
        Amount:      pgRes.Amount,
        Claimed:     pgRes.Claimed,
        CreatedDate: pgRes.CreatedDate,
    }, nil
}
```

#### Update: batch claim (raw SQL for efficiency)

```go
// UpdateClaimBatch marks multiple transactions as claimed
func (s *DailyCheckinStore) UpdateClaimBatch(
    ctx context.Context,
    exec boil.ContextExecutor,
    transactionIDs []string,
) (int, error) {
    if len(transactionIDs) == 0 {
        return 0, nil
    }

    tbl := pgmodel.TableNames.WingsEcnTransaction
    cols := pgmodel.WingsEcnTransactionColumns

    query := fmt.Sprintf(`
        UPDATE %s
        SET %s = TRUE, %s = NOW()
        WHERE %s = ANY($1) AND %s = FALSE
    `, tbl, cols.Claimed, cols.LastUpdated, cols.ID, cols.Claimed)

    result, err := exec.ExecContext(ctx, query, transactionIDs)
    if err != nil {
        return 0, fmt.Errorf("update claim batch: %w", err)
    }

    rowsAffected, _ := result.RowsAffected()
    return int(rowsAffected), nil
}
```

---

### 2. NEW: Models (models.go)

```go
// CheckinTransaction represents a daily check-in transaction
type CheckinTransaction struct {
    ID          string    `boil:"id"`
    UserID      string    `boil:"user_id"`
    Amount      int       `boil:"amount"`
    Claimed     bool      `boil:"claimed"`
    CreatedDate time.Time `boil:"created_date"`
}

// QueryFilterCheckinTransaction for filtering transactions
type QueryFilterCheckinTransaction struct {
    UserID  null.String
    Claimed null.Bool
}

// InsertCheckinTransaction for inserting new transaction
type InsertCheckinTransaction struct {
    UserID      string
    ActionLogID string
    Amount      int
}

// CheckinStatus for API response
type CheckinStatus struct {
    CheckedInToday bool `json:"checked_in_today"`
    UnclaimedWings int  `json:"unclaimed_wings"`
    CanClaim       bool `json:"can_claim"`
}
```

---

### 3. UPDATE: EconomyStores (store/store.go)

```go
type EconomyStores struct {
    // existing...
    DailyCheckinStore *DailyCheckinStore
}

func NewEconomyStores(l applog.Logger) *EconomyStores {
    return &EconomyStores{
        DailyCheckinStore: NewDailyCheckinStore(l),
    }
}
```

---

### 4. NEW: Logic Layer (daily_checkin.go)

Business logic lives here. Composes store calls.

```go
package economy

//counterfeiter:generate . dailyCheckinStorer
type dailyCheckinStorer interface {
    Transactions(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterCheckinTransaction) ([]CheckinTransaction, error)
    TransactionCount(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterCheckinTransaction) (int, error)
    LastTransaction(ctx context.Context, exec boil.ContextExecutor, userID string) (*CheckinTransaction, error)
    Insert(ctx context.Context, exec boil.ContextExecutor, inserter *InsertCheckinTransaction) (*CheckinTransaction, error)
    UpdateClaimBatch(ctx context.Context, exec boil.ContextExecutor, transactionIDs []string) (int, error)
}

type DailyCheckinLogic struct {
    logger           applog.Logger
    checkinStore     dailyCheckinStorer
    userTotalsStore  userTotalsStorer
    actionLogStore   actionLogStorer
}

// PerformCheckin - business logic for daily check-in
func (d *DailyCheckinLogic) PerformCheckin(
    ctx context.Context,
    exec boil.ContextExecutor,
    userID string,
) error {
    // 1. Check if already checked in today
    last, err := d.checkinStore.LastTransaction(ctx, exec, userID)
    if err != nil {
        return fmt.Errorf("get last transaction: %w", err)
    }
    if last != nil && isToday(last.CreatedDate) {
        return ErrAlreadyCheckedInToday
    }

    // 2. Check cap: max 7 unclaimed
    count, err := d.checkinStore.TransactionCount(ctx, exec, &QueryFilterCheckinTransaction{
        UserID:  null.StringFrom(userID),
        Claimed: null.BoolFrom(false),
    })
    if err != nil {
        return fmt.Errorf("count unclaimed: %w", err)
    }
    if count >= MaxUnclaimedCheckins {
        return ErrMaxUnclaimedReached
    }

    // 3. Insert action log
    actionLog, err := d.actionLogStore.Insert(ctx, exec, string(ActionDailyCheckIn), &InsertActionLog{
        UserID: userID,
        RefID:  uuid.New().String(),
        Type:   ActionDailyCheckIn,
    })
    if err != nil {
        return fmt.Errorf("insert action log: %w", err)
    }

    // 4. Insert unclaimed transaction
    _, err = d.checkinStore.Insert(ctx, exec, &InsertCheckinTransaction{
        UserID:      userID,
        ActionLogID: actionLog.ID,
        Amount:      DailyCheckinWings,
    })
    if err != nil {
        return fmt.Errorf("insert transaction: %w", err)
    }

    return nil
}

// ClaimWings - business logic for claiming wings
func (d *DailyCheckinLogic) ClaimWings(
    ctx context.Context,
    exec boil.ContextExecutor,
    userID string,
) (int, error) {
    // 1. Get unclaimed transactions
    unclaimed, err := d.checkinStore.Transactions(ctx, exec, &QueryFilterCheckinTransaction{
        UserID:  null.StringFrom(userID),
        Claimed: null.BoolFrom(false),
    })
    if err != nil {
        return 0, fmt.Errorf("get unclaimed: %w", err)
    }
    if len(unclaimed) == 0 {
        return 0, nil
    }

    // 2. Sum amounts + collect IDs
    total := 0
    ids := make([]string, len(unclaimed))
    for i, tx := range unclaimed {
        total += tx.Amount
        ids[i] = tx.ID
    }

    // 3. Mark as claimed (batch update)
    _, err = d.checkinStore.UpdateClaimBatch(ctx, exec, ids)
    if err != nil {
        return 0, fmt.Errorf("claim batch: %w", err)
    }

    // 4. Add wings to user_totals.total_wings
    totals, err := d.userTotalsStore.Totals(ctx, exec, userID)
    if err != nil {
        return 0, fmt.Errorf("get totals: %w", err)
    }
    err = d.userTotalsStore.Update(ctx, exec, &UpdateUserTotals{
        ID:    totals.ID,
        Wings: null.IntFrom(totals.Wings + total),
    })
    if err != nil {
        return 0, fmt.Errorf("update totals: %w", err)
    }

    return total, nil
}

// GetStatus - get current check-in status
func (d *DailyCheckinLogic) GetStatus(
    ctx context.Context,
    exec boil.ContextExecutor,
    userID string,
) (*CheckinStatus, error) {
    // Get unclaimed transactions
    unclaimed, err := d.checkinStore.Transactions(ctx, exec, &QueryFilterCheckinTransaction{
        UserID:  null.StringFrom(userID),
        Claimed: null.BoolFrom(false),
    })
    if err != nil {
        return nil, fmt.Errorf("get unclaimed: %w", err)
    }

    // Get last transaction
    last, err := d.checkinStore.LastTransaction(ctx, exec, userID)
    if err != nil {
        return nil, fmt.Errorf("get last: %w", err)
    }

    sum := 0
    for _, tx := range unclaimed {
        sum += tx.Amount
    }

    checkedInToday := last != nil && isToday(last.CreatedDate)

    return &CheckinStatus{
        CheckedInToday: checkedInToday,
        UnclaimedWings: sum,
        CanClaim:       sum > 0,
    }, nil
}

func isToday(t time.Time) bool {
    now := time.Now().UTC()
    return t.UTC().Year() == now.Year() && t.UTC().YearDay() == now.YearDay()
}
```

---

### 5. NEW: Constants & Errors

```go
// consts.go
const MaxUnclaimedCheckins = 7

// errors.go
var ErrAlreadyCheckedInToday = errors.New("already checked in today")
var ErrMaxUnclaimedReached = errors.New("max unclaimed check-ins reached (7)")
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/users/economy/checkin` | Get check-in status |
| POST | `/users/economy/checkin` | Perform daily check-in |
| POST | `/users/economy/checkin/claim` | Claim accumulated wings |

---

## Implementation Order

1. **models.go** - Add `CheckinTransaction`, `QueryFilterCheckinTransaction`, `InsertCheckinTransaction`, `CheckinStatus`
2. **consts.go** - Add `MaxUnclaimedCheckins = 7`
3. **errors.go** - Add errors
4. **store/daily_checkin.go** - NEW store (CRUD: Transactions, TransactionCount, LastTransaction, Insert, UpdateClaimBatch)
5. **store/store.go** - Add to EconomyStores
6. **daily_checkin.go** - NEW logic layer (PerformCheckin, ClaimWings, GetStatus)
7. **DI** - Wire up
8. **API handlers**
9. **Tests**

---

## Key Pattern Differences

| Layer | Responsibility | Example |
|-------|---------------|---------|
| Store | Data access, CRUD | `Transactions()`, `Insert()`, `UpdateClaimBatch()` |
| Repo | Mutations via pgmodel | `repo.InsertWingsEcnTransaction()` |
| Logic | Business rules, orchestration | `PerformCheckin()`, `ClaimWings()` |

**Store does NOT have business methods like `ClaimCheckins()` - that's logic layer.**
