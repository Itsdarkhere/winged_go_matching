-- Teardown auth schema test setup
DROP FUNCTION IF EXISTS auth.uid();
DROP TABLE IF EXISTS auth.users;
DROP SCHEMA IF EXISTS auth CASCADE;
