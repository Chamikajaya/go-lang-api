/*
name: <MethodName> :<command>

:one: Expects exactly one row (returns a struct and an error).
:many: Expects zero or more rows (returns a slice).
:exec: Executes a query without returning rows (returns only an error).
:execrows: Returns the number of affected rows.

*/

-- name: CreateUser :one
-- Creates a new user and returns the created user
INSERT INTO users (
    first_name,
    last_name,
    email,
    phone,
    age,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetUserByID :one
-- Retrieves a single user by their ID
SELECT * FROM users
WHERE user_id = $1;

-- name: GetUserByEmail :one
-- Retrieves a single user by their email
SELECT * FROM users
WHERE email = $1;

-- name: ListUsers :many
-- Retrieves all users with optional filtering
SELECT * FROM users
ORDER BY created_at DESC;

-- name: ListUsersByStatus :many
-- Retrieves users filtered by status
SELECT * FROM users
WHERE status = $1
ORDER BY created_at DESC;

-- name: UpdateUser :one
-- Updates a user's information
UPDATE users
SET
    first_name = COALESCE($2, first_name),
    last_name = COALESCE($3, last_name),
    email = COALESCE($4, email),
    phone = COALESCE($5, phone),
    age = COALESCE($6, age),
    status = COALESCE($7, status),
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
RETURNING *;

-- name: DeleteUser :exec
-- Deletes a user by ID
DELETE FROM users
WHERE user_id = $1;

-- name: UserExists :one
-- Checks if a user exists by ID
SELECT EXISTS(
    SELECT 1 FROM users WHERE user_id = $1
);

-- name: EmailExists :one
-- Checks if an email is already registered
SELECT EXISTS(
    SELECT 1 FROM users WHERE email = $1
);