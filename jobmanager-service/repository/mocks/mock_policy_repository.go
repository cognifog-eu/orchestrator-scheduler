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

type MockPolicyRepository struct {
	mock.Mock
}

func (m *MockPolicyRepository) SaveRemediation(incompliance *models.Remediation) (*models.Remediation, error) {
	args := m.Called(incompliance)
	return args.Get(0).(*models.Remediation), args.Error(1)
}
