package controllers

import (
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/responses"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	lighthouseBaseURL  = "http://lighthouse.icos-project.eu:8080"
	apiV3              = "/api/v3"
	matchmackerBaseURL = ""
)

func (server *Server) GetJobByUUID(w http.ResponseWriter, r *http.Request) {

}

func (server *Server) CreateJob(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	job := models.Job{}

}
