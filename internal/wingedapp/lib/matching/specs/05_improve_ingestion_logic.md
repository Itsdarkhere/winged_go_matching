#### Tech Specs:

We need to improve ingestion logic:
- it should return how many records were ingested
- it should return how many records had errors
- it should (very important) - clean any conflict records that already exist in: (backend_app, supabase_app, and profiles)
- by clear delete those conflicting entries.

... do those and add thorough tests, but only in the API layer.