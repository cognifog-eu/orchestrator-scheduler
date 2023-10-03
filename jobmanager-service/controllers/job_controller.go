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
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// gorm retrieve
	job := models.Job{}
	jobGotten, err := job.FindJobByUUID(server.DB, uint32(uid))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, models.Job{})

}

func (server *Server) GetAllJobs(w http.ResponseWriter, r *http.Request) {
}

func (server *Server) GetJobsByState(w http.ResponseWriter, r *http.Request) {
}

func (server *Server) CreateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	job := models.Job{
		State: models.Created,
		Type:  models.CreateDeployment,
	}

	// matchmaking + optimization = targets -> async?

	//

	// gorm save

	jobCreated, err := job.SaveJob(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, job)
}

func (server *Server) DeleteJob(w http.ResponseWriter, r *http.Request) {
}
