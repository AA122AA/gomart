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

-- name: UpdateOrderStatus :exec
UPDATE orders
SET  status = $1, accrual = $2
WHERE oid = $3;

-- Bonus Accounts --
-- name: CreateBonusAccount :exec
INSERT INTO bonus_accounts (user_id, current_balance, total_bonuses_spent)
VALUES (
    (SELECT id FROM users WHERE username = $1 LIMIT 1),
    $2,
    $3
);

-- name: GetBalanceByUserName :one
SELECT current_balance, total_bonuses_spent
FROM bonus_accounts
JOIN users ON users.id = bonus_accounts.user_id
WHERE users.username = $1;

-- name: UpdateBalance :exec
UPDATE bonus_accounts SET current_balance = $1, total_bonuses_spent = $2
WHERE user_id = (SELECT id FROM users WHERE username = $3 LIMIT 1);

-- Bonus Transactions --
-- name: InsertBonusTransaction :exec
INSERT INTO bonus_transactions (bonus_id, oid, bonuses_withdraw, processed_at)
SELECT ba.id, $2, $3, $4
FROM users u
JOIN bonus_accounts ba ON ba.user_id = u.id
WHERE u.username = $1
LIMIT 1;

-- name: GetBonusTransactionsByUserName :many
SELECT oid, bonuses_withdraw, processed_at
FROM bonus_transactions bt
JOIN bonus_accounts ba ON bt.bonus_id = ba.id
JOIN users u ON ba.user_id = u.id
WHERE u.username = $1
ORDER BY bt.processed_at DESC;

-- Accrual --

-- name: GetAllOrdersForAccrual :many
SELECT users.username, orders.oid, orders.status, orders.accrual, orders.uploaded_at
FROM orders
JOIN users ON orders.user_id = users.id
WHERE orders.status NOT IN ('PROCESSED', 'INVALID');

-- name: GetBalanceByOrderID :one
SELECT u.username, ba.current_balance, ba.total_bonuses_spent
FROM bonus_accounts ba
JOIN orders o ON o.user_id = ba.user_id
JOIN users u ON o.user_id = u.id
WHERE o.oid = $1;
