### Spend — messages 

---

#### Description
This is a wings economy table for spending wings on messages, 
1st iteration,  and much simplier.

**Simplified changes:**
- reduced the amount of tables to be just 3
- simplified exposed apis to be: 
  - `AddEntry` with categories: `AddMessage`, `AddInvite`, `AddDaily`
- simple proxy database field:
  - id
  - refID
  - refTypeID
  - JSON for extensibility
- decided where it lives (`internal/wingedapp/economy`)

--- 

#### Tables
- `wing_ecn_action`
- `wing_ecn_user_totals`
- `wing_ecn_transaction`

--- 

#### Logic

- Legend:
  - `ext` = external signal, prefix, not part of original table name

---
- Flow
  - `ext_user_ai_convo` -> `syncs` ->
  - `wing_ecn_action` -> `increments` ->
  - `wing_ecn_user_totals` -> `possible sync counter` ->
  - `wing_ecn_transaction` -> `possible insert` ->
  - `wing_ecn_user_totals` -> `possible wings deduction`

--- 

#### Notes

---

- **Important!** This is an iteration of `01_spend_messages.md`

--- 

#### Steps (initial guidance)

---

- update config table, add meta JSON column
- create `wing_ecn_spend_transaction` table (heavy design consideration -> have to make it general enough to be mapped across all wing econ tables)
- create `wing_ecn_user_total` table (heavy design consideration -> holds all totals including increments (increments could be diff table, but just compressed)