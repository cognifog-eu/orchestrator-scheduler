package controllers

import (
	m "icos/server/jobmanager-service/middlewares"
)

func (s *Server) initializeRoutes() {
	// Home Route
	s.Router.HandleFunc("/jobmanager", m.SetMiddlewareLog(m.SetMiddlewareJSON(s.Home))).Methods("GET")
	// JobManager Routes
	// get all jobs GET <- driver
	s.Router.HandleFunc("/jobmanager/jobs", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetAllJobs)))).Methods("GET")
	// request deployment POST <- shell
	s.Router.HandleFunc("/jobmanager/jobs/create", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.CreateJob)))).Methods("POST")
	// delete job / undeploy? DELETE <- shell
	s.Router.HandleFunc("/jobmanager/jobs/delete", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.DeleteJob)))).Methods("DELETE")
	// get job status GET <- driver
	s.Router.HandleFunc("/jobmanager/jobs/:id", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobByUUID)))).Methods("GET")
	// get all jobs with specific state GET <- driver
	s.Router.HandleFunc("/jobmanager/jobs/:state", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobsByState)))).Methods("GET")
	// update job
	s.Router.HandleFunc("/jobmanager/jobs/:state", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.UpdateAJob)))).Methods("PUT")

	//healthcheck TODO
}
