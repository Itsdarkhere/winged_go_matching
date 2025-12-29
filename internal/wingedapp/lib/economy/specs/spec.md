### Spec

- Features
- tables:
    - subscription_plan (buys a plan — purchasing)
        - cols:
            - cadence (cadence_ref_id to category)
              - (category_type) (cadence)
              - (category)      (one-time, weekly, monthly, yearly, daily, etc)
            - name 
              - (category_type) (subscripion_plan)
              - (category) (fff)
            - price (10 EU)
            - winged equivalents (20 wings)
        - UNIQUE_COMPOSITE(cadence, name) (wingedx, weeklyy)
                                          (winged,  weeklyy)
        - 
  - user_subscription_plan (buys a plan — purchasing)
    - cols:
      - user_id (FK to subscription_plan)
      - start_date
      - end_date
      - is_active (bool)
  - user_wings_balance (totals)
    - cols:
        - user_id (FK to subscription_plan)
        - amount
        - last_updated_at
  - user_wings_transaction (granular transactions / aka ledger)
    - notes:
      - these can result from actions or purchases
        - going on a date
        - referring a friend etc etc
        - this is where you can receive a PAyMENT (So we need extra JSON colmn lol)
        - can be positive or negative (like a real trans table)
        - I think it can joined to different cols
          - we CAN branch off of this table, and extend to more details IF we want to like (dates, visits, etc etc)
  # Should we do it like this ? to associate to user_winges_transaction ? 
  # approach A - should it be diff tables?
  - user_winged_action_daily_check_in
  - user_winged_action_refer_action
  - user_winged_action_attend_date
  - user_winged_action_user_invite_completion
  # approach B - should be one table? (ok maam leaning towards here)
    - uuid
    - action_ref_id FK to category (daily_check_in, refer_action, attend_date, user_invite_completion)
                    FK to category_type (action_rewards)
    - additional_data? (JSONB)
  - user_earn_wing
  - spends 

#### Technical decisions RANT lol (not-finalised):
- concern
  - engaging with this domain means you are dealing with money
- database
  - extend repo, merely add those tables above
- business:
  - this would fall under the `economy` domain, what about if its used in other domains? 
  - whats the simpliest way for this to be coded to be domain coherent, and injectable to other modules?
    - well I tg
- how is this going to be tied up?
  - domain is going to house all those API ingestion logic...
    - how about those other logic? like  hmmm