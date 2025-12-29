#### Tech Specs: Admin Match Set Search

Full search/filter support for admin match sets, following the same patterns as match result search.

##### Endpoint
- `GET /admin/users/matching/sets` - Admin-only endpoint to query match sets with pagination and filters

##### Query Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | string | Filter by match set name (partial match) |
| `min_participants` | int | Minimum number of participants |
| `max_participants` | int | Maximum number of participants |
| `time_start_after` | RFC3339 | Time start after this timestamp |
| `time_start_before` | RFC3339 | Time start before this timestamp |
| `time_end_after` | RFC3339 | Time end after this timestamp |
| `time_end_before` | RFC3339 | Time end before this timestamp |
| `created_after` | RFC3339 | Created after this timestamp |
| `created_before` | RFC3339 | Created before this timestamp |
| `order_by` | string | Column to sort by (id, name, number_of_participants, time_start, time_end, created_at) |
| `sort` | string | Sort direction: `+` for ASC, `-` for DESC |
| `page` | int | Page number (1-indexed) |
| `per_page` | int | Items per page |

##### Response
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "string",
      "number_of_participants": 10,
      "match_configuration": {...},
      "time_start": "2025-01-01T00:00:00Z",
      "time_end": "2025-01-02T00:00:00Z",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

##### Implementation Layers
1. **Models** (`lib/matching/models.go`): `MatchSet`, `MatchSetPaginated`, `QueryFilterMatchSet`
2. **Store** (`lib/matching/store/match_set.go`): `MatchSets()`, `MatchSet()`, `qModsMatchSet()`
3. **Logic** (`lib/matching/queries.go`): `MatchSets()`, `MatchSet()`
4. **Business** (`business/domain/matching/matching_admin.go`): `AdminMatchSets()`
5. **API** (`api/api_matching.go`): `adminMatchSets` handler

##### Security
- Admin-only access via `mwAdmin` middleware
