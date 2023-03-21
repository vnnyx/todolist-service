include .env

lint:
	golangci-lint run

tidy:
	go mod tidy

wire:
	wire ./internal/routes/di