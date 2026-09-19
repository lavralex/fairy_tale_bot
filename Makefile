include .env

.PHONY: build vet lint run db-up db-down migrate-up migrate-down f test

build:
	go build ./... 

vet:
	go vet ./...

# требует golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
lint:
	golangci-lint run ./...

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

f:
	gofmt -l -w .

test:
	go test ./...