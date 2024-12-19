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
package controllers

import (
	m "etsn/server/jobmanager-service/middlewares"
	"etsn/server/jobmanager-service/utils/logs"
	"net/http"
)

func (s *Server) initializeRoutes(enableJWT bool) {

	middlewares := []func(http.HandlerFunc) http.HandlerFunc{
		m.SetMiddlewareLog,
		m.SetMiddlewareJSON,
	}

	if enableJWT {
		logs.Logger.Println("JWT Middleware is enabled")
		middlewares = append(middlewares, m.JWTValidation)
	} else {
		logs.Logger.Println("JWT Middleware is disabled")
	}

	// Home Route
	s.Router.HandleFunc("/jobmanager", applyMiddlewares(s.Home, middlewares[0], middlewares[1])).Methods("GET")

	// Healthcheck
	s.Router.HandleFunc("/jobmanager/healthz", s.HealthCheck).Methods("GET")

	// Job Routes
	s.Router.HandleFunc("/jobmanager/jobs", applyMiddlewares(s.GetAllJobs, middlewares...)).Methods("GET")
	s.Router.HandleFunc("/jobmanager/jobs", applyMiddlewares(s.UpdateAJob, middlewares...)).Methods("PUT")
	s.Router.HandleFunc("/jobmanager/jobs/executable/{orchestrator}/{owner_id}", applyMiddlewares(s.GetJobsByState, middlewares...)).Methods("GET")
	s.Router.HandleFunc("/jobmanager/jobs/{job_uuid}", applyMiddlewares(s.GetJobByUUID, middlewares...)).Methods("GET")
	s.Router.HandleFunc("/jobmanager/jobs/{job_uuid}", applyMiddlewares(s.DeleteJob, middlewares...)).Methods("DELETE")
	s.Router.HandleFunc("/jobmanager/jobs/promote/{job_uuid}", applyMiddlewares(s.PromoteJobByUUID, middlewares...)).Methods("PATCH")

	// Job Group Routes
	s.Router.HandleFunc("/jobmanager/groups", applyMiddlewares(s.CreateJobGroup, middlewares...)).Methods("POST")
	s.Router.HandleFunc("/jobmanager/groups", applyMiddlewares(s.GetAllJobGroups, middlewares...)).Methods("GET")
	s.Router.HandleFunc("/jobmanager/groups", applyMiddlewares(s.UpdateJobGroup, middlewares...)).Methods("PUT")
	s.Router.HandleFunc("/jobmanager/groups/{group_uuid}", applyMiddlewares(s.GetJobGroupByUUID, middlewares...)).Methods("GET")
	s.Router.HandleFunc("/jobmanager/groups/{group_uuid}", applyMiddlewares(s.DeleteJobGroup, middlewares...)).Methods("DELETE")
	s.Router.HandleFunc("/jobmanager/groups/stop/{group_uuid}", applyMiddlewares(s.StopJobGroupByUUID, middlewares...)).Methods("PUT")
	s.Router.HandleFunc("/jobmanager/groups/start/{group_uuid}", applyMiddlewares(s.StartJobGroupByUUID, middlewares...)).Methods("PUT")

	// Resource Routes
	s.Router.HandleFunc("/jobmanager/resources/status/{job_uuid}", applyMiddlewares(s.GetResourceStateByJobUUID, middlewares...)).Methods("GET")
	s.Router.HandleFunc("/jobmanager/resources/status", applyMiddlewares(s.UpdateResourceStateByUUID, middlewares...)).Methods("PUT")

	// Policy Incompliance
	s.Router.HandleFunc("/jobmanager/policies/incompliance", applyMiddlewares(s.CreatePolicyIncompliance, middlewares...)).Methods("POST")

}

func applyMiddlewares(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
