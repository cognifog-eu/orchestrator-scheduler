package controllers

import (
	"encoding/json"
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/responses"
	"io"
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

	responses.JSON(w, http.StatusOK, jobGotten)

}

func (server *Server) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	// gorm retrieve
	job := models.Job{}
	jobsGotten, err := job.FindAllJobs(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobsGotten)
}

func (server *Server) GetJobsByState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	state, err := strconv.ParseInt(vars["state"], 10, 32)
	// state validation
	if !models.StateIsValid(int(state)) {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// gorm retrieve
	job := models.Job{}
	jobGotten, err := job.FindJobsByState(server.DB, int(state))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGotten)
}

func (server *Server) CreateJob(w http.ResponseWriter, r *http.Request) {
	job := models.Job{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	err = json.Unmarshal(body, &job)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// validate job
	// matchmaking + optimization = targets -> sync?
	// append targets to jobs app description
	// gorm save

	jobCreated, err := job.SaveJob(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobCreated)
}

func (server *Server) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// gorm retrieve
	job := models.Job{}
	jobDeleted, err := job.DeleteAJob(server.DB, uint32(uid))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobDeleted)
}
