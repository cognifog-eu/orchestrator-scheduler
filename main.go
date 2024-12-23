/*
Copyright 2023-2024 Bull SAS

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	_ "etsn/server/docs"
	jobmanager_service "etsn/server/jobmanager-service"
)

//	@title			Swagger Job Manager API
//	@version		1.4.1-latest
//	@description	Job Manager Microservice.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8082
//	@BasePath	/

//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				"Type 'Bearer TOKEN' to correctly set the API Key"

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {

	// // programmatically set swagger info
	// docs.SwaggerInfo.Title = "Swagger Example API"
	// docs.SwaggerInfo.Description = "This is a sample server Petstore server."
	// docs.SwaggerInfo.Version = "1.0"
	// docs.SwaggerInfo.Host = "petstore.swagger.io"
	// docs.SwaggerInfo.BasePath = "/v2"
	// docs.SwaggerInfo.Schemes = []string{"http", "https"}

	jobmanager_service.Run()

}
