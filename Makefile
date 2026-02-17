## >>> Environments <<<

### Server
SERVER_PORT ?= 8082

### DB
DB_USER ?= gomart
DB_PASSWORD ?= StrongPass123!
DB_NAME ?= gomart

DSN ?= "postgresql://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable"

####################################################################################################

## >>> Developer <<<

### golang
.PHONY: build
build:
	go build -o ./cmd/gophermart/gomart ./cmd/gophermart/main.go

.PHONY: run
run: build
	./cmd/gophermart/gomart -a "localhost:${SERVER_PORT}" -d ${DSN}

tidy:
	go mod tidy

### Migrations
goose-create-init:
	goose postgres ${DSN} -s -dir ./db/schema create init sql

goose-status:
	goose postgres ${DSN} -dir ./db/schema status

goose-up:
	goose postgres ${DSN} -dir ./db/schema up

####################################################################################################

## >>> API Test <<<
register:
	curl -v -X POST -d '{"login":"gromartem", "password":"qwerty123"}' http://localhost:${SERVER_PORT}/api/user/register | jq "."

login:
	curl -v -X POST -d '{"login":"gromartem", "password":"qwerty123"}' http://localhost:${SERVER_PORT}/api/user/login | jq "."

TOKEN ?=
logout:
	curl -v -X POST -H 'Authorization: Bearer ${TOKEN}' http://localhost:${SERVER_PORT}/api/user/logout

RTOKEN ?=
refresh:
	curl -v -X POST -H 'Authorization: Bearer ${TOKEN}' -d '{"refresh_token":"${RTOKEN}"}' http://localhost:${SERVER_PORT}/api/user/refresh

create-order:
	curl -v -X POST http://localhost:${SERVER_PORT}/api/user/orders

get-order:
	curl -v -X GET http://localhost:${SERVER_PORT}/api/user/orders

get-balance:
	curl -v -X GET http://localhost:${SERVER_PORT}/api/user/balance

withdraw:
	curl -v -X GET http://localhost:${SERVER_PORT}/api/user/balance/withdraw

history:
	curl -v -X GET http://localhost:${SERVER_PORT}/api/user/withdrawals

####################################################################################################

## >>> Database <<<
select-all-users:
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "SELECT * from users ORDER BY id;"

select-all-orders:
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "SELECT * from orders ORDER BY id;"

select-all-bonus-accounts:
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "SELECT * from bonus_accounts ORDER BY id;"

select-all-bonus-transactions:
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "SELECT * from bonus_transactions ORDER BY id;"

select-all-user-tokens:
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "SELECT * from user_tokens ORDER BY id;"

delete-all:
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE users RESTART IDENTITY;"
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE orders RESTART IDENTITY;"
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE bonus_accounts RESTART IDENTITY;"
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE bonus_transactions RESTART IDENTITY;"
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE user_tokens RESTART IDENTITY;"

####################################################################################################

## >>> DevOps <<<

pg-up:
	docker compose -f ./dockers/docker-compose.yaml up -d

pg-down:
	docker compose -f ./dockers/docker-compose.yaml down
