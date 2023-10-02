package controllers

import (
	m "icos/server/jobmanager-service/middlewares"
)

func (s *Server) initializeRoutes() {

	// Home Route
	s.Router.HandleFunc("/jobmanager", m.SetMiddlewareLog(m.SetMiddlewareJSON(s.Home))).Methods("GET")

	//Lobbies routes
	s.Router.HandleFunc("/jobmanager/jobs/:id", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobByUUID)))).Methods("GET")
	s.Router.HandleFunc("/jobmanager/jobs/:id", m.SetMiddlewareLog(m.SetMiddlewareJSON(m.JWTValidation(s.GetJobByUUID)))).Methods("POST")
}
