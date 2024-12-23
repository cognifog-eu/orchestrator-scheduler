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

func initResourceRepo(db *gorm.DB) interface{} {
	return NewResourceRepository(db)
}

func TestSaveResource(t *testing.T) {
	repo := mocks.SetupTest(t, initResourceRepo).(ResourceRepository)

	resource := &models.Resource{
		ID:           "12kjgksdj3",
		ResourceName: "test",
		Conditions:   []models.Condition{},
	}

	result, err := repo.SaveResource(resource)

	assert.NoError(t, err)
	assert.NotNil(t, result.ID)
	assert.Equal(t, resource.ResourceName, result.ResourceName)

	// Additional verification to ensure the resource is actually saved in the database
	savedResource, err := repo.FindResourceByJobUUID(resource.JobID)
	assert.NoError(t, err)
	assert.Equal(t, resource.ResourceName, savedResource.ResourceName)
}

func TestUpdateAResource(t *testing.T) {
	repo := mocks.SetupTest(t, initResourceRepo).(ResourceRepository)

	resource := &models.Resource{ResourceName: "test", Conditions: []models.Condition{}}
	repo.SaveResource(resource)

	resource.ResourceName = "updated"
	result, err := repo.UpdateAResource(resource)

	assert.NoError(t, err)
	assert.Equal(t, resource.ResourceName, result.ResourceName)

}

func TestAddCondition(t *testing.T) {
	repo := mocks.SetupTest(t, initResourceRepo).(ResourceRepository)

	resource := &models.Resource{
		ID:           "12kjgksdj3",
		ResourceName: "test",
		Conditions:   []models.Condition{}}
	repo.SaveResource(resource)

	condition := &models.Condition{Type: "Ready"}
	result, err := repo.AddCondition(resource, condition)

	assert.NoError(t, err)
	assert.NotNil(t, result.Conditions[0].ID)

}

func TestRemoveConditions(t *testing.T) {
	repo := mocks.SetupTest(t, initResourceRepo).(ResourceRepository)

	resource := &models.Resource{ResourceName: "test", Conditions: []models.Condition{}}
	repo.SaveResource(resource)

	condition := &models.Condition{Type: "Ready"}
	repo.AddCondition(resource, condition)

	err := repo.RemoveConditions(resource)

	assert.NoError(t, err)
	// assert.Equal(t, 0, len(res.Conditions))

}

func TestFindResourceByJobUUID(t *testing.T) {
	repo := mocks.SetupTest(t, initResourceRepo).(ResourceRepository)

	resource := &models.Resource{
		ID:           "12kjgksdj3",
		ResourceName: "test",
		Conditions:   []models.Condition{}}

	repo.SaveResource(resource)

	result, err := repo.FindResourceByJobUUID(resource.JobID)

	assert.NoError(t, err)

	assert.Equal(t, resource.ID, result.ID)

}
