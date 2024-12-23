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
package repository

import (
	"etsn/server/jobmanager-service/models"

	"github.com/stretchr/testify/mock"
)

// MockJobRepository is a mock implementation of JobRepository
type MockJobRepository struct {
	mock.Mock
}

// UpdateStoppedJob implements repository.JobRepository.
func (m *MockJobRepository) UpdateStoppedJob(job *models.Job) (*models.Job, error) {
	args := m.Called(job)
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) SaveJob(job *models.Job) (*models.Job, error) {
	args := m.Called(job)
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) UpdateJob(job *models.Job) (*models.Job, error) {
	args := m.Called(job)
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) DeleteJob(id string) (int64, error) {
	args := m.Called(id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockJobRepository) FindJobByUUID(id string) (*models.Job, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) FindJobByResourceUUID(id string) (*models.Job, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) FindAllJobs() (*[]models.Job, error) {
	args := m.Called()
	return args.Get(0).(*[]models.Job), args.Error(1)
}

func (m *MockJobRepository) FindJobsByState(state models.JobState) (*[]models.Job, error) {
	args := m.Called(state)
	return args.Get(0).(*[]models.Job), args.Error(1)
}

func (m *MockJobRepository) FindJobsToExecute(orchestratorType, ownerID string) (*[]models.Job, error) {
	args := m.Called(orchestratorType, ownerID)
	return args.Get(0).(*[]models.Job), args.Error(1)
}

func (m *MockJobRepository) JobPromote(job *models.Job) (*models.Job, error) {
	args := m.Called(job)
	return args.Get(0).(*models.Job), args.Error(1)
}
