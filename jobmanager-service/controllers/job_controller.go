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
	"etsn/server/jobmanager-service/utils/logs"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	uuid "github.com/google/uuid"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
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
	jobGotten, err := job.FindJobsToExecute(server.DB, orch)
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
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	bodyString := string(bodyBytes)
	bodyStringTrimmed := strings.Trim(bodyString, "\r\n")
	logs.Logger.Println("Trimmed body: " + bodyStringTrimmed)
	// var mMResponseMapper models.MMResponseMapper

	splittedManifests := strings.Split(bodyStringTrimmed, "---")

	// extract header block
	headers := splittedManifests[0]
	headerStruct := models.JobGroupHeader{}
	yaml.Unmarshal([]byte(headers), &headerStruct)

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

		// mocking MM response for now
		// start mock
		mMResponseJson := `{
			"components": [
				{
					"name": "consumer",
					"type": "kubernetes",
					"manifests": [
						{
							"name": "mjpeg"
						},
						{
							"name": "mjpeg-service"
						}
					],
					"targets": [
						{
							"cluster_name": "nuvlabox/55c7953e-2aa0-4d18-834c-b4d76d824bb9",
							"node_name": "john-rasbpi-5-1",
							"orchestrator": "nuvla"
						}
					]
				},
				{
					"name": "producer",
					"type": "kubernetes",
					"manifests": [
						"video-streaming-service",
						"video-streaming-deployment"
					],
					"targets": [
						{
							"cluster_name": "icos-cluster-2a",
							"node_name": "sim-k8s-master",
							"orchestrator": "ocm"
						}
					]
				}
			]
		}`

		bodyMM = []byte(mMResponseJson)
		// end mock

		dst := &bytes.Buffer{}
		_ = json.Indent(dst, bodyMM, "", "  ")
		logs.Logger.Println("MM response is: " + dst.String())
		err = json.Unmarshal(bodyMM, &headerStruct)
		if err != nil {
			logs.Logger.Println("ERROR " + err.Error())
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		logs.Logger.Printf("Matchmaking response details: %#v", headerStruct)
	*/

	// MM Mock from env
	mockTargetCluster := os.Getenv("MOCK_TARGET_CLUSTER")
	logs.Logger.Println("Mocking MM response: " + mockTargetCluster)
	if mockTargetCluster == "" {
		logs.Logger.Println("Warning: MOCK_TARGET_CLUSTER environment variable is not set")
	}
	for i := range headerStruct.Components {
		component := &headerStruct.Components[i]
		component.Targets = append(component.Targets, models.Target{
			ClusterName:  mockTargetCluster,
			Orchestrator: "ocm",
		})
	}

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
			Type:               "Created",
			Status:             "True",
			ObservedGeneration: 1,
			LastTransitionTime: time.Now(),
			Reason:             "AwaitingForExecution",
			Message:            "Waiting an Orchestrator to take the Job",
		},
	}

	// create Job Group
	jobGroup := models.JobGroup{
		AppName:        headerStruct.Name,
		AppDescription: headerStruct.Description,
	}

	// if no namespace is provided, create a unique one, in the future will be always generated
	if headerStruct.Namespace == "" {
		uuidNamespace := uuid.New().String()
		headerStruct.Namespace = uuidNamespace
	}

	// temporal fix until the following issues are resolved:
	// https://production.eng.it/gitlab/icos/meta-kernel/job-manager/-/issues/21
	// https://production.eng.it/gitlab/icos/meta-kernel/job-manager/-/issues/22

	// create job per component
	for i, comp := range headerStruct.Components {
		// declare job per received component
		job := models.Job{
			Type:         models.CreateDeployment,
			State:        models.JobCreated,
			Manifest:     splittedManifests[i+1], // TODO, improve
			Targets:      comp.Targets,
			JobGroupName: jobGroup.AppName,
			Orchestrator: comp.Targets[0].Orchestrator, // TODO, should not point to first element
			Namespace:    headerStruct.Namespace,       // Unique within a single cluster
			Resource: models.Resource{
				ResourceName: comp.Name,
				Conditions:   conditions,
			},
		}

		// populate manifests slice for each job
		// skipping the first element since its the header
		for _, manifest := range splittedManifests[1:] {
			for _, headerCompManifest := range comp.Manifests {
				var k8sManifest *models.K8sMapper
				err = yaml.Unmarshal([]byte(manifest), &k8sManifest)
				if err != nil {
					logs.Logger.Println("ERROR during job manifests population " + err.Error())
				}
				if headerCompManifest.Name == k8sManifest.Metadata.Name {
					logs.Logger.Println("Manifest to be populated: " + headerCompManifest.Name)
					job.Manifests = append(job.Manifests, models.PlainManifests{
						YamlString: manifest,
					})
				}
			}
		}

		jobGroup.Jobs = append(jobGroup.Jobs, job)
		logs.Logger.Println("New Job appended to JobGroup: " + job.JobGroupID.String())
	}
	// gorm save
	_, err = jobGroup.SaveJobGroup(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// notify policy manager
	models.NotifyPolicyManager(server.DB, bodyStringTrimmed, jobGroup, r.Header.Get("Authorization"))
	responses.JSON(w, http.StatusCreated, jobGroup)
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
		logs.Logger.Println("JOB's ID is empty!")
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
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	jobDeleted, err := jobGotten.DeleteAJob(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusServiceUnavailable, err)
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
	jobUpdated, err := job.UpdateAJob(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobUpdated)
}

// LockJobByUUID example
//
//	@Description	Get LockJobByUUID
//	@ID				get-lockjobbyuuid
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication header"
//	@Success		200				{string}	string	"Ok"
//	@Failure		404				{object}	string				"Can not find Job"
//	@Failure		403				{object}	string				"Forbidden"
//	@Router			/jobmanager/jobs/lock/ [Patch]
func (server *Server) LockJobByUUID(w http.ResponseWriter, r *http.Request) {
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
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	b := new(bool)
	*b = true
	// job already locked.
	if jobGotten.Locker == b {
		// job has been locked for less than timeout. cannot be locked again
		timeOutThreshold := time.Now().Local().Add(time.Second * time.Duration(-300))
		if jobGotten.UpdatedAt.After(timeOutThreshold) {
			responses.ERROR(w, http.StatusForbidden, err)
			return
		}
	}
	// job already finished, cannot be locked
	if jobGotten.State == models.JobFinished {
		responses.ERROR(w, http.StatusForbidden, err)
		return
	}

	// TODO: only the owner and JobManager itself can lock/unlock jobs

	// if got to this step, job was not locked(executable) as expected
	// if the job was not locked it was at state=JobCreated or state=JobDegraded
	// locking the job, promotion to owner happens here
	jobGotten.Locker = b
	//gorm update
	_, err = jobGotten.JobLocker(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	// job locked and can be now executed
	responses.JSON(w, http.StatusNoContent, http.NoBody)
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
	logs.Logger.Println("Parsed JobGroup ID to lookup: " + uuid.String())
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
		logs.Logger.Println("JobGroup's ID is empty!")
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

// GetAllJobGroups example
//
//	@Description	Get All JobGroups
//	@ID				get-all-jobGroups
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string			true	"Authentication header"
//	@Success		200				{string}	[]models.JobGroup	"Ok"
//	@Failure		404				{object}	string			"Can not find JobGroups"
//	@Router			/jobmanager/jobs/group/all [get]
func (server *Server) GetAllJobGroups(w http.ResponseWriter, r *http.Request) {
	// gorm retrieve
	jobGroup := models.JobGroup{}
	jobGroupsGotten, err := jobGroup.FindAllJobGroups(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGroupsGotten)
}

// UndeployJobGroupByUUID example
//
//	@Description	Undeploy JobGroup
//	@ID				undeploy-jobgroup
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Authentication header"
//	@Success		200				{string}	[]models.JobGroup	"Ok"
//	@Failure		404				{object}	string				"Can not find JobGroup"
//	@Router			/jobmanager/jobs [get]
func (server *Server) UndeployJobGroupByUUID(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println("Parsed JobGroup ID to lookup: " + uuid.String())
	// gorm retrieve
	jobGroup := models.JobGroup{}
	jobGroupGotten, err := jobGroup.FindJobGroupByUUID(server.DB, uuid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// logic
	// check jobs are actually deployed, if resources not found: TODO
	for i := len(jobGroupGotten.Jobs) - 1; i >= 0; i-- {
		job := jobGroupGotten.Jobs[i]
		// job must  be already deployed or progressing to do so
		if job.State != models.JobCreated {
			// resource exists
			if job.Resource.ID.String() != "00000000-0000-0000-0000-000000000000" {
				// schedule the jobs for execution
				job.State = models.JobCreated
				job.Type = models.DeleteDeployment
				// check jobs are locked, they should be, now I must unlock them for undeployment
				if *job.Locker {
					// unlock the job
					*job.Locker = false
				} else {
					responses.ERROR(w, http.StatusBadRequest, err)
					return
				}
				// no need to schedule, just delete the job
			} else {
				_, err := job.DeleteAJob(server.DB)
				if err != nil {
					responses.ERROR(w, http.StatusNotFound, err)
					return
				}
				// Remove element at index i without affecting the loop
				jobGroupGotten.Jobs = append(jobGroupGotten.Jobs[:i], jobGroupGotten.Jobs[i+1:]...)
			}

		}
	}
	// gorm update
	_, err = jobGroupGotten.UpdateJobGroupState(server.DB, uuid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGroupGotten)
}
