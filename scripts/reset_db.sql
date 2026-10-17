-- Reset database schema. For development and testing environments only.
-- WARNING: Do not execute against databases containing production data.
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO public;
