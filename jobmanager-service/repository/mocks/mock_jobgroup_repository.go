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

type MockJobGroupRepository struct {
	mock.Mock
}

func (m *MockJobGroupRepository) SaveJobGroup(jg *models.JobGroup) (*models.JobGroup, error) {
	args := m.Called(jg)
	return args.Get(0).(*models.JobGroup), args.Error(1)
}

func (m *MockJobGroupRepository) UpdateJobGroup(jg *models.JobGroup) (*models.JobGroup, error) {
	args := m.Called(jg)
	return args.Get(0).(*models.JobGroup), args.Error(1)
}

func (m *MockJobGroupRepository) DeleteJobGroup(id string) (int64, error) {
	args := m.Called(id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockJobGroupRepository) FindJobGroupByUUID(id string) (*models.JobGroup, error) {
	args := m.Called(id)
	return args.Get(0).(*models.JobGroup), args.Error(1)
}

func (m *MockJobGroupRepository) FindAllJobGroups() (*[]models.JobGroup, error) {
	args := m.Called()
	return args.Get(0).(*[]models.JobGroup), args.Error(1)
}
