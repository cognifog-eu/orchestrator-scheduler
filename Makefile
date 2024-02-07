include .env

## Containerized with embeded database

build-container:	# Build application into container
	docker build --pull --rm -f "Dockerfile" -t registry.atosresearch.eu:18509/orch-scheduler-jobmanager "."

start-database:	# Start application from container with embeded database
	docker-compose up -d

stop-database:	# Stop application from container with embeded database
	docker-compose down

## Local application

build:	# Build application
	go build -o bin/main main.go

run:	# Start application
	go run main.go