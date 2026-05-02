include .env
export

GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

run-checker:
	go run cmd/checker/main.go

run-manager:
	go run cmd/manager/main.go

goose-up:
	goose -dir ./migrations postgres "$(GOOSE_DBSTRING)" up

goose-down:
	goose -dir ./migrations postrgres "$(GOOSE_DBSTRING)" down