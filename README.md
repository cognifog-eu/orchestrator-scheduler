# COGNIFOG Orchestrator Scheduler

COGNIFOG Orchestrator Scheduler Microservice: In charge of Creating and Managing Applications that are pulled by underlaying orchestrators

# Getting Started

## Set enviroment variables

| Variable         | Description     |
| ---------------- | --------------- |
| DB_DRIVER        | Example: mysql           |
| DB_NAME          | Example: jobmanager      |
| DB_USER          | Database user            |
| DB_PASSWORD      | Database password  |
| DB_ROOT_PASSWORD | Database root password |
| DB_HOST          | Example: localhost       |
| DB_PORT          | Example: 3306            |
| SERVER_PORT      | Example: 8082            |


## Build the application

`make build`

## Start test database

`make start-database`

## Run application

`make run`