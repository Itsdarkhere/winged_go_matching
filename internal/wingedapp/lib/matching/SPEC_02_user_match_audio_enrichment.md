# SPEC_02: User Match Audio Enrichment & Batch Seen Updates

## Overview

This spec documents the implementation of:
1. Batch update for marking matches as "seen" by users
2. Audio intro URL enrichment for user matches from `ai_backend`

## Changes

### 1. Batch Seen Update

**Problem**: The previous implementation updated matches one-by-one in a loop, which is inefficient for marking multiple matches as seen.

**Solution**: Added `UpdateSeenBatch` method to the store that uses a single SQL query per user position (A or B).

**Files Changed**:
- `lib/matching/store/user_match_actions.go` - Added `UpdateSeenBatch` method
- `lib/matching/adapters.go` - Added method to `userMatchActionsStorer` interface
- `lib/matching/logic.go` - Updated `MarkMatchesSeen` to use batch method

**SQL Query Pattern**:
```sql
UPDATE match_result
SET user_a_seen_at = $now
WHERE id IN ($matchIDs...)
  AND user_a_ref_id = $userID
  AND user_a_seen_at IS NULL
```
(Separate query for user_b_seen_at)

### 2. Audio Intro URL Enrichment

**Problem**: User matches need to display the intro audio URL for both the current user and their partner. Audio files are stored in `ai_backend.audio_files` table, keyed by `supabase_id` (not backend_app user ID).

**Solution**:
1. Added `audioGetter` interface and `AudioStore` implementation
2. Extended `UserMatches` query to include `supabase_id` for both users
3. Added enrichment logic in `Logic.UserMatches` and `Logic.UserMatch`

**Files Changed**:
- `lib/matching/store/audio.go` - New file with `AudioStore`
- `lib/matching/store/store.go` - Added `AudioStore` to `MatchingStores`
- `lib/matching/adapters.go` - Added `audioGetter` interface
- `lib/matching/logic.go` - Added `audioGetter` field, `enrichMatchAudio` method
- `lib/matching/models.go` - Added `YourSupabaseID`, `PartnerSupabaseID`, `YourIntroURL`, `PartnerIntroURL` fields to `UserMatch`
- `di/deps_lib.go` - Wired up `AudioStore` to `Logic`

**Query Flow**:
1. `UserMatches` query includes `your_supabase_id` and `partner_supabase_id` via CASE expressions
2. If `filter.EnrichAudio = true` and `audioGetter` is available:
   - Collect unique supabase IDs from results
   - Batch fetch audio URLs via `IntroAudioURLs`
   - Apply URLs to matches

**Audio Store Query**:
```go
aipgmodel.AudioFiles(
    aipgmodel.AudioFileWhere.UserID.IN(userIDStrings),
    aipgmodel.AudioFileWhere.Category.EQ(null.StringFrom("intro")),
    aipgmodel.AudioFileWhere.StorageURL.IsNotNull(),
).All(ctx, exec)
```

### 3. Model Updates

**QueryFilterUserMatch**:
- Added `EnrichAudio bool` flag

**UserMatch**:
- Added `YourSupabaseID null.String` (internal, `json:"-"`)
- Added `PartnerSupabaseID null.String` (internal, `json:"-"`)
- Added `YourIntroURL null.String` (enriched)
- Added `PartnerIntroURL null.String` (enriched)

## Interface Conventions

Per CLAUDE.md standards:
- Interfaces are lowercase and unexported
- `audioGetter` interface has only 2 methods (CRUD-like)
- Store methods use SQLBoiler `Where` helpers with pgmodel constants

## Testing Notes

- Tests pass with existing test suites
- Audio enrichment is optional - if `audioGetter` is nil or audio not found, returns `null.String{}`
- Batch seen update handles both user_a and user_b positions in separate queries

## Future Considerations

- Consider adding index on `audio_files(user_id, category)` for performance
- May want to cache audio URLs if frequently accessed
- Could add `EnrichAudio` to API query params if needed
