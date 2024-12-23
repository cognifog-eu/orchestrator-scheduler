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
	"etsn/server/jobmanager-service/models"
	"etsn/server/jobmanager-service/utils/logs"
	"fmt"
	"io"
	"log"
	"net/http"

	"gopkg.in/yaml.v3"
)

type AllocatorService interface {
	Allocate(allocationResponseDTO *models.AllocationDTO, opts ...models.AllocationOptionDTO) error
	AssignTargets(job *models.Job, targets interface{}) error
}

type allocatorService struct{}

func NewAllocatorService() AllocatorService {
	return &allocatorService{}
}

func (as *allocatorService) Allocate(allocationResponseDTO *models.AllocationDTO, opts ...models.AllocationOptionDTO) error {

	for _, opt := range opts {
		opt(allocationResponseDTO)
	}

	return nil
}

func (as *allocatorService) AssignTargets(job *models.Job, targets interface{}) error {
	switch t := targets.(type) {
	case map[string]interface{}:
		targetBytes, err := json.Marshal(t)
		if err != nil {
			logs.Logger.Println("ERROR " + err.Error())
			return err
		}

		var targetStruct models.Target
		err = json.Unmarshal(targetBytes, &targetStruct)
		if err != nil {
			logs.Logger.Println("ERROR " + err.Error())
			return err
		}

		job.Target = targetStruct
		job.Orchestrator = targetStruct.Orchestrator

	case []interface{}:
		if len(t) == 0 {
			job.Target = models.Target{}
		} else {
			logs.Logger.Fatalf("Unexpected non-empty array for targets")
		}

	case interface{}:
		targetBytes, err := yaml.Marshal(t)
		if err != nil {
			logs.Logger.Println("ERROR " + err.Error())
			return err
		}

		var targetStruct models.Target
		err = yaml.Unmarshal(targetBytes, &targetStruct)
		if err != nil {
			logs.Logger.Println("ERROR " + err.Error())
			return err
		}

		job.Target = targetStruct

	default:
		logs.Logger.Fatalf("Unexpected type for targets: %T", targets)
	}

	return nil
}

// Option used to set the allocation request body when creating a new deployment
func WithRawYamlBody(rawYamlBody []byte, header http.Header) models.AllocationOptionDTO {
	return func(allocOutput *models.AllocationDTO) {
		err := allocationRequest(rawYamlBody, header, allocOutput)
		logs.Logger.Printf("allocation body: %s", string(rawYamlBody))
		if err != nil {
			fmt.Println("Error in AllocationRequest:", err)
			return
		}

		logs.Logger.Printf("AllocationRequest output: %#v", allocOutput)
	}
}

func allocationRequest(bodyBytes []byte, header http.Header, allocOutput *models.AllocationDTO) error {

	req, err := http.NewRequest("POST", models.MatchmakerBaseURL+"/matchmake", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-yaml")
	if header != nil {
		req.Header.Add("Authorization", header.Get("Authorization"))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyMM, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	dst := &bytes.Buffer{}
	if err := json.Indent(dst, bodyMM, "", "  "); err != nil {
		return err
	}

	logs.Logger.Printf("AllocationRequest response: %s", dst.String())

	if err := json.Unmarshal(bodyMM, &allocOutput); err != nil {
		return err
	}

	logs.Logger.Printf("AllocationRequest output: %#v", allocOutput)

	return nil
}

func WithJobGroup(jg *models.JobGroup, header http.Header) models.AllocationOptionDTO {
	return func(allocOutput *models.AllocationDTO) {
		// TODO reconstruct the YAML app descriptor from persisted job group
		rawYamlBody, err := yaml.Marshal(createAllocationBody(jg))
		logs.Logger.Printf("allocation body: %s", string(rawYamlBody))
		if err != nil {
			logs.Logger.Println("Error marshaling job group header:", err)
			return
		}

		err = allocationRequest(rawYamlBody, header, allocOutput)
		if err != nil {
			logs.Logger.Println("Error in AllocationRequest:", err)
			return
		}
	}
}

// createAllocationBody converts a JobGroup into an AllocationDTO.
func createAllocationBody(jg *models.JobGroup) *models.AllocationDTO {
	var components []models.ComponentDTO
	var allPlainYamlManifests []models.Content

	for _, job := range jg.Jobs {
		var manifestRefsDTO []models.ManifestRefDTO

		for _, content := range job.Instruction.Contents {
			manifestRefsDTO = append(manifestRefsDTO, models.ManifestRefDTO{Name: content.Name})
			allPlainYamlManifests = append(allPlainYamlManifests, content)
		}

		components = append(components, models.ComponentDTO{
			Name: job.Instruction.ComponentName,
			Type: job.Instruction.Type,
			Requirement: models.Requirement{
				CPU:          job.Instruction.Requirement.CPU,
				Memory:       job.Instruction.Requirement.Memory,
				Device:       job.Instruction.Requirement.Device,
				Architecture: job.Instruction.Requirement.Architecture,
			},
			// Policies:  job.Component.Policies, // we need to create a DTO for this. Does MM take policies into account?
			Manifests: manifestRefsDTO,
			// Target:    job.Target,
		})
	}

	return &models.AllocationDTO{
		Name:        jg.AppName,
		Description: jg.AppDescription,
		Components:  components,
		Manifests:   plainYAMLtoJSONManifest(allPlainYamlManifests),
	}
}

// plainYAMLtoJSONManifest converts a slice of PlainManifest to JSON-like maps.
func plainYAMLtoJSONManifest(manifests []models.Content) []map[string]interface{} {
	var allManifests []map[string]interface{}

	for _, manifest := range manifests {
		manifestMap := make(map[string]interface{})
		err := yaml.Unmarshal([]byte(manifest.Yaml), &manifestMap)
		if err != nil {
			log.Printf("Error unmarshalling YAML: %v", err)
			continue
		}
		allManifests = append(allManifests, manifestMap)
	}

	return allManifests
}
