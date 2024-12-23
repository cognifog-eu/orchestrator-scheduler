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

type MockResourceRepository struct {
	mock.Mock
}

func (m *MockResourceRepository) SaveResource(r *models.Resource) (*models.Resource, error) {
	args := m.Called(r)
	return args.Get(0).(*models.Resource), args.Error(1)
}

func (m *MockResourceRepository) UpdateAResource(r *models.Resource) (*models.Resource, error) {
	args := m.Called(r)
	return args.Get(0).(*models.Resource), args.Error(1)
}

func (m *MockResourceRepository) AddCondition(r *models.Resource, c *models.Condition) (*models.Resource, error) {
	args := m.Called(r, c)
	return args.Get(0).(*models.Resource), args.Error(1)
}

func (m *MockResourceRepository) RemoveConditions(r *models.Resource) error {
	args := m.Called(r)
	return args.Error(1)
}

func (m *MockResourceRepository) FindResourceByJobUUID(jobId string) (*models.Resource, error) {
	args := m.Called(jobId)
	return args.Get(0).(*models.Resource), args.Error(1)
}

func (m *MockResourceRepository) DeleteResource(resourceId string) (int64, error) {
	args := m.Called(resourceId)
	return args.Get(0).(int64), args.Error(1)
}
