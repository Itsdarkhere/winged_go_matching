### Spend — messages

---

#### Description
This is a wings economy table for spending wings on messages.

--- 

#### Tables
- wing_ecn_spend_message
- wing_ecn_user_totals
- wing_ecn_transaction

--- 

#### Logic

- Legend:
  - `ext` = external signal, prefix, not part of original table name

---
- Flow
  - `ext_user_ai_convo` -> `syncs` ->
  - `wing_ecn_spend_message` -> `increments` ->
  - `wing_ecn_user_totals` -> `possible sync counter` ->
  - `wing_ecn_transaction` -> `possible insert` ->
  - `wing_ecn_user_totals` -> `possible wings deduction`

--- 

#### Notes

---

- **Important!** This is your first attempt essentially making it the architecture of wings economy. Expected revisions, mistakes, etc...
- Scalability:
  - new totals/increments? Alter user_totals, add new column(s)
  - new billing table? Same — alter user_totals, add new column(s)
  - ... seems like we're all good here

--- 

#### Steps (initial guidance)

---

- update config table, add meta JSON column
- create `wing_ecn_spend_transaction` table (heavy design consideration -> have to make it general enough to be mapped across all wing econ tables)
- create `wing_ecn_user_total` table (heavy design consideration -> holds all totals including increments (increments could be diff table, but just compressed)