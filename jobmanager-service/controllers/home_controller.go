package controllers

import (
	"icos/server/jobmanager-service/responses"
	"net/http"
)

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, "Welcome To ICOS Job Manager API")

}
