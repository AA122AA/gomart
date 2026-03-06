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
RTOKEN ?=
TOKEN ?=

register:
	curl -v -X POST -d '{"login":"gromartem", "password":"qwerty123"}' http://localhost:${SERVER_PORT}/api/user/register | jq "."

login:
	curl -v -X POST -d '{"login":"gromartem", "password":"qwerty123"}' http://localhost:${SERVER_PORT}/api/user/login | jq "."

logout:
	curl -v -X POST -H 'Authorization: Bearer ${TOKEN}' http://localhost:${SERVER_PORT}/api/user/logout

refresh:
	curl -v -X POST -H 'Authorization: Bearer ${TOKEN}' -d '{"refresh_token":"${RTOKEN}"}' http://localhost:${SERVER_PORT}/api/user/refresh

create-order:
	# curl -v -X POST -H 'Authorization: Bearer ${TOKEN}' -d '9278923470' http://localhost:${SERVER_PORT}/api/user/orders
	curl -v -X POST -H 'Authorization: Bearer ${TOKEN}' -d '12345678903' http://localhost:${SERVER_PORT}/api/user/orders

get-order:
	curl -v -X GET -H 'Authorization: Bearer ${TOKEN}' http://localhost:${SERVER_PORT}/api/user/orders | jq "."

get-balance:
	curl -v -X GET -H 'Authorization: Bearer ${TOKEN}' http://localhost:${SERVER_PORT}/api/user/balance | jq "."

withdraw:
	curl -v -X POST -H 'Authorization: Bearer ${TOKEN}' -d '{"order": "12345678903","sum": 80}' http://localhost:${SERVER_PORT}/api/user/balance/withdraw

history:
	curl -v -X GET -H 'Authorization: Bearer ${TOKEN}' http://localhost:${SERVER_PORT}/api/user/withdrawals | jq "."

create-accural-order:
	# curl -v -X POST -H "Content-Type: application/json" -d '{"order":"927923470", "goods":[{"descripiton":"BMW 3","price":4788399.99}]}' http://localhost:8080/api/orders
	curl -v -X POST -H "Content-Type: application/json" -d '{"order":"12345678903", "goods":[{"descripiton":"BMW 3","price":4788399.99}]}' http://localhost:8080/api/orders

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
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE users RESTART IDENTITY CASCADE;"
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE orders RESTART IDENTITY CASCADE;"
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE bonus_accounts RESTART IDENTITY CASCADE;"
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE bonus_transactions RESTART IDENTITY CASCADE;"
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE user_tokens RESTART IDENTITY CASCADE;"

####################################################################################################

## >>> DevOps <<<

pg-up:
	docker compose -f ./dockers/docker-compose.yaml up -d

pg-down:
	docker compose -f ./dockers/docker-compose.yaml down

####################################################################################################

## >>> Autotests <<<

static-test:
	go vet -vettool=./statictest ./...

autotests: delete-all build static-test
	./gophermarttest \
	    -test.v -test.run=^TestGophermart$ \
		-gophermart-binary-path=cmd/gophermart/gomart \
        -gophermart-host=localhost \
        -gophermart-port=8084 \
        -gophermart-database-uri=${DSN} \
        -accrual-binary-path=cmd/accrual/accrual_darwin_arm64 \
        -accrual-host=localhost \
        -accrual-port=8085 \
        -accrual-database-uri=${DSN}
        # -accrual-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable"
