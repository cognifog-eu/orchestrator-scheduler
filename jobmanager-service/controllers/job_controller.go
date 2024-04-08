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
	"encoding/json"
	"errors"
	"etsn/server/jobmanager-service/models"
	"etsn/server/jobmanager-service/responses"
	"etsn/server/jobmanager-service/utils"
	"etsn/server/jobmanager-service/utils/logs"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	uuid "github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
	// lighthouseBaseURL  = os.Getenv("LIGHTHOUSE_BASE_URL")
	// apiV3              = "/api/v3"
	matchmackerBaseURL = os.Getenv("MATCHMAKING_URL")
)

// GetAllJobs example
//
//	@Description	Get All Jobs
//	@ID				get-all-jobs
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string			true	"Authentication header"
//	@Success		200				{string}	[]models.Job	"Ok"
//	@Failure		404				{object}	string			"Can not find Jobs"
//	@Router			/jobmanager/jobs [get]
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

// GetJobByUUID example
//
//	@Description	get Job by UUID
//	@ID				get-job-by-uuid
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"Authentication header"
//	@Param			id				path		string		true	"Job ID"
//	@Success		200				{string}	models.Job	"Ok"
//	@Failure		400				{object}	string		"Job ID is required"
//	@Failure		404				{object}	string		"Can not find Job by ID"
//	@Router			/jobmanager/jobs/{id} [get]
func (server *Server) GetJobByUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["id"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uuid, err := uuid.Parse(stringID)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// gorm retrieve
	job := models.Job{}
	jobGotten, err := job.FindJobByUUID(server.DB, uuid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGotten)

}

// GetJobsByState example
//
//	@Description	get Jobs by State
//	@ID				get-jobs-by-state
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string			true	"Authentication header"
//	@Param			orchestrator	header		string			true	"Orchestrator type [ ocm | nuvla ]"
//	@Success		200				{string}	[]models.Job	"Ok"
//	@Failure		400				{object}	string			"Orchestrator type is required"
//	@Failure		404				{object}	string			"Can not find executable Jobs"
//	@Router			/jobmanager/jobs/executable [get]
//
// retrieves only executable jobs for now; TODO: improve
func (server *Server) GetJobsByState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orch := vars["orchestrator"]
	// state validation
	if models.None == models.OrchestratorTypeMapper(orch) {
		err := errors.New("no valid Orchestrator type provided")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// gorm retrieve
	job := models.Job{}
	// retrieves jobs that are created && not locked or progressing && locked for more than 5 minutes
	jobGotten, err := job.FindJobsToExecute(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGotten)
}

// CreateJob example
//
//	@Description	create new Job
//	@ID				create-new-job
//	@Accept			plain
//	@Produce		json
//	@Param			Authorization	header		string		true	"Authentication header"
//	@Param			app_name		path		string		true	"Application name"
//	@Param			application		body		string		true	"Application manifest YAML"
//	@Success		200				{object}	models.Job	"Ok"
//	@Failure		400				{object}	string		"Application name is required"
//	@Router			/jobmanager/jobs/create/{app_name} [post]
func (server *Server) CreateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app_name"]
	appDesc := vars["job_group_description"]
	logs.Logger.Println("Job group name is: " + appName)
	logs.Logger.Println("Job group description is: " + appDesc)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	bodyString := string(bodyBytes)
	bodyStringTrimmed := strings.Trim(bodyString, "\r\n")
	logs.Logger.Println("Trimmed body: " + bodyStringTrimmed)
	var mMResponseMapper models.MMResponseMapper

	splittedManifests := strings.Split(bodyStringTrimmed, "---")

	// MM Mock from env

	logs.Logger.Println("Mocking MM response: " + os.Getenv("MOCK_TARGET_CLUSTER"))
	for i := 1; i < len(splittedManifests); i++ { //starts in 1 to drop first application description manifest
		mMResponseMapper.Components = append(mMResponseMapper.Components, models.Component{
			Name:      "name-" + strconv.FormatInt(int64(i), 10),
			Kind:      "kind-" + strconv.FormatInt(int64(i), 10),
			Manifests: "",
			Targets: []models.Target{{
				ClusterName:  os.Getenv("MOCK_TARGET_CLUSTER"),
				NodeName:     os.Getenv("MOCK_TARGET_CLUSTER"),
				Orchestrator: "ocm",
			}},
		})
	}

	/* // create MM request
	req, err := http.NewRequest("POST", matchmackerBaseURL+"/matchmake", bytes.NewBuffer(bodyBytes))
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// add content type
	req.Header.Set("Content-Type", "application/x-yaml")
	// forward the authorization token
	req.Header.Add("Authorization", r.Header.Get("Authorization"))

	// do request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	defer resp.Body.Close()

	// if response is OK
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		// direct body read
		bodyMM, err := io.ReadAll(resp.Body)
		if err != nil {
			logs.Logger.Println("ERROR " + err.Error())
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		// debug
		dst := &bytes.Buffer{}
		_ = json.Indent(dst, bodyMM, "", "  ")
		logs.Logger.Println("MM response is: " + dst.String())
		// parse to application objects
		err = json.Unmarshal(bodyMM, &mMResponseMapper)
		if err != nil {
			logs.Logger.Println("ERROR " + err.Error())
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
	*/
	// initialize conditions
	conditions := []models.Condition{
		{
			Type:               "Created",
			Status:             "True",
			ObservedGeneration: 1,
			LastTransitionTime: time.Now(),
			Reason:             "AwaitingForTarget",
			Message:            "Waiting for the Target",
		},
		{
			Type:               "Available",
			Status:             "True",
			ObservedGeneration: 1,
			LastTransitionTime: time.Now(),
			Reason:             "Waiting for Job to be taken",
			Message:            "Resource does not exist",
		},
	}

	// create Job Group
	jobGroup := models.JobGroup{
		AppName:        appName,
		AppDescription: appDesc,
	}

	// remove policies block before parsing
	// mMResponseMapper.Components = mMResponseMapper.Components[:len(mMResponseMapper.Components)-1]
	// create job per component
	for i, comp := range mMResponseMapper.Components {

		// create unique namespace
		rs := utils.RandomString(10)

		// since we are iterating targets, for each target a job to create the namespace must be created
		// this job contains a namespace declaration, but in only applies if the target is a kubernetes cluster
		if comp.Targets[0].Orchestrator == "ocm" {
			nSJob := models.Job{
				Type:         models.CreateDeployment,
				State:        models.JobCreated,
				Manifest:     splittedManifests[i+1], // TODO, this could fail so easily..
				Targets:      comp.Targets,
				JobGoupName:  jobGroup.AppName,
				Orchestrator: comp.Targets[0].Orchestrator, // TODO, should not point to first element
				Namespace:    appName + rs,                 // Unique within a single cluster
				Resource: models.Resource{
					ResourceName: comp.Name,
					Conditions:   conditions,
				},
			}
			jobGroup.Jobs = append(jobGroup.Jobs, nSJob)
			logs.Logger.Println("Namespace appended to JobGroup: " + nSJob.Namespace)
		}

	}
	// gorm save
	_, err = jobGroup.SaveJobGroup(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// notify policy manager
	//models.NotifyPolicyManager(server.DB, bodyStringTrimmed, jobGroup, r.Header.Get("Authorization"))

	logs.Logger.Println("APPDESC: " + jobGroup.AppDescription)
	responses.JSON(w, http.StatusCreated, jobGroup) // TODO change
	/*
		 	} else {
				err := errors.New("Matchmaking process did not return valid targets: status code - " + string(rune(resp.StatusCode)))
				responses.ERROR(w, http.StatusInternalServerError, err)
				return
			}
	*/
}

// DeleteJob example
//
//	@Description	delete job by id
//	@ID				delete-job-by-id
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication header"
//	@Param			id				path		string	true	"Job ID"
//	@Success		200				{string}	string	"Ok"
//	@Failure		400				{object}	string	"ID is required"
//	@Failure		404				{object}	string	"Can not find Job to delete"
//	@Router			/jobmanager/jobs/{id} [delete]
func (server *Server) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["id"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		fmt.Println("JOB's ID is empty!")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uuid, err := uuid.Parse(stringID)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// gorm retrieve
	job := models.Job{}
	jobDeleted, err := job.DeleteAJob(server.DB, uuid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobDeleted)
}

// UpdateAJob example
//
//	@Description	update a job
//	@ID				update-a-job
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string		true	"Authentication header"
//	@Param			id				path		string		true	"Job ID"
//	@Param			Job				body		models.Job	true	"Job information"
//	@Success		200				{object}	models.Job	"Ok"
//	@Failure		400				{object}	string		"ID is required"
//	@Failure		404				{object}	string		"Can not find Job to update"
//	@Router			/jobmanager/jobs/{id} [put]
func (server *Server) UpdateAJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["id"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		fmt.Println("JOB's ID is empty!")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	id, err := uuid.Parse(stringID)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	resource := models.Resource{}
	job := models.Job{}
	bodyJob, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	err = json.Unmarshal(bodyJob, &job)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	logs.Logger.Println("Reading job to update " + job.ID.String())

	// update resource details first
	_, err = resource.UpdateAResource(server.DB, job.ID, job.UUID)
	if err != nil {
		logs.Logger.Println("Resource were not found during Job update")
		// responses.ERROR(w, http.StatusBadRequest, err)
		// return
	}
	jobUpdated, err := job.UpdateAJob(server.DB, id)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobUpdated)
}

// GetJobGroup example
//
//	@Description	Get JobGroup
//	@ID				get-jobgroup
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication header"
//	@Success		200				{string}	[]models.JobGroup	"Ok"
//	@Failure		404				{object}	string				"Can not find JobGroup"
//	@Router			/jobmanager/jobs [get]
func (server *Server) GetJobGroupByUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["id"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uuid, err := uuid.Parse(stringID)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// gorm retrieve
	jobGroup := models.JobGroup{}
	jobGroupGotten, err := jobGroup.FindJobGroupByUUID(server.DB, uuid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGroupGotten)
}

// DeleteJobGroup example
//
//	@Description	delete job group by id
//	@ID				delete-job-group-by-id
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authentication header"
//	@Param			id				path		string	true	"JobGroup ID"
//	@Success		200				{string}	string	"Ok"
//	@Failure		400				{object}	string	"ID is required"
//	@Failure		404				{object}	string	"Can not find JobGroup to delete"
//	@Router			/jobmanager/jobs/group/{id} [delete]
func (server *Server) DeleteJobGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["id"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		fmt.Println("JobGroup's ID is empty!")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uuid, err := uuid.Parse(stringID)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// gorm retrieve
	jobGroup := models.JobGroup{}
	jobGroupDeleted, err := jobGroup.DeleteAJobGroup(server.DB, uuid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGroupDeleted)
}
