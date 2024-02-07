# Job Manager

Job Manager Microservice: In charge of Creating and Managing Jobs that are pulled by underlaying orchestrators

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


## Run containerized with embeded database

Build application into container

`make build-container`

Start application from container with embeded database

`make start-database`

## Run application locally

`make run`
