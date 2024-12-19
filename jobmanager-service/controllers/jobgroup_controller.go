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
	"errors"
	"etsn/server/jobmanager-service/responses"
	"etsn/server/jobmanager-service/utils/logs"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

// CreateJobGroup godoc
//
//	@Summary		create new JobGroup
//	@Description	create new jobgroup
//	@Tags			jobgroups
//	@Accept			plain
//	@Produce		json
//	@Param			application	body		string			true	"Application manifest YAML"
//	@Success		201			{object}	models.JobGroup	"Created"
//	@Failure		422			{object}	string			"Unprocessable Entity"
//	@Router			/jobmanager/groups [post]
func (server *Server) CreateJobGroup(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	jobGroup, err := server.JobGroupService.CreateJobGroup(bodyBytes, r.Header)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// notify policy manager
	server.PolicyService.NotifyPolicyManager(string(bodyBytes), jobGroup, r.Header.Get("Authorization"))
	responses.JSON(w, http.StatusCreated, jobGroup)
}

// GetJobGroupByUUID godoc
//
//	@Summary		Get JobGroup by UUID
//	@Description	get jobgroup by uuid
//	@Tags			jobgroups
//	@Accept			json
//	@Produce		json
//	@Param			group_uuid	path		string	true	"JobGroup UUID"
//	@Success		200			{object}	models.JobGroup
//	@Failure		400			{object}	string	"Bad Request"
//	@Failure		404			{object}	string	"Not Found"
//	@Router			/jobmanager/groups/{group_uuid} [get]
func (server *Server) GetJobGroupByUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["group_uuid"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	jobGroupGotten, err := server.JobGroupService.FindJobGroupByUUID(stringID)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGroupGotten)
}

// DeleteJobGroup godoc
//
//	@Summary		delete job group by UUID
//	@Description	delete job group by uuid
//	@Tags			jobgroups
//	@Accept			json
//	@Produce		json
//	@Param			group_uuid	path		string	true	"JobGroup UUID"
//	@Success		200			{string}	string
//	@Failure		400			{object}	string	"Bad Request"
//	@Failure		404			{object}	string	"Not Found"
//	@Router			/jobmanager/groups/{group_uuid} [delete]
func (server *Server) DeleteJobGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["group_uuid"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		logs.Logger.Println("JobGroup's ID is empty!")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Handle the deletion through the service
	jobGroupDeleted, err := server.JobGroupService.DeleteJobGroupByID(stringID)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGroupDeleted)
}

// GetAllJobGroups godoc
//
//	@Summary		Get All JobGroups
//	@Description	get all jobgroups
//	@Tags			jobgroups
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		[]models.JobGroup
//	@Failure		400	{object}	string	"Bad Request"
//	@Failure		404	{object}	string	"Not Found"
//	@Router			/jobmanager/groups [get]
func (server *Server) GetAllJobGroups(w http.ResponseWriter, r *http.Request) {
	jobGroupsGotten, err := server.JobGroupService.FindAllJobGroups()
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGroupsGotten)
}

// StopJobGroupByUUID godoc
//
//	@Summary		Stop JobGroup by UUID
//	@Description	stop jobgroup by uuid
//	@Tags			jobgroups
//	@Accept			json
//	@Produce		json
//	@Param			group_uuid	path		string	true	"JobGroup UUID"
//	@Success		200			{object}	models.JobGroup
//	@Failure		400			{object}	string	"Bad Request"
//	@Failure		404			{object}	string	"Not Found"
//	@Router			/jobmanager/groups/stop/{group_uuid} [put]
func (server *Server) StopJobGroupByUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["group_uuid"]
	if id == "" {
		err := errors.New("ID Cannot be empty")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Handle the stopping through the service
	jobGroupStopped, err := server.JobGroupService.StopJobGroupByID(id)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGroupStopped)
}

// StartJobGroupByUUID godoc
//
//	@Summary		Start JobGroup by UUID
//	@Description	Start jobgroup by uuid
//	@Tags			jobgroups
//	@Accept			json
//	@Produce		json
//	@Param			group_uuid	path		string	true	"JobGroup UUID"
//	@Success		200			{object}	models.JobGroup
//	@Failure		400			{object}	string	"Bad Request"
//	@Failure		404			{object}	string	"Not Found"
//	@Router			/jobmanager/groups/start/{group_uuid} [put]
func (server *Server) StartJobGroupByUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["group_uuid"]
	if id == "" {
		err := errors.New("ID Cannot be empty")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Handle the stopping through the service
	jobGroupStopped, err := server.JobGroupService.StartJobGroupByID(id, r.Header)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGroupStopped)
}

// UpdateJobGroup godoc
//
//	@Summary		update a JobGroup
//	@Description	update a jobgroup
//	@Tags			jobgroups
//	@Accept			json
//	@Produce		json
//	@Param			JobGroup	body		models.JobGroup	true	"JobGroup information"
//	@Success		200			{object}	models.JobGroup
//	@Failure		400			{object}	string	"Bad Request"
//	@Failure		404			{object}	string	"Not Found"
//	@Router			/jobmanager/groups [put]
func (server *Server) UpdateJobGroup(w http.ResponseWriter, r *http.Request) {
	bodyJob, err := io.ReadAll(r.Body)
	if err != nil {
		logs.Logger.Println("Error reading request body:", err)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	jobGroupUpdated, err := server.JobGroupService.ReplaceJobGroup(bodyJob)
	if err != nil {
		logs.Logger.Println("Error updating job group:", err)
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	responses.JSON(w, http.StatusOK, jobGroupUpdated)
}
