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
	"encoding/json"
	"errors"
	"etsn/server/jobmanager-service/models"
	"etsn/server/jobmanager-service/repository"
	"etsn/server/jobmanager-service/utils/logs"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"
)

// JobGroupService interface defines the methods for job group operations
type JobGroupService interface {
	CreateJobGroup(bodyBytes []byte, header http.Header) (*models.JobGroup, error)
	UpdateJobGroup(*models.JobGroup) (*models.JobGroup, error)
	ReplaceJobGroup(bodyJob []byte) (*models.JobGroup, error)
	AllocateJobGroup(stringID string, header http.Header) (*models.JobGroup, *models.AllocationDTO, error)
	FindJobGroupByUUID(string) (*models.JobGroup, error)
	FindAllJobGroups() (*[]models.JobGroup, error)
	DeleteJobGroupByID(id string) (*models.JobGroup, error)
	StopJobGroupByID(stringID string) (*models.JobGroup, error)
	StartJobGroupByID(stringID string, header http.Header) (*models.JobGroup, error)
}

// jobGroupService struct implements the JobGroupService interface
type jobGroupService struct {
	repo             repository.JobGroupRepository
	allocatorService AllocatorService
	manifestService  ManifestService
}

// NewJobGroupService returns a new instance of jobGroupService
func NewJobGroupService(repo repository.JobGroupRepository, allocatorService AllocatorService, manifestService ManifestService) JobGroupService {
	return &jobGroupService{
		repo:             repo,
		allocatorService: allocatorService,
		manifestService:  manifestService}
}

func (s *jobGroupService) CreateJobGroup(bodyBytes []byte, header http.Header) (*models.JobGroup, error) {
	bodyString := string(bodyBytes)
	bodyStringTrimmed := strings.Trim(bodyString, "\r\n")
	logs.Logger.Println("Trimmed body: " + bodyStringTrimmed)

	var allocResDTO models.AllocationDTO
	err := yaml.Unmarshal([]byte(bodyStringTrimmed), &allocResDTO)
	if err != nil {
		return nil, err
	}

	// request matchmaking service for components allocation
	err = s.allocatorService.Allocate(&allocResDTO, WithRawYamlBody(bodyBytes, header))
	if err != nil {
		return nil, err
	}

	jobGroup := models.CreateJobGroupFromAllocation(&allocResDTO)
	logs.Logger.Printf("Components in applicationDescriptor: %#v", allocResDTO.Components)

	// Before loop
	if len(allocResDTO.Components) == 0 {
		logs.Logger.Println("No components found in applicationDescriptor")
	}

	for _, comp := range allocResDTO.Components {

		contentList, err := s.manifestService.ProcessManifests(comp, allocResDTO.Manifests)
		if err != nil {
			logs.Logger.Println("Error processing manifests: " + err.Error())
			return nil, err
		}
		instruction := models.Instruction{
			InstructionBase: models.InstructionBase{
				ComponentName: comp.Name,
				Type:          comp.Type,
				Requirement:   comp.Requirement,
				Policies:      comp.Policies,
			},
			//Manifests: compManifests,
			Contents: contentList,
		}
		job := models.Job{
			Type:         models.CreateDeployment,
			State:        models.Created,
			Instruction:  &instruction,
			Namespace:    jobGroup.AppName,
			Orchestrator: "ocm",
		}

		s.allocatorService.AssignTargets(&job, comp.Target)

		jobGroup.Jobs = append(jobGroup.Jobs, job)
		logs.Logger.Printf("New Job appended to JobGroup: %#v", job)
	}

	logs.Logger.Printf("Final JobGroup: %#v", jobGroup)

	jobGroup, err = s.repo.SaveJobGroup(jobGroup)
	if err != nil {
		logs.Logger.Println("Error saving JobGroup " + err.Error())
		return nil, err
	}

	return jobGroup, nil
}

func (s *jobGroupService) StartJobGroupByID(stringID string, header http.Header) (*models.JobGroup, error) {

	jobGroupGotten, allocResDTO, err := s.AllocateJobGroup(stringID, header)
	if err != nil {
		return nil, err
	}
	// Process the allocation results and update the job group with the new job targets
	for i := range jobGroupGotten.Jobs {
		job := &jobGroupGotten.Jobs[i]
		logs.Logger.Printf("StartJobGroupByID: Processing Job %s in JobGroup %s", job.ID, stringID)

		// Reset job fields to prepare for execution
		job.State = models.Created
		job.Type = models.CreateDeployment
		// job.OwnerID = ""

		// Match the component in the job with the component in the allocation results
		for _, comp := range allocResDTO.Components {
			if job.Instruction.ComponentName == comp.Name {
				logs.Logger.Printf("StartJobGroupByID: Matching component found for Job %s - Component: %s", job.ID, comp.Name)
				s.allocatorService.AssignTargets(job, comp.Target)
			}
		}
	}

	// Update the job group
	jobGroupUpdated, err := s.UpdateJobGroup(jobGroupGotten)
	if err != nil {
		logs.Logger.Printf("StartJobGroupByID: Error updating JobGroup %s, error: %v", stringID, err)
		return nil, errors.New("error updating JobGroup")
	}

	//jobGroupUpdatedDTO := models.MapJobGroupToDTO(jobGroupUpdated)
	logs.Logger.Printf("StartJobGroupByID: Successfully updated JobGroup %s", stringID)

	return jobGroupUpdated, nil
}

func (s *jobGroupService) AllocateJobGroup(stringID string, header http.Header) (*models.JobGroup, *models.AllocationDTO, error) {
	jobGroupGotten, err := s.FindJobGroupByUUID(stringID)
	if err != nil {
		logs.Logger.Printf("AllocateJobGroup: JobGroup not found for ID %s, error: %v", stringID, err)
		return nil, nil, errors.New("jobgroup not found")
	}

	var allocResDTO models.AllocationDTO

	// Request matchmaking service for components allocation
	err = s.allocatorService.Allocate(&allocResDTO, WithJobGroup(jobGroupGotten, header))
	if err != nil {
		logs.Logger.Printf("AllocateJobGroup: Allocation failed for JobGroup %s, error: %v", stringID, err)
		return nil, nil, err
	}

	return jobGroupGotten, &allocResDTO, nil
}

// UpdateJobGroup updates an existing job group
func (s *jobGroupService) ReplaceJobGroup(bodyJob []byte) (*models.JobGroup, error) {
	var jobGroupUpdate models.JobGroup
	logs.Logger.Println("Starting UpdateJobGroup process")

	// Unmarshal the input request body
	if err := json.Unmarshal(bodyJob, &jobGroupUpdate); err != nil {
		logs.Logger.Printf("Error unmarshaling request body: %v. Body: %s", err, string(bodyJob))
		return nil, err
	}

	// Find the existing job group by UUID
	existingJobGroup, err := s.FindJobGroupByUUID(jobGroupUpdate.ID)
	if err != nil {
		logs.Logger.Printf("Error finding job group by UUID: %v. JobGroup ID: %s", err, jobGroupUpdate.ID)
		return nil, err
	}

	// Update the existing job group with new values
	existingJobGroup.AppName = jobGroupUpdate.AppName
	existingJobGroup.AppDescription = jobGroupUpdate.AppDescription

	// Handle job updates
	if len(jobGroupUpdate.Jobs) > 0 {
		jobMap := make(map[string]models.Job)
		for _, updatedJob := range jobGroupUpdate.Jobs {
			jobMap[updatedJob.ID] = updatedJob
		}

		for i := range existingJobGroup.Jobs {
			job := &existingJobGroup.Jobs[i]
			if updatedJob, ok := jobMap[job.ID]; ok {
				*job = updatedJob
			}
		}
	}

	// Update job states and types
	for i := range existingJobGroup.Jobs {
		job := &existingJobGroup.Jobs[i]
		job.State = models.Created

		if job.OwnerID != "" {
			job.Type = models.UpdateDeployment
			job.SubType = models.Replace // This could be improved with a more dynamic logic
		} else {
			job.Type = models.CreateDeployment
		}

	}

	// Update the job group in the repository
	jobGroupUpdated, err := s.UpdateJobGroup(existingJobGroup)
	if err != nil {
		logs.Logger.Printf("Error updating job group in repository: %v. JobGroup ID: %s", err, existingJobGroup.ID)
		return nil, err
	}

	// Map the updated job group to DTO
	//jobGroupUpdatedDTO := models.MapJobGroupToDTO(jobGroupUpdated)
	logs.Logger.Printf("Successfully mapped updated job group to DTO: ID=%s", jobGroupUpdated.ID)

	return jobGroupUpdated, nil
}

// DeleteJobGroup deletes a job group
func (s *jobGroupService) DeleteJobGroupByID(id string) (*models.JobGroup, error) {
	if id == "" {
		err := errors.New("ID Cannot be empty")
		logs.Logger.Println("JobGroup's ID is empty!")
		return nil, err
	}

	// Validate job group can be deleted
	jobGroupGotten, err := s.repo.FindJobGroupByUUID(id)
	if err != nil {
		return nil, err
	}

	for _, job := range jobGroupGotten.Jobs {
		logs.Logger.Printf("checking job with ID: %s, Type: %s, State: %s\n", job.ID, job.Type, job.State)
		// Job can only be deleted if it is of type DeleteDeployment
		if job.Type == models.DeleteDeployment {
			// Job has resources and it is not in JobFinished or JobCreated state
			if job.State != models.Finished && job.State != models.Created {
				err := errors.New("JobGroup cannot be deleted, one or more jobs are not in JobFinished state")
				logs.Logger.Println("JobGroup cannot be deleted, one or more jobs are not in JobFinished state")
				return nil, err
			}
		} else {
			err := errors.New("JobGroup cannot be deleted, one or more jobs are not of type DeleteDeployment")
			logs.Logger.Println("JobGroup cannot be deleted, one or more jobs are not of type DeleteDeployment")
			return nil, err
		}
	}

	// Delete the job group
	_, err = s.repo.DeleteJobGroup(id)
	if err != nil {
		return nil, err
	}

	return jobGroupGotten, nil
}

func (s *jobGroupService) StopJobGroupByID(stringID string) (*models.JobGroup, error) {
	// Validate the input ID
	if stringID == "" {
		return nil, errors.New("ID cannot be empty")
	}

	// Fetch the job group by UUID
	jobGroup, err := s.FindJobGroupByUUID(stringID)
	if err != nil {
		return nil, errors.New("job group not found")
	}

	// Validate that the job group has jobs
	if len(jobGroup.Jobs) == 0 {
		return nil, errors.New("job group has no jobs")
	}

	// Iterate through jobs to validate and update them
	for i := range jobGroup.Jobs {
		job := &jobGroup.Jobs[i]

		// Ensure the job is in the Finished state before proceeding
		if job.State != models.Finished {
			return nil, errors.New("job group cannot be stopped, one or more jobs are not in the finished state")
		}

		// Ensure that each job has an associated resource
		if job.Resource == nil {
			return nil, errors.New("job group cannot be stopped, one or more jobs have no associated resources")
		}

		// Reset job fields to prepare for deletion
		//job.OwnerID = ""
		job.State = models.Created
		job.Type = models.DeleteDeployment
	}

	// Update the job group in the repository
	updatedJobGroup, err := s.UpdateJobGroup(jobGroup)
	if err != nil {
		return nil, errors.New("error updating job group")
	}

	return updatedJobGroup, nil
}

// FindJobGroupByUUID finds a job group by its UUID
func (s *jobGroupService) FindJobGroupByUUID(id string) (*models.JobGroup, error) {
	if id == "" {
		logs.Logger.Println("FindJobGroupByUUID: ID is empty")
		return nil, errors.New("ID cannot be empty")
	}
	return s.repo.FindJobGroupByUUID(id)
}

// FindAllJobGroups finds all job groups
func (s *jobGroupService) FindAllJobGroups() (*[]models.JobGroup, error) {
	// Implementation for finding all job groups
	return s.repo.FindAllJobGroups()
}

// UpdateJobGroup implements JobGroupService.
func (s *jobGroupService) UpdateJobGroup(jobGroup *models.JobGroup) (*models.JobGroup, error) {
	return s.repo.UpdateJobGroup(jobGroup)
}
