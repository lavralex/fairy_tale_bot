include .env

.PHONY: build vet run db-up db-down migrate-up migrate-down

build:
	go build ./... 

vet:
	go vet ./...

run:
	go run ./cmd/bot/

db-up:
	docker-compose up -d

db-down:
	docker-compose down -v

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down
