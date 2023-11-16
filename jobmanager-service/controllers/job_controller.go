package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/responses"
	"icos/server/jobmanager-service/utils/logs"
	"io"
	"net/http"
	"os"
	"strings"

	uuid "github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
	lighthouseBaseURL  = os.Getenv("LIGHTHOUSE_BASE_URL")
	apiV3              = "/api/v3"
	matchmackerBaseURL = os.Getenv("MATCHMAKING_URL")
)

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

func (server *Server) GetJobByUUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stringID := vars["id"]
	if stringID == "" {
		err := errors.New("ID Cannot be empty")
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uuid, err := uuid.Parse(stringID)

	// gorm retrieve
	job := models.Job{}
	jobGotten, err := job.FindJobByUUID(server.DB, uuid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGotten)

}

// retrieves only executable jobs for now
func (server *Server) GetJobsByState(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// state, err := strconv.ParseInt(vars["state"], 10, 32)
	// // state validation
	// if !models.StateIsValid(int(state)) {
	// 	responses.ERROR(w, http.StatusBadRequest, err)
	// 	return
	// }

	// gorm retrieve
	job := models.Job{}
	// retrieves jobs that are created && not locked or progressing && locked for more than a minute
	jobGotten, err := job.FindJobsToExecute(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobGotten)
}

func (server *Server) CreateJob(w http.ResponseWriter, r *http.Request) {

	// receive manifest and unmarshall it to Manifest object
	// jobTemp := models.Job{}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	bodyString := string(bodyBytes)
	bodyStringTrimmed := strings.Trim(bodyString, "\r\n")
	logs.Logger.Println("Trimmed body: " + bodyStringTrimmed)

	// err = json.Unmarshal(body, &jobTemp.Manifest)
	// if err != nil {
	// 	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	// 	return
	// }

	// validate job -> if unmarshalled without error = OK
	// matchmaking + optimization = targets -> sync?
	var targets []models.Target

	// MM Mock
	targets = append(targets, models.Target{
		ClusterName: "k3s-worker1",
		NodeName:    "k3s-worker1",
	})

	// // create MM request
	// req, err := http.NewRequest("GET", matchmackerBaseURL, bytes.NewBuffer([]byte{}))
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
	// // logger.Info("Rancher response is: " + resp.Status)
	// if err != nil {
	// 	logs.Logger.Println("ERROR " + err.Error())
	// 	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	// 	return
	// }
	// defer resp.Body.Close()

	// // direct body read
	// bodyMM, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	logs.Logger.Println("ERROR " + err.Error())
	// 	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	// 	return
	// }

	// // parse to application objects
	// err = json.Unmarshal(bodyMM, &targets)
	// if err != nil {
	// 	logs.Logger.Println("ERROR " + err.Error())
	// 	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	// 	return
	// }

	// append targets to jobs app description
	// gorm save
	job := models.Job{
		Type:     models.CreateDeployment,
		State:    models.Created,
		Manifest: bodyStringTrimmed,
		Targets:  targets,
	}
	jobCreated, err := job.SaveJob(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobCreated)

}

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
	// gorm retrieve
	job := models.Job{}
	jobDeleted, err := job.DeleteAJob(server.DB, uuid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobDeleted)
}

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

	jobUpdated, err := job.UpdateAJob(server.DB, id)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, jobUpdated)
}
