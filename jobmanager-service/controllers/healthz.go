package controllers

import (
	"icos/server/jobmanager-service/responses"
	"net/http"
)

func (server *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, "Job Manager working properly!")
}
