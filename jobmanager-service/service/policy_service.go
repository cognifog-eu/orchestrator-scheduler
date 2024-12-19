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
package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"etsn/server/jobmanager-service/models"
	"etsn/server/jobmanager-service/repository"
	"etsn/server/jobmanager-service/utils/logs"
	"fmt"
	"net/http"

	"moul.io/http2curl"
)

type PolicyService interface {
	HandlePolicyIncompliance(incomplianceBody []byte, header http.Header) (*models.Remediation, error)
	NotifyPolicyManager(manifest string, jobGroup *models.JobGroup, token string) error
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// PolicyService struct implements the PolicyService interface
type policyService struct {
	policyRepository repository.PolicyRepository
	jobService       JobService
	httpClient       HTTPClient
}

// NewPolicyService returns a new instance of policyService
func NewPolicyService(policyRepository repository.PolicyRepository, jobService JobService, httpClient HTTPClient) PolicyService {
	return &policyService{policyRepository: policyRepository, jobService: jobService, httpClient: httpClient}
}

// HandlePolicyIncompliance processes incompliance and applies remediation
func (s *policyService) HandlePolicyIncompliance(incomplianceBody []byte, header http.Header) (*models.Remediation, error) {
	// We use an anonymous struct because we don't need all fields from the incompliance request
	var parsedData struct {
		ID          string `json:"id"`
		Remediation string `json:"remediation"`
		ExtraLabels struct {
			Container string `json:"k8s_container_name"`
			PodUID    string `json:"k8s_pod_uid"`
			Pod       string `json:"k8s_pod_name"`
			Node      string `json:"k8s_node_name"`
			Namespace string `json:"k8s_namespace_name"`
			Command   string `json:"command"`
		} `json:"extraLabels"`
		Subject struct {
			Manifest string `json:"manifest"`
		} `json:"subject"`
	}

	err := json.Unmarshal([]byte(incomplianceBody), &parsedData)
	if err != nil {
		return nil, err
	}

	remediation := models.Remediation{
		RemediationType: models.RemediationType(parsedData.Remediation),
		Status:          models.Pending,
		ResourceID:      parsedData.Subject.Manifest,
	}

	if parsedData.Remediation == "secure" {
		remediation.RemediationTarget = &models.RemediationTarget{

			Container: parsedData.ExtraLabels.Container,
			PodUID:    parsedData.ExtraLabels.PodUID,
			Pod:       parsedData.ExtraLabels.Pod,
			Node:      parsedData.ExtraLabels.Node,
			Namespace: parsedData.ExtraLabels.Namespace,
			Command:   parsedData.ExtraLabels.Command,
		}
	}

	// Retrieve the job related to the incompliance
	jobGotten, err := s.jobService.FindJobByResourceUUID(remediation.ResourceID)
	if err != nil {
		return nil, err
	}
	logs.Logger.Println("job found " + jobGotten.ID)

	// Validate job for remediation
	if jobGotten.OwnerID == "" {
		return nil, errors.New("job OwnerID cannot be empty")
	}
	if jobGotten.State != models.Finished {
		return nil, fmt.Errorf("job cannot be remediated; expected state 'Finished', got '%s'", jobGotten.State)
	}
	if jobGotten.Resource == nil {
		return nil, errors.New("job resource is nil")
	}
	// Use the new method to update the job
	err = s.jobService.UpdateJobForRemediation(jobGotten, &remediation)
	if err != nil {
		return nil, err
	}

	return &remediation, nil
}

func (s *policyService) NotifyPolicyManager(manifest string, jobGroup *models.JobGroup, token string) error {
	notification := models.Notification{
		AppInstance: jobGroup.ID,
		Service:     "job-manager",
		CommonAction: models.Action{
			URI:                "/jobmanager/policies/incompliance/create",
			HTTPMethod:         "POST",
			IncludeAccessToken: true,
		},
	}
	bodyBytes, err := json.Marshal(notification)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return err
	}

	req, err := http.NewRequest("POST", models.PolicyManagerBaseURL+"/polman/registry/api/v1/etsn/", bytes.NewBuffer(bodyBytes))
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logs.Logger.Println("ERROR " + err.Error())
		return err
	}
	defer resp.Body.Close()

	command, _ := http2curl.GetCurlCommand(req)
	logs.Logger.Println("Request sent to Policy Manager Service: ")
	logs.Logger.Println(command)
	logs.Logger.Println("End Policy Manager Request.")

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	} else {
		err := errors.New("Bad response from Policy Manager: status code - " + string(rune(resp.StatusCode)))
		return err
	}
}
