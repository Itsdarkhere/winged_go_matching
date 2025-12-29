# PR: Refine CSV Ingestion Schema

## Summary

Refactors the population CSV ingestion schema to separate `sexuality` as a first-class column (matching the DB schema from migration 19) and renames `population_details` to `profile_details` for clarity. Adds comprehensive test coverage for the entire ingestion pipeline.

## Changes

### CSV Schema Breaking Change

**Before:**
```csv
email,age,gender,dating_preferences,height,latitude,longitude,population_details
"user@example.com","25","Male","Female","175","34.0522","-118.2437","{""sexuality_preferences"":""Heterosexual"",...}"
```

**After:**
```csv
email,age,gender,dating_preferences,height,latitude,longitude,sexuality,profile_details
"user@example.com","25","Male","Female","175","34.0522","-118.2437","Straight","{""sexuality_preferences"":""Heterosexual"",...}"
```

### Model Refactoring

- `PopulationDetails` → `ProfileDetails` (struct rename)
- `PopulationRow.Details` → `PopulationRow.ProfileDetails` (field rename)
- `HeaderPopulationDetails` → `HeaderProfileDetails` constant
- Added `HeaderSexuality` constant and `Sexuality` field to `PopulationRow`
- Added `Sexuality` field to `InsertPopulationUser`

### Sexuality Validation

Added validation for the new `sexuality` column with allowed values from migration 19:
- `Prefer not to say`
- `Straight`
- `Gay`
- `Lesbian`
- `Bisexual`
- `Asexual`
- `Questioning`
- `Other`

### New Test Coverage

Added 6 comprehensive ingestion tests (`internal/wingedapp/lib/matching/ingestion_test.go`):

| Test | Description |
|------|-------------|
| `TestIngestion_CSVParsingAllFields` | Verifies all CSV fields (including sexuality) are parsed correctly |
| `TestIngestion_UserFieldsPersisted` | Verifies user fields are stored in `backend_app.users` |
| `TestIngestion_DatingPreferencesPersisted` | Verifies dating preferences are stored correctly |
| `TestIngestion_ProfileDetailsPersisted` | Verifies qualitative/quantitative/categorical profile data in `ai_backend.profiles` |
| `TestIngestion_SupabaseUsersPersisted` | Verifies supabase auth users are created |
| `TestIngestion_MatchResultsCreated` | Verifies match results are created for all unique user pairs |

### Test Data Updates

Updated all CSV test files to use the new schema:
- `testdata/population_1.csv`
- `testdata/population_2.csv`
- `testdata/population_2_fail_all.csv`
- `testdata/population_3_pass_all.csv`
- `testdata/population_4_match_result_listing.csv`

### Docker Improvements

- Added `container_name: pg_tester_postgres` for easier identification
- Added `shared_buffers=256MB` for better performance
- Removed unused MySQL config and volumes
- Added `docker/init-postgres.sh` for PostgreSQL configuration

### API Test Fixes

Updated `api_matching_test.go`:
- All CSV test data uses new `sexuality,profile_details` columns
- Renamed test case `fail-invalid-json-population-details` → `fail-invalid-json-profile-details`
- Added sexuality assertions in parse tests

## Files Changed

```
internal/wingedapp/lib/matching/
├── ingestion_test.go          (+335 lines - new comprehensive tests)
├── logic.go                   (formatting)
├── models.go                  (formatting)
├── populate.go                (+Sexuality field, renamed types)
├── populator.go               (schema change, validation)
├── run_ingestion_set_test.go  (removed unnecessary tc := tc)
└── testdata/
    ├── population_1.csv       (schema update)
    ├── population_2.csv       (schema update)
    ├── population_2_fail_all.csv
    ├── population_3_pass_all.csv
    └── population_4_match_result_listing.csv

internal/wingedapp/api/
└── api_matching_test.go       (CSV schema updates)

docker/
├── docker-compose.yml         (postgres improvements)
└── init-postgres.sh           (new - pg config script)

scripts/
└── docker-start.sh            (updated)
```

## Migration Notes

Any existing CSV files used for ingestion must be updated to the new schema:
1. Add `sexuality` column after `longitude`
2. Rename `population_details` header to `profile_details`
3. Valid sexuality values: `Straight`, `Gay`, `Lesbian`, `Bisexual`, `Asexual`, `Questioning`, `Prefer not to say`, `Other`

## Test Plan

- [x] All existing tests pass
- [x] New ingestion tests pass
- [x] CSV parsing correctly extracts sexuality field
- [x] Invalid sexuality values are rejected
- [x] Profile details are persisted to ai_backend
- [x] Match results are created for all user pairs
