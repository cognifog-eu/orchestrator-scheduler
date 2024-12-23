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
	"etsn/server/jobmanager-service/utils/logs"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// JobGroupRepository interface defines the methods for CRUD operations
type JobGroupRepository interface {
	SaveJobGroup(*models.JobGroup) (*models.JobGroup, error)
	UpdateJobGroup(*models.JobGroup) (*models.JobGroup, error)
	DeleteJobGroup(string) (int64, error)
	FindJobGroupByUUID(string) (*models.JobGroup, error)
	FindAllJobGroups() (*[]models.JobGroup, error)
}

// jobGroupRepository is the implementation of JobGroupRepository
type jobGroupRepository struct {
	db *gorm.DB
}

// NewJobGroupRepository returns a new instance of jobGroupRepository
func NewJobGroupRepository(db *gorm.DB) JobGroupRepository {
	return &jobGroupRepository{db: db}
}

// SaveJobGroup saves a new job group to the database
func (repo *jobGroupRepository) SaveJobGroup(jg *models.JobGroup) (*models.JobGroup, error) {

	tx := repo.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			logs.Logger.Println("Error starting transaction:", r)
			tx.Rollback()
		}
	}()

	//jg.BeforeCreate(tx)
	if err := tx.Debug().Create(&jg).Error; err != nil {
		logs.Logger.Println("Error saving job group:", err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logs.Logger.Println("Error saving job group:", err)
		return nil, err
	}
	return jg, nil
}

func (repo *jobGroupRepository) UpdateJobGroup(jg *models.JobGroup) (*models.JobGroup, error) {

	tx := repo.db.Begin().Model(&jg).Where("id = ?", jg.ID)

	if tx.Error != nil {
		logs.Logger.Println("Error starting transaction:", tx.Error)
		return nil, tx.Error
	}

	if err := tx.Debug().Session(&gorm.Session{FullSaveAssociations: true}).Updates(&jg).Error; err != nil {
		logs.Logger.Println("Error saving job group:", err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logs.Logger.Println("Error committing transaction:", err)
		tx.Rollback()
		return nil, err
	}

	return jg, nil
}

// DeleteJobGroup deletes a job group from the database
func (repo *jobGroupRepository) DeleteJobGroup(id string) (int64, error) {
	tx := repo.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Debug().Where("id = ?", id).Delete(&models.JobGroup{})
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

// FindJobGroupByUUID finds a job group by its UUID
func (repo *jobGroupRepository) FindJobGroupByUUID(id string) (*models.JobGroup, error) {
	jobGroup := models.JobGroup{}
	err := repo.db.Debug().
		Preload("Jobs").
		Preload("Jobs.Instruction.Contents").
		Preload("Jobs.Target").
		Preload("Jobs.Resource.Conditions"). // TODO: Review if we need to preload conditions only
		Preload("Jobs.Resource.Remediations").
		Where("id = ?", id).
		First(&jobGroup).Error
	if err != nil {
		return nil, err
	}
	return &jobGroup, nil
}

// FindAllJobGroups returns all job groups from the database
func (repo *jobGroupRepository) FindAllJobGroups() (*[]models.JobGroup, error) {
	var err error
	jobGroups := []models.JobGroup{}
	err = repo.db.
		Preload(clause.Associations).
		Preload("Jobs." + clause.Associations).
		Preload("Jobs.Resource.Conditions").
		Preload("Jobs.Instruction.Contents").
		Preload("Jobs.Resource.Remediations").
		Find(&jobGroups).Error
	if err != nil {
		return &[]models.JobGroup{}, err
	}
	return &jobGroups, err
}
