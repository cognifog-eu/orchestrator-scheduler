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
	"etsn/server/jobmanager-service/responses"
	"etsn/server/jobmanager-service/utils/logs"
	"io"
	"net/http"
)

// CreatePolicyIncompliance godoc
//
//	@Summary		Triger a new remediation
//	@Description	Triger a new remediation.
//	@Tags			policies
//	@Accept			json
//	@Produce		json
//	@Param			application	body		string	true	"Remediation Object"
//	@Success		200			{object}	models.Remediation
//	@Failure		400			{object}	string	"Remediation Object is not correct"
//	@Failure		422			{object}	string	"Unprocessable Entity"
//	@Router			/jobmanager/policies/incompliance [post]
func (server *Server) CreatePolicyIncompliance(w http.ResponseWriter, r *http.Request) {
	incomplianceBody, err := io.ReadAll(r.Body)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Handle the incompliance through the service
	incompliance, err := server.PolicyService.HandlePolicyIncompliance(incomplianceBody, r.Header)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, incompliance)
}
