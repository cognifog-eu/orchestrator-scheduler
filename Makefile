include .env

start-database:
	docker-compose up -d

stop-database:
	docker-compose down

build:
	go build -o bin/main main.go

run:
	go run main.go