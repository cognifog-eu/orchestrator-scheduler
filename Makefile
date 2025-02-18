include .env

## Containerized with embeded database

build-container:	# Build application into container
	docker build --pull --rm -f "Dockerfile" -t harbor.cognifog.rid-intrasoft.eu/orchestrator-scheduler/orch-scheduler-jobmanager:dev "."

push: # Push container to repository
	docker push harbor.cognifog.rid-intrasoft.eu/orchestrator-scheduler/orch-scheduler-jobmanager:dev

install:
	kubectl get namespace jobmanager || kubectl create namespace jobmanager
	kubectl apply -f job-manager-charts/mysql/secret.yaml -n jobmanager
	helm upgrade -i mysql ./job-manager-charts/mysql/ -n jobmanager
	helm upgrade -i jobmanager job-manager-charts/job-manager/ -n jobmanager

uninstall:
	helm uninstall jobmanager -n jobmanager
	helm uninstall mysql -n jobmanager

start: # Start application from container with embeded database
	docker-compose up -d

stop: # Stop application from container with embeded database
	docker-compose down

start-database: # Start only database
	docker-compose -f docker-compose-database.yml up -d

stop-database: # Stop only database
	docker-compose -f docker-compose-database.yml down

## Local application

build:	# Build application
	go build -o bin/main main.go

run:	# Start application
	go run main.go