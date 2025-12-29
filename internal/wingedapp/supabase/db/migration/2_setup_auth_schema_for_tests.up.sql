-- Setup auth schema for test environments
-- In production Supabase, this schema already exists
CREATE SCHEMA IF NOT EXISTS auth;

-- Create minimal auth.users table for tests
-- In production Supabase, a more complete version exists
CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

-- Create auth.uid() function stub for tests
-- In production Supabase, this returns the actual authenticated user ID
CREATE OR REPLACE FUNCTION auth.uid()
RETURNS UUID AS $$
BEGIN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
