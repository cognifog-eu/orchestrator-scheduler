# Job Manager

Job Manager Microservice: In charge of Creating and Managing Jobs that are pulled by underlaying orchestrators

# Interfaces documentation

Swagger Documentation is implemented into the asset. It may be found in the `docs` folder

An example of the data model recieved by the component may be found at the [Dashboard documentation](https://github.com/cognifog-eu/comprehensive-dashboard/blob/9cef08ed227a661eca467d843ecbfa0520695c67/README.md#yaml-based-syntax)

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
| MOCK_TARGET_CLUSTER | Temporary until there is a working matchmaking service |


## Run containerized with embeded database

Build application into container

`make build-container`

Start application from container with embeded database

`make start-database`

## Run application locally

`make run`
