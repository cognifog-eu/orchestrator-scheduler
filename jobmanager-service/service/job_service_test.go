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
	"testing"

	"etsn/server/jobmanager-service/models"
	repository "etsn/server/jobmanager-service/repository/mocks"

	"github.com/stretchr/testify/assert"
)

func TestJobService(t *testing.T) {
	mockRepo := new(repository.MockJobRepository)
	service := NewJobService(mockRepo, nil, nil, nil) // TODO: Update test cases with new functionalities (new job group service)

	job := &models.Job{
		State: models.JobState(models.UpdateDeployment),
	}

	t.Run("SaveJob", func(t *testing.T) {
		mockRepo.On("SaveJob", job).Return(job, nil)
		result, err := service.SaveJob(job)
		assert.NoError(t, err)
		assert.Equal(t, job, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdateJob", func(t *testing.T) {
		mockRepo.On("UpdateJob", job).Return(job, nil)
		result, err := service.UpdateJob(job)
		assert.NoError(t, err)
		assert.Equal(t, job, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteJob", func(t *testing.T) {
		mockRepo.On("DeleteJob", "123").Return(int64(1), nil)
		result, err := service.DeleteJob("123")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindJobByUUID", func(t *testing.T) {
		mockRepo.On("FindJobByUUID", "123").Return(job, nil)
		result, err := service.FindJobByUUID("123")
		assert.NoError(t, err)
		assert.Equal(t, job, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindJobByResourceUUID", func(t *testing.T) {
		mockRepo.On("FindJobByResourceUUID", "res-123").Return(job, nil)
		result, err := service.FindJobByResourceUUID("res-123")
		assert.NoError(t, err)
		assert.Equal(t, job, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindAllJobs", func(t *testing.T) {
		jobs := &[]models.Job{*job}
		mockRepo.On("FindAllJobs").Return(jobs, nil)
		result, err := service.FindAllJobs()
		assert.NoError(t, err)
		assert.Equal(t, jobs, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindJobsByState", func(t *testing.T) {
		jobs := &[]models.Job{*job}
		mockRepo.On("FindJobsByState", models.JobState(models.UpdateDeployment)).Return(jobs, nil)
		result, err := service.FindJobsByState(models.JobState(models.UpdateDeployment))
		assert.NoError(t, err)
		assert.Equal(t, jobs, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindJobsToExecute", func(t *testing.T) {
		jobs := &[]models.Job{*job}
		mockRepo.On("FindJobsToExecute", "type", "owner").Return(jobs, nil)
		result, err := service.FindJobsToExecute("type", "owner")
		assert.NoError(t, err)
		assert.Equal(t, jobs, result)
		mockRepo.AssertExpectations(t)
	})

	// Refactor test to suit new implementation
	// t.Run("JobPromote", func(t *testing.T) {
	// 	mockRepo.On("JobPromote", job).Return(job, nil)
	// 	result, err := service.JobPromote(job)
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, job, result)
	// 	mockRepo.AssertExpectations(t)
	// })
}
