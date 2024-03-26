package main

import (
	_ "etsn/server/docs"
	jobmanager_service "etsn/server/jobmanager-service"
)

//	@title			Swagger Job Manager API
//	@version		1.0
//	@description	Job Manager Microservice.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8082
//	@BasePath	/

//	@securityDefinitions.basic	OAuth 2.0

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {

	jobmanager_service.Run()

}
