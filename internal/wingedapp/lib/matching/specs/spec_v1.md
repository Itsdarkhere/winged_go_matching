#### Matching logic spec

- Spec reference: https://docs.google.com/document/d/1qL-O-RcApb8QFwi4-TJYzv68mWl9iOb83t0KZeSDwTA/edit?tab=t.0

- must have different versions of:
  - matching logic
  - ingestion logic
- I cannot put visualise how it goes right now, but it must be revisited.

#### Nov 22, 2025
Making an /v1 api that can support basic admin tasks such as searching, and approving match results.

- Admin API
  - ` [GET] /matching/config` (prune)
    - description: display all config parameters here (prune lol - can live without for now)
  - ` [PATCH] /matching/config` (prune)
    - full logically guarded update-able settings
  - ` [PATCH] /matching/results`
    - approves, unapproved
  - ` [GET] /matching/results` (also reused in user context)
    - description: returns admin relevant matching results 
    - search criteria:
      - match results category (not user life cycle)
        - fields: `approved` / `unapproved` / `pending`
        - rationale: see all _approved_, _unapproved_, _pending review_ matches
    - fields:
      - id
      - user_a_details (admin enriched - skip when reused by regular user)
      - user_b_details (admin enriched - skip when reused by regular user)
      - is_approved
      - user_lifecycle_status (extra eye-candy) 
      - match_score
      - matchmaking_response (from Havard's service)
      - ideal_date_audio (currently not here)
      - user_fields:
        - lat
        - long 
        - raw_address (Helsinki, 00200, Mannerheimintie 1) 
    - implementation details:
      - public function prunes non-admin views
      - ensure this also works with pagination etc
      - need to learn that serialiser pattern lol (as early as now — learn display hygiene)
    - enrichable fields:
      - settings used (shows the settings for this current match)