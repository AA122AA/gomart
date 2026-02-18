-- USERS --
-- name: GetUser :one
SELECT id, username, password_hash, created_at, updated_at FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUsers :many
SELECT id, username, password_hash, created_at, updated_at FROM users;

-- name: GetUserPassword :one
SELECT password_hash FROM users
WHERE username = $1 LIMIT 1;

-- name: InsertUser :exec
INSERT INTO users (username, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4);


-- TOKEN --
-- name: GetRefreshTokenByUserName :one
SELECT user_tokens.token_id, user_tokens.is_valid
FROM user_tokens
JOIN users ON users.id = user_tokens.user_id
WHERE users.username = $1
LIMIT 1;

-- name: InsertRefreshToken :exec
INSERT INTO user_tokens (user_id, token_id, is_valid)
VALUES (
    (SELECT id FROM users WHERE username = $1 LIMIT 1),
    $2,
    $3
);

-- name: UpdateRefreshToken :exec
UPDATE user_tokens
SET token_id = $1, is_valid = $2
WHERE user_id = (SELECT id FROM users WHERE username = $3 LIMIT 1);

-- name: UpdateRefreshTokenIsValid :exec
UPDATE user_tokens SET is_valid = $1
WHERE user_id = (SELECT id FROM users WHERE username = $2 LIMIT 1);

-- ORDERS --
-- name: GetAll :many
SELECT users.username, orders.oid, orders.status, orders.accrual, orders.uploaded_at
FROM orders
JOIN users ON orders.user_id = users.id;

-- name: GetOrdersByUserName :many
SELECT OID, status, accrual, uploaded_at
FROM orders
JOIN users ON users.id = orders.user_id
WHERE users.username = $1;

-- name: GetOrderByOID :one
SELECT users.username, orders.oid, orders.status, orders.accrual, orders.uploaded_at
FROM orders
JOIN users ON orders.user_id = users.id
WHERE orders.oid = $1
LIMIT 1;

-- name: CreateOrder :exec
INSERT INTO orders (oid, user_id, status, uploaded_at, updated_at)
VALUES (
    $1,
    (SELECT id FROM users WHERE username = $2 LIMIT 1),
    $3,
    $4,
    $5
);
