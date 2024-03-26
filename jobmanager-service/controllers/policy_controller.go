package controllers

import (
	"encoding/json"
	"etsn/server/jobmanager-service/models"
	"etsn/server/jobmanager-service/responses"
	"etsn/server/jobmanager-service/utils/logs"
	"io"
	"net/http"
)

// CreatePolicyIncompliance example
//
//	@Description	create new Incompliance
//	@ID				create-new-incompliance
//	@Accept			plain
//	@Produce		json
//	@Param			Authorization	header		string		true	"Authentication header"
//	@Param			application		body		string		true	"Incompliance Object"
//	@Success		200				{object}	models.Job	"Ok"
//	@Failure		400				{object}	string		"Incompliance Object is not correct"
//	@Router			/jobmanager/policies/incompliance/create [post]
func (server *Server) CreatePolicyIncompliance(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// stringID := vars["id"]
	// if stringID == "" {
	// 	err := errors.New("ID Cannot be empty")
	// 	responses.ERROR(w, http.StatusBadRequest, err)
	// 	return
	// }
	// uuid, err := uuid.Parse(stringID)
	incompliance := models.Incompliance{}
	incomplianceBody, err := io.ReadAll(r.Body)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	// parse to incompliance objects
	err = json.Unmarshal(incomplianceBody, &incompliance)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// gorm save
	_, err = incompliance.SaveIncompliance(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	responses.JSON(w, http.StatusOK, incompliance)
}
