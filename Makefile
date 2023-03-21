include .env

lint:
	golangci-lint run

tidy:
	go mod tidy

wire:
	wire ./internal/routes/di

migrate:
	migrate -database 'mysql://${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DBNAME}?parseTime=true' -path internal/migrations up