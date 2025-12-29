# Matching Domain Onboarding

> For Jr. Backend Engineers coming from MVC (IMapp) → DDD (WingedApp)

---

## Part 1: Product Sense (The "Why")

### What Is Matching?

Think Hinge/Tinder but **curated**:
1. **Algorithm pairs users** based on compatibility (age, location, preferences, personality)
2. **Admin reviews & approves** matches before users see them
3. **Users propose/pass** on matches they receive
4. **Mutual proposal** → unlocks scheduling a date

```
┌─────────────────────────────────────────────────────────────────────┐
│                     MATCHING LIFECYCLE                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐      │
│  │  BATCH   │───▶│  REVIEW  │───▶│   DROP   │───▶│  ACTION  │      │
│  │ (Admin)  │    │ (Admin)  │    │  (Job)   │    │  (User)  │      │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘      │
│       │               │               │               │            │
│       ▼               ▼               ▼               ▼            │
│   Algorithm      Approve or      Deliver to      Propose or        │
│   evaluates      reject each     user's inbox    Pass              │
│   all pairs      match                                             │
│                                                     │              │
│                                                     ▼              │
│                                           ┌─────────────────┐      │
│                                           │ MUTUAL PROPOSE? │      │
│                                           └────────┬────────┘      │
│                                                    │               │
│                                     ┌──────────────┴──────────┐    │
│                                     ▼                         ▼    │
│                              ┌──────────┐              ┌──────────┐│
│                              │SCHEDULING│              │ WAITING  ││
│                              │ UNLOCKED │              │  PARTNER ││
│                              └──────────┘              └──────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

### Why Curated?
- Quality > quantity (unlike swipe apps)
- Admin filters out bad matches algorithm missed
- Personality scoring via AI (not just demographics)

---

## Part 2: MVC vs DDD Mental Model

### What You Know (IMapp - MVC)

```
┌─────────────────────────────────────────────┐
│                   MVC                        │
├─────────────────────────────────────────────┤
│                                             │
│   Controller ──▶ Model ──▶ Database         │
│       │            │                        │
│       ▼            ▼                        │
│     View      ORM/ActiveRecord              │
│                                             │
│   • Routes hit controllers directly         │
│   • Models = DB tables + business logic     │
│   • Fat models, thin controllers            │
│                                             │
└─────────────────────────────────────────────┘
```

### What You're Learning (WingedApp - DDD)

```
┌─────────────────────────────────────────────────────────────────────┐
│                           DDD LAYERS                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌─────────┐   ┌──────────┐   ┌─────────┐   ┌───────┐   ┌──────┐ │
│   │   API   │──▶│ Business │──▶│   Lib   │──▶│ Store │──▶│  DB  │ │
│   │ Handler │   │  (Biz)   │   │ (Logic) │   │(Query)│   │Model │ │
│   └─────────┘   └──────────┘   └─────────┘   └───────┘   └──────┘ │
│                                                                     │
│   HTTP only     TX boundary    Pure logic    SQL only    SQLBoiler │
│   Parse req     Orchestrate    No database   No biz      Generated │
│   Format res    Call lib       Atomic ops    Pure CRUD             │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Translation Table

| MVC Concept | DDD Equivalent | Location |
|-------------|----------------|----------|
| Controller | API Handler | `api/api_matching.go` |
| Model (ActiveRecord) | Store + pgmodel | `lib/matching/store/` |
| Model (business logic) | Lib | `lib/matching/logic.go` |
| Service class | Business layer | `business/domain/matching/` |
| DB migration | SQL migration | `db/migration/` |
| Model instance | Domain type | `lib/matching/models.go` |

### Key Difference: Where Logic Lives

```
MVC:
  User.rb → validates, callbacks, business rules, DB ops (all in one)

DDD:
  api/        → "I parse HTTP and return JSON"
  business/   → "I own the transaction and call lib"
  lib/        → "I AM the business rules" ← LOGIC LIVES HERE
  store/      → "I just run SQL"
  pgmodel/    → "I'm auto-generated from DB schema"
```

---

## Part 3: Admin Matching Flow

### Admin Endpoints Overview

| Endpoint | What It Does |
|----------|--------------|
| `POST /admin/.../batch` | Run algorithm on all users |
| `GET /admin/.../` | List all match results |
| `GET /admin/.../{id}` | Single match detail |
| `POST /admin/.../approve` | Allow match to be dropped to users |
| `POST /admin/.../unapprove` | Revoke approval |
| `GET/PATCH /admin/.../config` | View/update matching settings |
| `POST /admin/.../ingest-csv` | Bulk upload test users |

### Batch Creation Flow (Deep Dive)

```
┌────────────────────────────────────────────────────────────────────────┐
│                    BATCH CREATION DATA FLOW                            │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│  POST /admin/users/matching/batch                                      │
│           │                                                            │
│           ▼                                                            │
│  ┌─────────────────┐                                                   │
│  │  API Handler    │  api/api_matching.go:adminMatchingBatch()         │
│  │  (Parse req)    │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                            │
│           ▼                                                            │
│  ┌─────────────────┐                                                   │
│  │    Business     │  business/domain/matching/matching_admin.go       │
│  │  (Begin TX)     │  → Ingest()                                       │
│  └────────┬────────┘                                                   │
│           │                                                            │
│           ▼                                                            │
│  ┌─────────────────┐                                                   │
│  │      Lib        │  lib/matching/ingestion.go                        │
│  │  (Algorithm)    │  → IngestAll() → RunIngestionSet()                │
│  └────────┬────────┘                                                   │
│           │                                                            │
│           ├──────────────────────────────────────────────┐             │
│           │                                              │             │
│           ▼                                              ▼             │
│  ┌─────────────────┐                            ┌─────────────────┐    │
│  │  Create Pairs   │                            │ Run Qualifiers  │    │
│  │  n*(n-1)/2      │                            │ For Each Pair   │    │
│  └─────────────────┘                            └────────┬────────┘    │
│                                                          │             │
│                                    ┌─────────────────────┼─────────┐   │
│                                    │                     │         │   │
│                                    ▼                     ▼         ▼   │
│                               ┌────────┐           ┌────────┐ ┌──────┐ │
│                               │  Age   │           │Distance│ │Height│ │
│                               │Qualify │           │Qualify │ │Qualif│ │
│                               └────────┘           └────────┘ └──────┘ │
│                                                                        │
│  Result: MatchSet + MatchResults with is_possible_match flags          │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

### Qualifiers Explained

**Hard Qualifiers** (must ALL pass):
- `qualifier_age.go` - Within each other's age range
- `qualifier_dating_prefs.go` - Gender preferences match
- `qualifier_height.go` - Height preference satisfied
- `qualifier_distance.go` - Within radius (e.g., 30km)

**Soft Qualifier**:
- `extmatcher/` - AI personality scoring

```go
// lib/matching/matching.go - ProcessMatchResult()
func (l *Lib) ProcessMatchResult(ctx context.Context, exec boil.ContextExecutor, mr *MatchResult) error {
    // Run each qualifier
    results := make(map[string]QualifierResult)

    for _, q := range l.qualifiers {
        result := q.Evaluate(userA, userB, config)
        results[q.Name()] = result

        if !result.Passed && q.IsHard() {
            mr.IsPossibleMatch = false  // Hard fail = no match
            return nil
        }
    }

    mr.IsPossibleMatch = true
    mr.QualifierResults = results  // Stored as JSONB
    return nil
}
```

### Admin Review & Approval

```
┌────────────────────────────────────────────────────────────────────┐
│                     ADMIN REVIEW FLOW                              │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  GET /admin/users/matching?is_possible_match=1&is_approved=0       │
│                          │                                         │
│                          ▼                                         │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Match List with Filters                                     │  │
│  │                                                              │  │
│  │  • user_id - either user in pair                             │  │
│  │  • is_possible_match - algorithm approved                    │  │
│  │  • is_approved - admin approved                              │  │
│  │  • is_dropped - delivered to users                           │  │
│  │  • match_set_id - filter by batch                            │  │
│  │  • enrich_users=1 - include full user profiles               │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                          │                                         │
│                          ▼                                         │
│  Admin sees: User A ↔ User B, qualifier scores, distance           │
│                          │                                         │
│                          ▼                                         │
│  POST /admin/.../matches/{id}/approve                              │
│  → Sets is_approved = true                                         │
│  → Match eligible for next drop cycle                              │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## Part 4: User Matching Flow

### User Endpoints Overview

| Endpoint | What It Does |
|----------|--------------|
| `GET /users/matching/matches` | My inbox of matches |
| `GET /users/matching/matches/{id}` | Single match detail |
| `POST /users/matching/matches/{id}/propose` | Express interest |
| `POST /users/matching/matches/{id}/pass` | Decline (permanent) |
| `PATCH /users/matching/seen` | Mark as viewed |
| `GET /users/matching/matches/{id}/overlaps` | Scheduling times (after mutual) |

### User Viewing Matches

```
┌────────────────────────────────────────────────────────────────────┐
│                   USER MATCH INBOX QUERY                           │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  GET /users/matching/matches                                       │
│           │                                                        │
│           ▼                                                        │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Store Query (lib/matching/store/user_match_actions.go)     │   │
│  │                                                             │   │
│  │  SELECT                                                     │   │
│  │    CASE                                                     │   │
│  │      WHEN user_a_ref_id = $1 THEN user_b_ref_id             │   │
│  │      ELSE user_a_ref_id                                     │   │
│  │    END as partner_id,                                       │   │
│  │    ...                                                      │   │
│  │  FROM match_result                                          │   │
│  │  WHERE (user_a_ref_id = $1 OR user_b_ref_id = $1)           │   │
│  │    AND is_dropped = true    -- delivered to user            │   │
│  │    AND is_approved = true   -- admin approved               │   │
│  │    AND is_expired = false   -- not timed out                │   │
│  │                                                             │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                    │
│  Key insight: Same match_result row, different "view" per user     │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

### Propose/Pass State Machine

```
┌────────────────────────────────────────────────────────────────────┐
│                    USER ACTION STATE MACHINE                       │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│                    ┌──────────┐                                    │
│                    │ PENDING  │ ← Initial state                    │
│                    └────┬─────┘                                    │
│                         │                                          │
│           ┌─────────────┴─────────────┐                            │
│           │                           │                            │
│           ▼                           ▼                            │
│    ┌──────────┐                ┌──────────┐                        │
│    │ PROPOSED │                │  PASSED  │ ← PERMANENT            │
│    └────┬─────┘                └──────────┘   (no undo)            │
│         │                                                          │
│         │                                                          │
│         ▼                                                          │
│  ┌─────────────────────────────────────────────────┐               │
│  │  Check partner's action                         │               │
│  │                                                 │               │
│  │  Partner = PENDING?  → Wait for them            │               │
│  │  Partner = PASSED?   → Match dead               │               │
│  │  Partner = PROPOSED? → MUTUAL! Create date      │               │
│  │                                                 │               │
│  └─────────────────────────────────────────────────┘               │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

### Mutual Proposal → Scheduling

```go
// lib/matching/logic.go - ProposeMatch()
func (l *Lib) ProposeMatch(ctx context.Context, exec boil.ContextExecutor, userID, matchID uuid.UUID) (*ProposeResult, error) {
    // 1. Get match, verify user is part of it
    match, _ := l.store.MatchResult(ctx, exec, matchID)

    // 2. Determine if user is A or B
    isUserA := match.UserARefID == userID

    // 3. Update user's action to "Proposed"
    // ...

    // 4. Check if partner already proposed
    partnerAction := match.UserBActionRefID  // (or A if we're B)

    if partnerAction == ProposedActionID {
        // MUTUAL PROPOSAL!
        // → Create date_instance
        // → Update match lifecycle to "Scheduling"
        // → Return mutual_proposal: true, date_instance_id: ...
    }

    return &ProposeResult{MutualProposal: false}, nil
}
```

---

## Part 5: Code Navigation Guide

### File Map

```
internal/wingedapp/
├── api/
│   ├── api_matching.go          ← HTTP handlers (13 endpoints)
│   ├── route_matching.go        ← Route registration + middleware
│   └── consts.go                ← Query param constants
│
├── business/domain/matching/
│   ├── matching.go              ← User-facing TX boundaries
│   └── matching_admin.go        ← Admin TX boundaries
│
├── lib/matching/
│   ├── matching.go              ← Qualifier runner (ProcessMatchResult)
│   ├── logic.go                 ← ProposeMatch, PassMatch, MarkSeen
│   ├── ingestion.go             ← Batch creation (IngestAll)
│   ├── drops.go                 ← Drop scheduler integration
│   ├── models.go                ← Domain types (MatchResult, Config, etc.)
│   │
│   ├── qualifier_age.go         ← Age compatibility check
│   ├── qualifier_dating_prefs.go← Gender preference check
│   ├── qualifier_height.go      ← Height preference check
│   ├── qualifier_distance.go    ← Distance radius check
│   │
│   └── store/
│       ├── match_result.go      ← Admin match queries
│       ├── user_match_actions.go← User match queries + updates
│       └── config.go            ← Config CRUD
│
└── extmatcher/
    └── qualitative_matcher.go   ← External AI scoring API
```

### Reading Order (Recommended)

1. **Start with models** - `lib/matching/models.go`
   - Understand the domain types first

2. **Read one qualifier** - `lib/matching/qualifier_age.go`
   - See how evaluation works

3. **Read store queries** - `lib/matching/store/user_match_actions.go`
   - See the SQL perspective transformation

4. **Read lib logic** - `lib/matching/logic.go`
   - ProposeMatch is the most interesting

5. **Read API handler** - `api/api_matching.go:userMatchingPropose()`
   - See how HTTP maps to lib calls

---

## Part 6: Common Tasks

### Task: Add a New Filter to Admin Endpoint

```go
// 1. Add constant (api/consts.go)
const QueryMatchResultNewFilter = "new_filter"

// 2. Add to model (lib/matching/models.go)
type MatchResultQueryFilter struct {
    NewFilter null.String  // ← Add here
    // ...
}

// 3. Add SQL condition (lib/matching/store/match_result.go)
func qModsMatchResult(filter *matching.MatchResultQueryFilter) []qm.QueryMod {
    if filter.NewFilter.Valid {
        mods = append(mods, qm.Where("new_column = ?", filter.NewFilter.String))
    }
    // ...
}

// 4. Parse in handler (api/api_matching.go)
filter.NewFilter = null.StringFromPtr(c.Query(QueryMatchResultNewFilter))

// 5. Add swagger annotation
// @Param new_filter query string false "Filter description"
```

### Task: Add a New Qualifier

```go
// 1. Create file: lib/matching/qualifier_example.go

type ExampleQualifier struct {
    config *Config
}

func (q *ExampleQualifier) Name() string { return "example" }
func (q *ExampleQualifier) IsHard() bool { return true }  // or false for soft

func (q *ExampleQualifier) Evaluate(userA, userB *User, config *Config) QualifierResult {
    passed := /* your logic */
    return QualifierResult{
        Passed: passed,
        Reason: "explanation",
        Data:   map[string]any{"detail": value},
    }
}

// 2. Register in lib/matching/matching.go constructor
func NewLib(...) *Lib {
    return &Lib{
        qualifiers: []Qualifier{
            &AgeQualifier{},
            &ExampleQualifier{},  // ← Add here
        },
    }
}
```

### Task: Add a New User Endpoint

```go
// 1. Add route (api/route_matching.go)
user.POST("/matches/:match_id/new-action", a.userMatchingNewAction)

// 2. Add handler (api/api_matching.go)
func (a *Mux) userMatchingNewAction(c *gin.Context) {
    userID := getUserID(c)
    matchID := uuid.MustParse(c.Param("match_id"))

    result, err := a.matching.NewAction(c.Request.Context(), userID, matchID)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, result)
}

// 3. Add business method (business/domain/matching/matching.go)
func (m *Matching) NewAction(ctx context.Context, userID, matchID uuid.UUID) (*Result, error) {
    return m.transactor.InTx(ctx, func(exec boil.ContextExecutor) (*Result, error) {
        return m.lib.NewAction(ctx, exec, userID, matchID)
    })
}

// 4. Add lib method (lib/matching/logic.go)
func (l *Lib) NewAction(ctx context.Context, exec boil.ContextExecutor, userID, matchID uuid.UUID) (*Result, error) {
    // Your logic here
}
```

---

## Part 7: Testing Patterns

### API Test Structure

```go
// api/api_matching_test.go

func TestUserMatchingPropose(t *testing.T) {
    for _, tt := range userMatchingProposeCases() {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            testSuite := testsuite.New(t)
            t.Cleanup(testSuite.UseBackendDB())

            // Setup
            th := testSuite.Helper()
            match := tt.setup(th)

            // Execute
            url := strings.Replace(api.PathUserMatchingPropose, ":match_id", match.ID.String(), 1)
            resp := th.AuthedRequest(http.MethodPost, url, nil, th.UserA.Token)

            // Assert
            require.Equal(t, tt.wantStatus, resp.Code)
            tt.extraAssertions(th, resp)
        })
    }
}

func userMatchingProposeCases() []testCasePropose {
    return []testCasePropose{
        {
            name: "success-propose-first",
            setup: func(th *testsuite.Helper) *pgmodel.MatchResult {
                // Create users, match, etc.
            },
            wantStatus: http.StatusOK,
            extraAssertions: func(th *testsuite.Helper, resp *httptest.ResponseRecorder) {
                var result api.ProposeResponse
                json.Unmarshal(resp.Body.Bytes(), &result)
                require.False(th.T, result.MutualProposal)

                // ALWAYS verify DB state
                match, _ := pgmodel.MatchResults(qm.Where("id = ?", matchID)).One(ctx, th.BackendAppDb())
                require.Equal(th.T, proposedActionID, match.UserAActionRefID)
            },
        },
        {
            name: "success-mutual-proposal",
            // ...
        },
        {
            name: "error-not-found",
            // ...
        },
    }
}
```

---

## Quick Reference

### Layer Responsibilities Cheat Sheet

| When you need to... | Go to... |
|---------------------|----------|
| Parse HTTP request | `api/api_matching.go` |
| Add route | `api/route_matching.go` |
| Add query param constant | `api/consts.go` |
| Wrap in transaction | `business/domain/matching/` |
| Add business rule | `lib/matching/logic.go` |
| Add qualifier | `lib/matching/qualifier_*.go` |
| Add/modify SQL query | `lib/matching/store/` |
| Add domain type | `lib/matching/models.go` |

### Debug Checklist

- [ ] Is the route registered? (`route_matching.go`)
- [ ] Is middleware correct? (admin vs user auth)
- [ ] Is TX boundary in business layer?
- [ ] Is lib accepting `exec boil.ContextExecutor`? (NOT transactor)
- [ ] Did you run `make wingedapp-swagger` after adding annotations?
