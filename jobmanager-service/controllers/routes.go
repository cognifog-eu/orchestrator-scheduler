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
	m "etsn/server/jobmanager-service/middlewares"
)

func (s *Server) initializeRoutes() {
	// Home Route
	s.Router.HandleFunc("/jobmanager", m.SetMiddlewareLog(m.SetMiddlewareJSON(s.Home))).Methods("GET")
	//healthcheck
	s.Router.HandleFunc("/jobmanager/healthz", s.HealthCheck).Methods("GET")
	// get all jobs GET
	s.Router.HandleFunc("/jobmanager/jobs", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetAllJobs)))).Methods("GET")
	// get job status GET <- driver
	s.Router.HandleFunc("/jobmanager/jobs/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobByUUID)))).Methods("GET")
	// update job
	s.Router.HandleFunc("/jobmanager/jobs/update", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.UpdateAJob)))).Methods("PUT")
	// delete job. DELETE <- shell
	s.Router.HandleFunc("/jobmanager/jobs/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.DeleteJob)))).Methods("DELETE")
	// lock a job. PATCH
	s.Router.HandleFunc("/jobmanager/jobs/lock/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.LockJobByUUID)))).Methods("PATCH")
	// get job group GET
	s.Router.HandleFunc("/jobmanager/jobs/group/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobGroupByUUID)))).Methods("GET")
	// delete jobGroup / undeploy. DELETE <- shell
	s.Router.HandleFunc("/jobmanager/jobs/group/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.DeleteJobGroup)))).Methods("DELETE")
	// request deployment POST <- shell
	s.Router.HandleFunc("/jobmanager/jobs/create/{app_name}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.CreateJob)))).Methods("POST")
	// get all jobs with specific state GET <- driver
	s.Router.HandleFunc("/jobmanager/jobs/executable/orchestrator/{orchestrator}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobsByState)))).Methods("GET")
	// get all job groups GET
	s.Router.HandleFunc("/jobmanager/jobgroups", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetAllJobGroups)))).Methods("GET")
	// undeploy JobGroup PUT
	s.Router.HandleFunc("/jobmanager/jobgroups/undeploy/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.UndeployJobGroupByUUID)))).Methods("PUT")
	// get resource status
	s.Router.HandleFunc("/jobmanager/resources/status/{job_id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetResourceStateByJobUUID)))).Methods("GET")
	// update status PUT
	s.Router.HandleFunc("/jobmanager/resources/status/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.UpdateResourceStateByUUID)))).Methods("PUT")
	// create policy violation JM <- PM POST
	s.Router.HandleFunc("/jobmanager/policies/incompliance/create", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.CreatePolicyIncompliance)))).Methods("POST")
}
