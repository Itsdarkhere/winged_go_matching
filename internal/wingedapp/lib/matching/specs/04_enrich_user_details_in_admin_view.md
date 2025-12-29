# PR: Admin Matching View - User Details Enrichment

## Summary

Enhanced the Admin Matching View API to provide **complete user details** for each match result. Admins can now see full profiles for both users in a match without navigating to separate screens.

---

## Changes

### API Response Enhancements

When calling `GET /admin/matching/views?enrich_users=1`, the response now includes:

| New Field | What It Contains |
|-----------|-----------------|
| `user_a_details` | Age, gender, height, location, dating preferences |
| `user_b_details` | Age, gender, height, location, dating preferences |
| `user_a_profile` | Personality traits, interests, values, lifestyle |
| `user_b_profile` | Personality traits, interests, values, lifestyle |

### Example Response

```json
{
  "user_a_details": {
    "id": "user-uuid",
    "age": 25,
    "gender": "male",
    "height": 180,
    "latitude": 60.1699,
    "longitude": 24.9384,
    "dating_preferences": [...]
  },
  "user_b_details": {
    "id": "user-uuid",
    "age": 27,
    "gender": "female",
    "height": 165,
    "latitude": 60.1699,
    "longitude": 24.9384,
    "dating_preferences": [...]
  },
  "user_a_profile": {
    "Qualitative": { "Interests": "...", "Life Goals": "...", ... },
    "Quantitative": { "Extroversion": 7, "Agreeableness": 8, ... },
    "Categorical": { "Religion": "...", "Conflict Resolution Style": "..." }
  },
  "user_b_profile": { ... }
}
```

---

## Benefits for Product/Ops

- **Full visibility** into both users' matching parameters side-by-side
- **Better decision making** when manually approving/rejecting matches
- **Quality assurance** - verify algorithm decisions with complete context
- **Graceful handling** - missing profiles show `null` instead of errors

---

## Architecture Improvements

Refactored monolithic `PopulateStore` into atomic stores:

| Before | After |
|--------|-------|
| `PopulateStore` | `UserStore` |
| (one big store) | `UserDatingPrefsStore` |
| | `ProfileStore` |
| | `SupabaseUserStore` |

Each store now maps 1-1 to a database table for better maintainability.

---

## Test Coverage

- `TestAdminMatching_Views/success-enrich-users` - All fields populated correctly
- `TestAdminMatching_Views/success-enrich-users-fails-gracefully` - Graceful null handling
- `TestLogic_MatchResults_Success/*` - Comprehensive lib layer tests
