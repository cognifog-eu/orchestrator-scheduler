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
	mocks "etsn/server/jobmanager-service/repository/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func initPolicyRepo(db *gorm.DB) interface{} {
	return NewPolicyRepository(db)
}

func TestSaveRemediation(t *testing.T) {
	repo := mocks.SetupTest(t, initPolicyRepo).(PolicyRepository)

	incom := &models.Remediation{}

	result, err := repo.SaveRemediation(incom)
	assert.NoError(t, err)
	assert.Equal(t, incom.ID, result.ID)
}
