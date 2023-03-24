include .env

lint:
	golangci-lint run

tidy:
	go mod tidy

wire:
	wire ./internal/routes/di

ddos:
	ddosify -t http://localhost:3030/todo-items -n 10000

vegeta:
	echo "GET http://localhost:3030/todo-items" | vegeta attack -rate=10000 -duration=1s | vegeta report
