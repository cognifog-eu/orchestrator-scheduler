include .env

## Containerized with embeded database

build:	# Build application into container
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

template:
	helm template --name-template mysql ./job-manager-charts/mysql/ -n cognifog-dev \
	> jenkins/manifests.yaml
	helm template --name-template jobmanager job-manager-charts/job-manager/ -n cognifog-dev \
	>> jenkins/manifests.yaml
	echo "---" >> jenkins/manifests.yaml
	sed '/metadata:/a\  namespace: cognifog-dev' job-manager-charts/mysql/secret.yaml >> jenkins/manifests.yaml

start: # Start application from container with embeded database
	docker-compose up -d

stop: # Stop application from container with embeded database
	docker-compose down

start-database: # Start only database
	docker-compose -f docker-compose-database.yml up -d

stop-database: # Stop only database
	docker-compose -f docker-compose-database.yml down

## Local application
run:	# Start application
	go run main.go