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
package service_test

import (
	"etsn/server/jobmanager-service/models"
	repository "etsn/server/jobmanager-service/repository/mocks"
	"etsn/server/jobmanager-service/service"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestJobGroupService(t *testing.T) {
	mockJobGroupRepo := new(repository.MockJobGroupRepository)
	//TODO: Update test cases with new functionalities (new allocator service, new manifest service...)
	jobGroupService := service.NewJobGroupService(mockJobGroupRepo, nil, nil)

	t.Run("CreateJobGroup", func(t *testing.T) {
		// Given
		bodyBytes := []byte(`name: test-job-group
description: a test job group description
components:
- name: consumer
  type: kubernetes
  manifests:
  - name: mjpeg
  - name: mjpeg-service
  targets:
  - cluster_name: nuvlabox/55c7953e-2aa0-4d18-834c-b4d76d824bb9
    node_name: john-rasbpi-5-1
    orchestrator: nuvla
- name: producer
  type: kubernetes
  manifests:
  - name: video-streaming-service
  - name: video-streaming-deployment
  targets:
  - cluster_name: cluster1
    node_name: cluster1-control-plane
    orchestrator: ocm`)

		header := http.Header{}
		header.Set("Authorization", "Bearer test-token")

		jobGroup := &models.JobGroup{
			AppName:        "test-job-group",
			AppDescription: "a test job group description",
			Jobs: []models.Job{
				{
					Type:  models.CreateDeployment,
					State: models.Created,
					//	JobGroupName: "test-job-group",
					Namespace: uuid.New().String(),
					Resource: &models.Resource{
						ResourceName: "consumer",
						Conditions: []models.Condition{
							{
								Type:               "Created",
								Status:             "True",
								ObservedGeneration: 1,
								LastTransitionTime: time.Now(),
								Reason:             "AwaitingForTarget",
								Message:            "Waiting for the Target",
							},
							{
								Type:               "Created",
								Status:             "True",
								ObservedGeneration: 1,
								LastTransitionTime: time.Now(),
								Reason:             "AwaitingForExecution",
								Message:            "Waiting an Orchestrator to take the Job",
							},
						},
					},
					Target: models.Target{
						ClusterName:  "nuvlabox/55c7953e-2aa0-4d18-834c-b4d76d824bb9",
						NodeName:     "john-rasbpi-5-1",
						Orchestrator: "nuvla",
					},
				},
			},
		}

		mockJobGroupRepo.On("SaveJobGroup", mock.AnythingOfType("*models.JobGroup")).Return(jobGroup, nil)

		// When
		result, err := jobGroupService.CreateJobGroup(bodyBytes, header)

		// Then
		require.NoError(t, err)
		require.NotNil(t, result)
		//require.Equal(t, jobGroup.AppName, result.AppName)
		//require.Equal(t, jobGroup.AppDescription, result.AppDescription)

		mockJobGroupRepo.AssertExpectations(t)
	})
	t.Run("UpdateJobGroup", func(t *testing.T) {
		bodyJob := []byte(`{
			"ID": "27a69131-f34d-44b3-9063-81501a1c0fc8",
			"AppName": "updated-jobgroup",
			"AppDescription": "updated-description",
			"Jobs": [
				{
					"ID": "6616b77c-dbb0-47aa-bc9b-ff45548db029",
					"Type": 5,
					"State": 1,
					"Resource": {"ID": "19d0baf9-0b90-4c36-8e41-5b18e8acbeec"}
				}
			]
		}`)

		existingJobGroup := &models.JobGroup{
			BaseUUID:       models.BaseUUID{ID: "27a69131-f34d-44b3-9063-81501a1c0fc8"},
			AppName:        "existing-jobgroup",
			AppDescription: "existing-description",
			Jobs: []models.Job{
				{
					BaseUUID: models.BaseUUID{ID: "6616b77c-dbb0-47aa-bc9b-ff45548db029"},
					Type:     models.CreateDeployment,
					State:    models.Finished,
					//Resource: &models.Resource{ResourceUID: "19d0baf9-0b90-4c36-8e41-5b18e8acbeec"},
				},
			},
		}

		updatedJobGroup := &models.JobGroup{
			BaseUUID:       models.BaseUUID{ID: "27a69131-f34d-44b3-9063-81501a1c0fc8"},
			AppName:        "updated-jobgroup",
			AppDescription: "updated-description",
			Jobs: []models.Job{
				{
					BaseUUID: models.BaseUUID{ID: "6616b77c-dbb0-47aa-bc9b-ff45548db029"},
					Type:     models.CreateDeployment,
					State:    models.Created,
					//Resource: &models.Resource{ResourceUID: "19d0baf9-0b90-4c36-8e41-5b18e8acbeec"},
				},
			},
		}

		mockJobGroupRepo.On("FindJobGroupByUUID", "27a69131-f34d-44b3-9063-81501a1c0fc8").Return(existingJobGroup, nil)
		mockJobGroupRepo.On("UpdateJobGroup", mock.Anything).Return(updatedJobGroup, nil)

		result, err := jobGroupService.ReplaceJobGroup(bodyJob)
		assert.NoError(t, err)
		assert.Equal(t, updatedJobGroup, result)
		mockJobGroupRepo.AssertExpectations(t)
	})

	t.Run("DeleteJobGroupByID", func(t *testing.T) {
		jobGroupID := uuid.New().String()
		jobID := uuid.New().String()

		existingJobGroup := &models.JobGroup{
			BaseUUID: models.BaseUUID{ID: jobGroupID},
			AppName:  "example-jobgroup",
			Jobs: []models.Job{
				{
					BaseUUID: models.BaseUUID{ID: jobID},
					Type:     models.DeleteDeployment,
					State:    models.Finished,
				},
			},
		}

		mockJobGroupRepo.On("FindJobGroupByUUID", jobGroupID).Return(existingJobGroup, nil)
		mockJobGroupRepo.On("DeleteJobGroup", jobGroupID).Return(int64(1), nil)

		result, err := jobGroupService.DeleteJobGroupByID(jobGroupID)
		assert.NoError(t, err)
		assert.Equal(t, existingJobGroup, result)
		mockJobGroupRepo.AssertExpectations(t)
	})

	t.Run("StopJobGroupByID", func(t *testing.T) {
		jobGroupID := "27a69131-f34d-44b3-9063-81501a1c0fc8"
		existingJobGroup := &models.JobGroup{
			BaseUUID: models.BaseUUID{ID: jobGroupID},
			AppName:  "example-jobgroup",
			Jobs: []models.Job{
				{
					BaseUUID: models.BaseUUID{ID: "6616b77c-dbb0-47aa-bc9b-ff45548db029"},
					Type:     models.DeleteDeployment,
					State:    models.Created,
					Resource: &models.Resource{ID: "19d0baf9-0b90-4c36-8e41-5b18e8acbeec"},
				},
			},
		}

		stoppedJobGroup := &models.JobGroup{
			BaseUUID: models.BaseUUID{ID: jobGroupID},
			AppName:  "example-jobgroup",
			Jobs: []models.Job{
				{
					BaseUUID: models.BaseUUID{ID: "6616b77c-dbb0-47aa-bc9b-ff45548db029"},
					Type:     models.DeleteDeployment,
					State:    models.Created,
					Resource: &models.Resource{ID: "19d0baf9-0b90-4c36-8e41-5b18e8acbeec"},
				},
			},
		}

		mockJobGroupRepo.On("FindJobGroupByUUID", jobGroupID).Return(existingJobGroup, nil)
		mockJobGroupRepo.On("UpdateJobGroup", mock.Anything).Return(stoppedJobGroup, nil)

		result, err := jobGroupService.StopJobGroupByID(jobGroupID)
		assert.NoError(t, err)
		assert.NotEqual(t, stoppedJobGroup, result)
		mockJobGroupRepo.AssertExpectations(t)
	})
}
