package controllers

import (
	m "icos/server/jobmanager-service/middlewares"
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
	s.Router.HandleFunc("/jobmanager/jobs/create/{app_name}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.CreateJob)))).Methods("POST")
	// get all jobs with specific state GET <- driver
	s.Router.HandleFunc("/jobmanager/jobs/executable", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobsByState)))).Methods("GET")
	// get job status GET <- driver
	s.Router.HandleFunc("/jobmanager/jobs/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobByUUID)))).Methods("GET")
	// update job
	s.Router.HandleFunc("/jobmanager/jobs/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.UpdateAJob)))).Methods("PUT")
	// delete job / undeploy? DELETE <- shell
	s.Router.HandleFunc("/jobmanager/jobs/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.DeleteJob)))).Methods("DELETE")
	// get resource status
	s.Router.HandleFunc("/jobmanager/resources/status/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetResourceStateByUUID)))).Methods("GET")
	// update status
	s.Router.HandleFunc("/jobmanager/resources/status/{id}", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.UpdateResourceStateByUUID)))).Methods("PUT")
}
