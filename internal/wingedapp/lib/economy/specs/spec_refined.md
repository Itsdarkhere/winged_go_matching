### Answers

- **Table arrangement**: <br/>
    - entrypoint (action logger) &nbsp; >
    - custom logic  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;> 
    - aggregate (optional) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;> 
    - transactions &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;> 
    - totals

- **Custom logic**: <br/>
  - Payment
    - pull payment from subscription
    - injest it.. 
    - shape the logic based on the data

- **Claims**: <br/>
    - you can claim wings earned w/configurable limits
    - unclaimed wings will not accrue until claimed
    - you can associate claims in the tx table via an FK
    - tx will have multiple fields (derived because of this):
      - tx_type (daily_check_in)
      - claimed (bool) (query non-claimed wings) (category_type for both entrypoint tables, and tx tables needed)
      - 
  
- **Void**: <br/>
  - follow your prev xp with dental systems, mark everything as void, and reverse the totals


- **Refinements TX structure + associated tables **: <br/>
  - transaction_type_id lives for 2 tables 
    - tx_table (winged_earn_referral)
    - action_tbl (winged_earn_referral)
  - this is how you can associate the claims
