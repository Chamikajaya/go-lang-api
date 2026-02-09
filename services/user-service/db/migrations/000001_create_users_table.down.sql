-- Drop indexes first
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_email;

-- Drop the users table
DROP TABLE IF EXISTS users;

-- Drop the enum type
DROP TYPE IF EXISTS user_status;

-- Drop the UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp";
