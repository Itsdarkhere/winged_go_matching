#### Tech Specs: improve matching format to be the same as what Niina wants

The ingestion format MUST match the `test group v1_rows.csv`

## Solution - COMPLETED

### Summary
Updated the CSV ingestion format from the old JSON-based `profile_details` format to Niina's flat column format ("test group v1"). The new format has all profile fields as separate columns and uses `birthday` instead of `age`.

### Key Format Changes

| Old Format | New Format |
|------------|------------|
| `email` | Not required (generated from first_name + last_name) |
| `firstname` | `first_name` |
| `lastname` | `last_name` |
| `age` (integer) | `birthday` (YYYY-MM-DD, age calculated) |
| `height` | `height_cm` |
| `sexuality` | `sexuality_preferences` (in profile) |
| `dating_preferences` (comma-separated) | `dating_preference` (singular: Male/Female/Any) |
| `profile_details` (JSON blob) | 33 flat columns |
| Quantitative values 1-10 | Quantitative values 0-1 floats |

### New Column Headers
```
first_name,last_name,address,birthday,gender,height_cm,dating_pref_age_range_start,dating_pref_age_range_end,latitude,longitude,self_portrait,interests,wellbeing_habits,self_care_habits,money_management,self_reflection_capabilities,moral_frameworks,life_goals,partnership_values,mutual_commitment,spirituality_growth_mindset,cultural_values,family_planning,ideal_date,red_green_flags,extroversion_social_energy,routine_vs_spontaneity,agreeableness,conscientiousness,neuroticism,dominance_level,emotional_expressiveness,sex_drive,geographical_mobility,conflict_resolution_style,sexuality_preferences,religion,dating_preference
```

### Files Changed

| File | What Changed |
|------|--------------|
| `lib/matching/populator.go` | New headers, `parseRow()` for flat columns, `calculateAge()`, `expandDatingPreference()`, updated validation |
| `lib/matching/populate.go` | `InsertPopulationUser` with `Birthday` field, `generateEmail()` for auto-generation, removed `CalculateBirthdayFromAge` |
| `lib/matching/store/user.go` | Uses `Birthday` directly instead of calculating from age |
| `lib/matching/testdata/*.csv` | All 4 test CSVs updated to new format |
| `lib/matching/ingestion_test.go` | Updated assertions for new field names |

### Email Generation
Since the new format doesn't include email, we generate unique emails:
```
Format: firstname.lastname.shortUUID@winged-test.local
Example: demby.abella.a1b2c3d4@winged-test.local
```

### Dating Preference Expansion
```
"Male" -> ["Male"]
"Female" -> ["Female"]
"Any" -> ["Male", "Female", "Non-Binary"]
```

### Post-Deploy Steps
- No database migrations required
- API endpoint accepts new CSV format immediately
- Old format CSVs will fail validation (missing required headers)
