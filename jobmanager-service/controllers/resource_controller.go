package controllers

import (
	"encoding/json"
	"errors"
	"etsn/server/jobmanager-service/models"
	"etsn/server/jobmanager-service/responses"
	"etsn/server/jobmanager-service/utils/logs"
	"io"
	"net/http"

	uuid "github.com/google/uuid"
	"github.com/gorilla/mux"
)

// var (
// 	deployManagerBaseURL = os.Getenv("DEPLOY_MANAGER_BASE_URL")
// )

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

// GetResourceStateByJobUUID example
//
//	@Description	get resource status by job uuid
//	@ID				get-resource-status-by-job-uuid
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string			true	"Authentication header"
//	@Param			job_id			path		string			true	"Job ID"
//	@Success		200				{object}	models.Resource	"Ok"
//	@Failure		400				{object}	string			"Job ID is required"
//	@Failure		404				{object}	string			"Can not find Job by ID"
//	@Router			/jobmanager/resources/status/{job_id} [get]
func (server *Server) GetResourceStateByJobUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["job_id"]
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

	// retrieve info from the job first. need extra info!
	resource := models.Resource{}
	resourceGotten, err := resource.FindResourceByUUID(server.DB, uuid)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	// ONLY VALID IN PUSH MODE
	// resourceStatus := models.Resource{}
	// jobString, err := json.Marshal(jobGotten)
	// if err != nil {
	// 	logs.Logger.Println("ERROR " + err.Error())
	// 	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	// 	return
	// }
	// logs.Logger.Println("Job found: " + string(jobString))
	// this can have some sort of cache to avoid making requests to DM

	// // create DM request
	// req, err := http.NewRequest("GET", deployManagerBaseURL+"/deploy-manager/resource", bytes.NewBuffer([]byte{}))
	// query := req.URL.Query()
	// query.Add("uuid", uuid.String())
	// query.Add("node_target", jobGotten.Targets[0].NodeName)
	// query.Add("resource_name", jobGotten.Resource.ResourceName)
	// query.Encode()
	// logs.Logger.Println("request status to DM: " + req.URL.String())
	// if err != nil {
	// 	logs.Logger.Println("ERROR " + err.Error())
	// 	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	// 	return
	// }

	// // forward the authorization token
	// req.Header.Add("Authorization", r.Header.Get("Authorization"))

	// // // do request
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	logs.Logger.Println("ERROR " + err.Error())
	// 	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	// 	return
	// }
	// defer resp.Body.Close()

	// direct body read
	// resourceStatusBody, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	logs.Logger.Println("ERROR " + err.Error())
	// 	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	// 	return
	// }

	// // parse to application objects
	// err = json.Unmarshal(resourceStatusBody, &resourceStatus)
	// if err != nil {
	// 	logs.Logger.Println("ERROR " + err.Error())
	// 	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	// 	return
	// }

	responses.JSON(w, http.StatusOK, resourceGotten.Conditions)

}

// UpdateResourceStateByUUID example
//
//	@Description	update a resource status by uuid
//	@ID				update-a-resource-status-by-uuid
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string			true	"Authentication header"
//	@Param			id				path		string			true	"Resource UUID"
//	@Param			resource		body		models.Resource	true	"Resource info"
//	@Success		200				{object}	string			"Ok"
//	@Failure		400				{object}	string			"ID is required"
//	@Failure		404				{object}	string			"Can not find Resource to update"
//	@Router			/jobmanager/resources/status/{id} [put]
func (server *Server) UpdateResourceStateByUUID(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// stringID := vars["id"]
	// if stringID == "" {
	// 	err := errors.New("ID Cannot be empty")
	// 	responses.ERROR(w, http.StatusBadRequest, err)
	// 	return
	// }
	// uuid, err := uuid.Parse(stringID)
	resource := models.Resource{}
	resourceBody, err := io.ReadAll(r.Body)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	logs.Logger.Println("Resource contents: " + string(resourceBody))
	// parse to application objects
	err = json.Unmarshal(resourceBody, &resource)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// get resource from db, retrieve the job first
	job := models.Job{}
	jobGotten, err := job.FindJobByResourceUUID(server.DB, resource.ResourceUUID)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	// debug
	j, err := json.Marshal(jobGotten)
	if err != nil {
		logs.Logger.Println("ERROR during debug" + err.Error())
	}
	logs.Logger.Println("Job contents: " + string(j))
	// update resource details
	// swap the ids.. TODO improve
	resource.ResourceUUID = jobGotten.UUID
	resource.ID = jobGotten.Resource.ID
	resource.JobID = jobGotten.ID
	logs.Logger.Println("Updating Resource Status, Resource ID: " + resource.ID.String())
	// updatedResource, err := resource.UpdateAResource(server.DB, job.ID, resource.ResourceUUID)
	resource.RemoveConditions(server.DB)
	for _, condition := range resource.Conditions {
		condition.ResourceID = resource.ID
		_, err = resource.AddCondition(server.DB, &condition)
		if err != nil {
			logs.Logger.Println("Resource were not found during status update")
			responses.ERROR(w, http.StatusBadRequest, err)
			return
		}
	}
	responses.JSON(w, http.StatusOK, resource)
}

func (server *Server) CreateResource(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// stringID := vars["id"]
	// if stringID == "" {
	// 	err := errors.New("ID Cannot be empty")
	// 	responses.ERROR(w, http.StatusBadRequest, err)
	// 	return
	// }
	// uuid, err := uuid.Parse(stringID)
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
	_, err = resource.SaveResource(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// update resource details first
	// logs.Logger.Println("Updating Resource Status: " + uuid.String())
	// resource.RemoveConditions(server.DB)
	// for _, condition := range resource.Conditions {
	// 	condition.ResourceID = resource.ID
	// 	_, err = resource.AddCondition(server.DB, &condition)
	// 	if err != nil {
	// 		logs.Logger.Println("Resource were not found during status update")
	// 		responses.ERROR(w, http.StatusBadRequest, err)
	// 		return
	// 	}
	// }
	responses.JSON(w, http.StatusOK, resource)
}
