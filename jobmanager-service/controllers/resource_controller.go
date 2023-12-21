package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/responses"
	"icos/server/jobmanager-service/utils/logs"
	"io"
	"net/http"
	"os"

	uuid "github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
	deployManagerBaseURL = os.Getenv("DEPLOY_MANAGER_BASE_URL")
)

func (server *Server) GetAllResources(w http.ResponseWriter, r *http.Request) {
	// gorm retrieve
	job := models.Job{}
	jobsGotten, err := job.FindAllJobs(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobsGotten)
}

func (server *Server) GetResourceStateByUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["id"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uuid, err := uuid.Parse(stringID)
	resourceStatus := models.Resource{}
	// retrieve info from the job first. need extra info!
	job := models.Job{}
	jobGotten, err := job.FindJobByResourceUUID(server.DB, uuid)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	jobString, err := json.Marshal(jobGotten)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	logs.Logger.Println("Job found: " + string(jobString))
	// this can have some sort of cache to avoid making requests to DM

	// create DM request
	req, err := http.NewRequest("GET", deployManagerBaseURL+"/deploy-manager/resource", bytes.NewBuffer([]byte{}))
	query := req.URL.Query()
	query.Add("uuid", uuid.String())
	query.Add("node_target", jobGotten.Targets[0].NodeName)
	query.Add("resource_name", jobGotten.Resource.ResourceName)
	query.Encode()
	logs.Logger.Println("request status to DM: " + req.URL.String())
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// forward the authorization token
	req.Header.Add("Authorization", r.Header.Get("Authorization"))

	// // do request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	defer resp.Body.Close()

	// direct body read
	resourceStatusBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// parse to application objects
	err = json.Unmarshal(resourceStatusBody, &resourceStatus)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusOK, resourceStatus)

}

func (server *Server) UpdateResourceStateByUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["id"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uuid, err := uuid.Parse(stringID)
	resource := models.Resource{}
	resourceBody, err := io.ReadAll(r.Body)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	// parse to application objects
	err = json.Unmarshal(resourceBody, &resource)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// update resource details first
	_, err = resource.UpdateAResourceStatus(server.DB, uuid)
	if err != nil {
		logs.Logger.Println("Resource were not found during status update")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, resource)
}
