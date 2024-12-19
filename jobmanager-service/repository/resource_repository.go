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
	"errors"
	"etsn/server/jobmanager-service/models"
	"etsn/server/jobmanager-service/utils/logs"

	"gorm.io/gorm"
)

type ResourceRepository interface {
	SaveResource(*models.Resource) (*models.Resource, error)
	UpdateAResource(*models.Resource) (*models.Resource, error)
	DeleteResource(string) (int64, error)
	AddCondition(*models.Resource, *models.Condition) (*models.Resource, error)
	RemoveConditions(*models.Resource) error
	FindResourceByJobUUID(string) (*models.Resource, error)
}

type resourceRepository struct {
	db *gorm.DB
}

func NewResourceRepository(db *gorm.DB) ResourceRepository {
	return &resourceRepository{db: db}
}

func (repo *resourceRepository) UpdateAResource(resource *models.Resource) (*models.Resource, error) {
	// logs.Logger.Println("Updating the resource: " + resource.ResourceUID)
	// repo.db = repo.db.Session(&gorm.Session{FullSaveAssociations: true}).Where("job_id = ?", resource.JobID).
	// 	Updates(&models.Resource{ResourceUID: resource.ResourceUID, ResourceName: resource.ResourceName})
	// if repo.db.Error != nil {
	// 	return &models.Resource{}, repo.db.Error
	// }

	// // // This is the display the updated Job
	// // err := repo.db.Debug().Model(models.Resource{}).Where("job_id = ?", resource.JobID).Preload("Conditions").Take(&resource).Error
	// // if err != nil {
	// // 	return &models.Resource{}, err
	// // }
	// return resource, repo.db.Error

	tx := repo.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			logs.Logger.Println("Panic occurred, rolling back transaction:", r)
			tx.Rollback()
		}
	}()

	if err := tx.Debug().Session(&gorm.Session{FullSaveAssociations: true}).Where("id = ?", resource.ID).Updates(resource).Error; err != nil {
		logs.Logger.Println("Error updating job:", err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logs.Logger.Println("Error committing transaction:", err)
		tx.Rollback()
		return nil, err
	}

	return resource, nil

}

func (repo *resourceRepository) AddCondition(resource *models.Resource, condition *models.Condition) (*models.Resource, error) {
	logs.Logger.Println("Updating the resource: " + resource.ID)
	err := repo.db.Debug().Create(&condition)
	if err != nil {
		return &models.Resource{}, repo.db.Error
	}
	// This is the display the updated Job
	err = repo.db.Debug().Model(models.Resource{}).Where("id =?", resource.ID).Preload("Conditions").Take(&resource)
	return resource, err.Error
}

func (repo *resourceRepository) RemoveConditions(resource *models.Resource) error {
	logs.Logger.Println("Removing old conditions of the resource: " + resource.ID)
	var err error
	tx := repo.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete job
	// CHECK IF RESOURCE_UID IS THE NAME OF THE COLUMN
	result := tx.Debug().Where("resource_id = ?", resource.ID).Delete(&models.Condition{})
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return err
}

func (repo *resourceRepository) SaveResource(resource *models.Resource) (*models.Resource, error) {
	err := repo.db.Debug().Create(&resource).Error
	if err != nil {
		return &models.Resource{}, err
	}
	return resource, nil
}

func (repo *resourceRepository) FindResourceByJobUUID(jobId string) (*models.Resource, error) {
	resource := &models.Resource{}
	err := repo.db.Debug().Model(models.Resource{}).Where("job_id = ?", jobId).Preload("Conditions").Take(&resource).Error
	if err != nil {
		return &models.Resource{}, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &models.Resource{}, errors.New("job Not Found")
	}
	return resource, err
}

func (repo *resourceRepository) DeleteResource(id string) (int64, error) {
	tx := repo.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete conditions associated with the resource
	if err := tx.Debug().Where("resource_id = ?", id).Delete(&models.Condition{}).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// Delete resource
	result := tx.Debug().Where("id = ?", id).Delete(&models.Resource{})
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}
