/*
Copyright 2023 Bull SAS

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
	m "cognifog/server/jobmanager-service/middlewares"
)

func (s *Server) initializeRoutes() {
	// Home Route
	s.Router.HandleFunc("/jobmanager", m.SetMiddlewareLog(m.SetMiddlewareJSON(s.Home))).Methods("GET")
	//healthcheck
	s.Router.HandleFunc("/jobmanager/healthz", s.HealthCheck).Methods("GET")
	// JobManager Routes
	// get all jobs GET <- driver
	s.Router.HandleFunc("/jobmanager/jobs", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetAllJobs)))).Methods("GET")
	// request deployment POST <- shell
	s.Router.HandleFunc("/jobmanager/jobs/create", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.CreateJob)))).Methods("POST")
	// get all jobs with specific state GET <- driver
	s.Router.HandleFunc("/jobmanager/jobs/executable", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobsByState)))).Methods("GET")
	// get job status GET <- driver
	s.Router.HandleFunc("/jobmanager/jobs/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobByUUID)))).Methods("GET")
	// update job
	s.Router.HandleFunc("/jobmanager/jobs/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.UpdateAJob)))).Methods("PUT")
	// delete job / undeploy? DELETE <- shell
	s.Router.HandleFunc("/jobmanager/jobs/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.DeleteJob)))).Methods("DELETE")
}
