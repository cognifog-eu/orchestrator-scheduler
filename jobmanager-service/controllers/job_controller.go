/*
Copyright 2023-2024 Bull SAS

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
	"etsn/server/jobmanager-service/utils/logs"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// GetAllJobs godoc
//
//	@Summary		List all Jobs
//	@Description	get all jobs
//	@Tags			jobs
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		[]models.Job
//	@Failure		404	{object}	string	"Can not find Jobs"
//	@Router			/jobmanager/jobs [get]
func (server *Server) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	jobsGotten, err := server.JobService.FindAllJobs()
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	responses.JSON(w, http.StatusOK, jobsGotten)
}

// GetJobByUUID godoc
//
//	@Summary		Get Job by UUID
//	@Description	get job by uuid
//	@Tags			jobs
//	@Accept			json
//	@Produce		json
//	@Param			job_uuid	path		string		true	"Job UUID"
//	@Success		200			{object}	models.Job	"Ok"
//	@Failure		400			{object}	string		"Job UUID is required"
//	@Failure		404			{object}	string		"Can not find Job by UUID"
//	@Router			/jobmanager/jobs/{job_uuid} [get]
func (server *Server) GetJobByUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["job_uuid"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	jobGotten, err := server.JobService.FindJobByUUID(stringID)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGotten)

}

// GetJobsByState godoc
//
//	@Summary		List Jobs to Execute
//	@Description	get jobs to execute
//	@Tags			jobs
//	@Accept			json
//	@Produce		json
//	@Param			orchestrator	path		string	true	"Orchestrator type [ocm | nuvla]"
//	@Param			owner_id		path		string	true	"Owner ID"
//	@Success		200				{array}		[]models.Job
//	@Failure		400				{object}	string	"Orchestrator type is required"
//	@Failure		404				{object}	string	"Cannot find executable Jobs"
//	@Failure		500				{object}	string	"Internal server error"
//	@Router			/jobmanager/jobs/executable/{orchestrator}/{owner_id} [get]
func (server *Server) GetJobsByState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orch := vars["orchestrator"]
	ownerID := vars["owner_id"] // temporal solution to test
	// state validation
	if models.None == models.OrchestratorTypeMapper(orch) {
		err := errors.New("no valid orchestrator type provided")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Fetch jobs to execute
	jobGotten, err := server.JobService.FindJobsToExecute(orch, ownerID)
	if err != nil {
		// Specific error handling for different scenarios
		if errors.Is(err, gorm.ErrRecordNotFound) {
			responses.ERROR(w, http.StatusNotFound, err)
		} else {
			responses.ERROR(w, http.StatusInternalServerError, err)
		}
		return
	}

	// Respond with the jobs found
	responses.JSON(w, http.StatusOK, jobGotten)
}

// DeleteJob godoc
//
//	@Summary		Delete Job by UUID
//	@Description	delete job by uuid
//	@Tags			jobs
//	@Accept			json
//	@Produce		json
//	@Param			job_uuid	path		string	true	"Job UUID"
//	@Success		200			{string}	string	"Ok"
//	@Failure		400			{object}	string	"Job UUID is required"
//	@Failure		404			{object}	string	"Can not find Job to delete"
//	@Router			/jobmanager/jobs/{job_uuid} [delete]
func (server *Server) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["job_uuid"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		logs.Logger.Println("JOB's ID is empty!")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	jobDeleted, err := server.JobService.DeleteJob(stringID)
	if err != nil {
		responses.ERROR(w, http.StatusServiceUnavailable, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobDeleted)
}

// UpdateAJob godoc
//
//	@Summary		Update a Job
//	@Description	update a job
//	@Tags			jobs
//	@Accept			json
//	@Produce		json
//	@Param			job_uuid	path		string		true	"Job UUID"
//	@Param			Job			body		models.Job	true	"Job information"
//	@Success		200			{object}	models.Job
//	@Failure		400			{object}	string	"Error reading request body"
//	@Failure		404			{object}	string	"Could not find job"
//	@Router			/jobmanager/jobs [put]
func (server *Server) UpdateAJob(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // Ensure the body is closed after reading

	job := models.Job{}
	bodyJob, err := io.ReadAll(r.Body)
	if err != nil {
		logs.Logger.Println("Error reading request body:", err)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = json.Unmarshal(bodyJob, &job)
	if err != nil {
		logs.Logger.Println("Error unmarshalling job:", err)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	jobUpdated, err := server.JobService.ReportJobState(&job)
	if err != nil {
		logs.Logger.Println("Error finding job:", err)
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	if jobUpdated.SubType == models.Reallocate && jobUpdated.Type == models.DeleteDeployment {
		jobUpdated, err = server.JobService.ReallocateJob(jobUpdated, r.Header)
		if err != nil {
			logs.Logger.Println("Error reallocating job:", err)
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
	}

	responses.JSON(w, http.StatusOK, jobUpdated)
}

// PromoteJobByUUID godoc
//
//	@Summary		Promote Job by UUID
//	@Description	promote job by uuid
//	@Tags			jobs
//	@Accept			json
//	@Produce		json
//	@Param			job_uuid	path		string	true	"Job UUID"
//	@Success		204			{string}	string	"Job Promoted"
//	@Failure		400			{object}	string	"Job UUID is required"
//	@Failure		404			{object}	string	"Can not find Job to promote"
//	@Router			/jobmanager/jobs/promote/{job_uuid} [patch]
func (server *Server) PromoteJobByUUID(w http.ResponseWriter, r *http.Request) {
	_, err := server.JobService.JobPromote(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusNoContent, http.NoBody)
}
