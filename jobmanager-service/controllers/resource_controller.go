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
)

// Deprecated - Add in future releases
func (server *Server) GetAllResources(w http.ResponseWriter, r *http.Request) {
	// gorm retrieve
	jobsGotten, err := server.JobService.FindAllJobs()
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobsGotten)
}

// GetResourceStateByJobUUID example
//
//	@Summary		Get resource status by job UUID
//	@Description	get resource status by job uuid
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			job_uuid	path		string	true	"Job UUID"
//	@Success		200			{object}	models.Resource
//	@Failure		400			{object}	string	"Job UUID is required"
//	@Failure		404			{object}	string	"Can not find Job by UUID"
//	@Router			/jobmanager/resources/status/{job_uuid} [get]
func (server *Server) GetResourceStateByJobUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["job_uuid"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	resourceGotten, err := server.ResourceService.FindResourceByJobUUID(stringID)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	responses.JSON(w, http.StatusOK, resourceGotten.Conditions)

}

// UpdateResourceStateByUUID example
//
//	@Summary		Update resource status by UUID
//	@Description	update resource status by uuid
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string			true	"Resource UUID"
//	@Param			resource	body		models.Resource	true	"Resource info"
//	@Success		200			{object}	string			"Resource updated"
//	@Failure		400			{object}	string			"Resource UUID is required"
//	@Failure		404			{object}	string			"Can not find Resource to update"
//	@Router			/jobmanager/resources/status [put]
func (server *Server) UpdateResourceStateByUUID(w http.ResponseWriter, r *http.Request) {
	resourceBody, err := io.ReadAll(r.Body)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	logs.Logger.Println("Resource contents: " + string(resourceBody))

	updatedResource, err := server.ResourceService.UpdateResourceState(resourceBody)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		if err.Error() == "not found" {
			responses.ERROR(w, http.StatusNotFound, err)
		} else {
			responses.ERROR(w, http.StatusBadRequest, err)
		}
		return
	}

	responses.JSON(w, http.StatusOK, updatedResource)
}

func (server *Server) CreateResource(w http.ResponseWriter, r *http.Request) {
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

	// gorm save
	_, err = server.ResourceService.SaveResource(&resource)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, resource)
}
