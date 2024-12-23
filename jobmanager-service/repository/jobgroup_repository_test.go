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

func initJobGroupRepo(db *gorm.DB) interface{} {
	return NewJobGroupRepository(db)
}

func TestSaveJobGroup(t *testing.T) {
	repo := mocks.SetupTest(t, initJobGroupRepo).(JobGroupRepository)

	jobGroup := models.JobGroup{}

	result, err := repo.SaveJobGroup(&jobGroup)
	assert.NoError(t, err)
	assert.Equal(t, jobGroup.ID, result.ID)
}

func TestUpdateJobGroup(t *testing.T) {
	repo := mocks.SetupTest(t, initJobGroupRepo).(JobGroupRepository)

	jobGroup := models.JobGroup{
		AppDescription: "Description",
	}
	repo.SaveJobGroup(&jobGroup)

	jobGroup.AppDescription = "Updated Description"
	result, err := repo.UpdateJobGroup(&jobGroup)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Description", result.AppDescription)
}

func TestDeleteJobGroup(t *testing.T) {
	repo := mocks.SetupTest(t, initJobGroupRepo).(JobGroupRepository)

	jobGroup := models.JobGroup{}
	repo.SaveJobGroup(&jobGroup)

	rowsAffected, err := repo.DeleteJobGroup(jobGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
}

func TestFindJobGroupByUUID(t *testing.T) {
	repo := mocks.SetupTest(t, initJobGroupRepo).(JobGroupRepository)

	jobGroup := models.JobGroup{}
	repo.SaveJobGroup(&jobGroup)

	result, err := repo.FindJobGroupByUUID(jobGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, jobGroup.ID, result.ID)
}

func TestFindAllJobGroups(t *testing.T) {
	repo := mocks.SetupTest(t, initJobGroupRepo).(JobGroupRepository)

	jobGroup1 := models.JobGroup{}
	jobGroup2 := models.JobGroup{}
	repo.SaveJobGroup(&jobGroup1)
	repo.SaveJobGroup(&jobGroup2)

	result, err := repo.FindAllJobGroups()
	assert.NoError(t, err)
	assert.Len(t, *result, 2)
}
