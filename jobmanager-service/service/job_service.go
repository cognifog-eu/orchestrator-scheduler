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
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type JobService interface {
	SaveJob(*models.Job) (*models.Job, error)
	UpdateJob(*models.Job) (*models.Job, error)
	UpdateStoppedJob(job *models.Job) (*models.Job, error)
	ReallocateJob(job *models.Job, header http.Header) (*models.Job, error)
	ReportJobState(*models.Job) (*models.Job, error)
	DeleteJob(string) (int64, error)
	FindJobByUUID(string) (*models.Job, error)
	FindJobByResourceUUID(string) (*models.Job, error)
	FindAllJobs() (*[]models.Job, error)
	FindJobsByState(state models.JobState) (*[]models.Job, error)
	FindJobsToExecute(orchestratorType, ownerID string) (*[]models.Job, error)
	JobPromote(r *http.Request) (*models.Job, error)
	UpdateJobForRemediation(job *models.Job, remediation *models.Remediation) error
}

type jobService struct {
	repo             repository.JobRepository
	allocatorService AllocatorService
	jobGroupService  JobGroupService
	resourceService  ResourceService
}

func NewJobService(repo repository.JobRepository, resourceService ResourceService, allocatorService AllocatorService, jobGroupService JobGroupService) JobService {
	return &jobService{repo: repo, resourceService: resourceService, allocatorService: allocatorService, jobGroupService: jobGroupService}
}

func (s *jobService) SaveJob(job *models.Job) (*models.Job, error) {
	return s.repo.SaveJob(job)
}

// ReallocateJob implements JobService.
func (s *jobService) ReallocateJob(job *models.Job, header http.Header) (*models.Job, error) {

	jobGroup, allocResDTO, err := s.jobGroupService.AllocateJobGroup(job.JobGroupID, header)
	if err != nil {
		logs.Logger.Printf("ReallocateJob: Error allocating JobGroup %s, error: %v", job.JobGroupID, err)
		return nil, err
	}

	var foundJob *models.Job
	for i := range jobGroup.Jobs {
		if job.ID == jobGroup.Jobs[i].ID {
			foundJob = &jobGroup.Jobs[i]
			break
		}
	}
	if foundJob == nil {
		logs.Logger.Printf("ReallocateJob: Job %s not found in JobGroup %s", job.ID, job.JobGroupID)
		return nil, errors.New("job not found in job group")
	}

	for _, comp := range allocResDTO.Components {
		if foundJob.Instruction.ComponentName == comp.Name {
			if err := s.allocatorService.AssignTargets(foundJob, comp.Target); err != nil {
				logs.Logger.Printf("ReallocateJob: Error assigning targets for Job %s, error: %v", job.ID, err)
				return nil, err
			}
			// Make Job executable
			foundJob.State = models.Created
			foundJob.Type = models.CreateDeployment
			foundJob.OwnerID = "" // to make sure the job is executable...
			break
		}
	}

	return s.UpdateJob(foundJob)

}

// UpdateStoppedJob implements JobService.
func (s *jobService) UpdateStoppedJob(job *models.Job) (*models.Job, error) {
	return s.repo.UpdateStoppedJob(job)
}

func (s *jobService) ReportJobState(job *models.Job) (*models.Job, error) {

	jobGotten, err := s.FindJobByUUID(job.ID)
	if err != nil {
		logs.Logger.Println("Error finding job:", err)
		return &models.Job{}, err
	}

	// ensure job is in a progressing state
	err = validateJobUpdate(jobGotten)
	if err != nil {
		logs.Logger.Println("Error validating job update:", err)
		jobGotten.State = models.Degraded
		return s.UpdateJob(jobGotten)
	}

	switch jobGotten.Type {
	case models.CreateDeployment:
		jobGotten.State = models.Finished
		jobGotten.Resource = job.Resource
	case models.DeleteDeployment:
		// Delete a resource conditions and then the resource
		ok, _ := s.resourceService.DeleteResource(jobGotten.Resource.ID)
		if ok == 0 {
			logs.Logger.Println("Error deleting resource")
			jobGotten.State = models.Degraded
			return s.UpdateJob(jobGotten)
		}
		jobGotten.Resource = nil // we set the reference to nil once we deleted the resource in db
		jobGotten.OwnerID = ""
		// now job can be considered finished
		jobGotten.State = models.Finished
		return s.UpdateStoppedJob(jobGotten)
	case models.UpdateDeployment:
		jobGotten.Resource = job.Resource
		jobGotten.State = models.Finished
	default:
		err := errors.New("unknown job type")
		job.State = models.Degraded
		logs.Logger.Printf("Job with ID %s is in state %s, cannot be updated: %v", job.ID, job.State, err)
		return s.UpdateJob(jobGotten)
	}

	return s.UpdateJob(jobGotten)
}

func (s *jobService) UpdateJob(job *models.Job) (*models.Job, error) {
	return s.repo.UpdateJob(job)
}

func validateJobUpdate(job *models.Job) error {
	if job.State != models.Progressing {
		return errors.New("job is not in a progressing state")
	}
	if job.OwnerID == "" {
		return errors.New("job owner ID is empty")
	}
	return nil
}

func (s *jobService) DeleteJob(id string) (int64, error) {
	return s.repo.DeleteJob(id)
}

func (s *jobService) FindJobByUUID(id string) (*models.Job, error) {
	return s.repo.FindJobByUUID(id)
}

func (s *jobService) FindJobByResourceUUID(id string) (*models.Job, error) {
	return s.repo.FindJobByResourceUUID(id)
}

func (s *jobService) FindAllJobs() (*[]models.Job, error) {
	return s.repo.FindAllJobs()
}

func (s *jobService) FindJobsByState(state models.JobState) (*[]models.Job, error) {
	return s.repo.FindJobsByState(state)
}

func (s *jobService) FindJobsToExecute(orchestratorType, ownerID string) (*[]models.Job, error) {
	jobs, err := s.repo.FindJobsToExecute(orchestratorType, ownerID)
	if err != nil {
		logs.Logger.Printf("Error retrieving jobs to execute: %v", err)
		return nil, err
	}

	return jobs, nil

}

func (s *jobService) JobPromote(r *http.Request) (*models.Job, error) {
	vars := mux.Vars(r)
	stringJobID := vars["job_uuid"]
	if stringJobID == "" {
		err := errors.New("job ID Cannot be empty")
		logs.Logger.Println("job ID Cannot be empty")
		return nil, err
	}

	var jobOwnershipDTO models.JobOwnershipDTO
	err := json.NewDecoder(r.Body).Decode(&jobOwnershipDTO)
	if err != nil {
		logs.Logger.Printf("Error decoding job patch body: %v", err)
		return nil, err
	}
	if jobOwnershipDTO.OwnerID == "" {
		err := errors.New("owner ID Cannot be empty")
		logs.Logger.Println("owner ID Cannot be empty")
		return nil, err
	}

	jobGotten, err := s.FindJobByUUID(stringJobID)
	if err != nil {
		logs.Logger.Printf("Error retrieving job: %v", err)
		return nil, err
	}

	switch jobGotten.State {
	case models.Created:
		jobGotten.OwnerID = jobOwnershipDTO.OwnerID
		jobGotten.State = models.Progressing
	case models.Progressing, models.Finished, models.Degraded:
		err := errors.New("job cannot be promoted")
		logs.Logger.Printf("Job with ID %s is in state %s, cannot be promoted", jobGotten.ID, jobGotten.State)
		return nil, err
	default:
		err := errors.New("job cannot be promoted")
		logs.Logger.Printf("Job with ID %s is in an unknown state, cannot be promoted", jobGotten.ID)
		return nil, err
	}

	updatedJob, err := s.repo.JobPromote(jobGotten)
	if err != nil {
		logs.Logger.Printf("Error updating job state: %v", err)
		return nil, err
	}

	return updatedJob, nil
}

func (s *jobService) UpdateJobForRemediation(job *models.Job, remediation *models.Remediation) error {

	// Set it to "" so that beforecreate hook assigns new ID
	remediation.ID = ""
	if remediation.RemediationTarget != nil {
		remediation.RemediationTarget.ID = ""
	}

	// Update job fields based on remediation
	job.State = models.Created
	job.Type = models.UpdateDeployment

	// Set job subtype
	switch remediation.RemediationType {
	case models.ScaleUp, models.ScaleDown, models.ScaleIn, models.ScaleOut,
		models.Replace, models.Secure, models.Patch:
		job.SubType = remediation.RemediationType
	case models.Reallocate:
		job.Type = models.DeleteDeployment
		job.SubType = models.Reallocate
	default:
		return fmt.Errorf("unsupported remediation type: %s", remediation.RemediationType)
	}

	// Append remediation to job resource
	if job.Resource.Remediations == nil {
		job.Resource.Remediations = []models.Remediation{}
	}
	job.Resource.Remediations = append(job.Resource.Remediations, *remediation)

	// Update the job in the repository
	_, err := s.repo.UpdateJob(job)
	return err
}
