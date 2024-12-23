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
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// JobRepository interface defines methods for CRUD operations on Job
type JobRepository interface {
	SaveJob(*models.Job) (*models.Job, error)
	UpdateJob(*models.Job) (*models.Job, error)
	UpdateStoppedJob(job *models.Job) (*models.Job, error)
	DeleteJob(string) (int64, error)
	FindJobByUUID(string) (*models.Job, error)
	FindJobByResourceUUID(string) (*models.Job, error)
	FindAllJobs() (*[]models.Job, error)
	FindJobsByState(state models.JobState) (*[]models.Job, error)
	FindJobsToExecute(orchestratorType, ownerID string) (*[]models.Job, error)
	JobPromote(*models.Job) (*models.Job, error)
}

// jobRepository is the implementation of JobRepository
type jobRepository struct {
	db *gorm.DB
}

// NewJobRepository returns a new instance of jobRepository
func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{db: db}
}

// SaveJob saves a new job to the database
func (repo *jobRepository) SaveJob(job *models.Job) (*models.Job, error) {

	tx := repo.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Debug().Create(&job).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return job, nil
}

// UpdateJob updates an existing job in the database
func (repo *jobRepository) UpdateJob(job *models.Job) (*models.Job, error) {
	tx := repo.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			logs.Logger.Println("Panic occurred, rolling back transaction:", r)
			tx.Rollback()
		}
	}()

	if err := tx.Debug().Session(&gorm.Session{FullSaveAssociations: true}).Where("id = ?", job.ID).Updates(job).Error; err != nil {
		logs.Logger.Println("Error updating job:", err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logs.Logger.Println("Error committing transaction:", err)
		tx.Rollback()
		return nil, err
	}

	return job, nil
}

func (repo *jobRepository) UpdateStoppedJob(job *models.Job) (*models.Job, error) {
	tx := repo.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			logs.Logger.Println("Panic occurred, rolling back transaction:", r)
			tx.Rollback()
		}
	}()

	// Create a map to store the fields and values for update
	updateData := map[string]interface{}{
		"State":    job.State,
		"owner_id": job.OwnerID,
	}

	if err := tx.Debug().
		Table("jobs"). // Specify the table here
		Session(&gorm.Session{FullSaveAssociations: true}).
		Omit("Resource").
		Where("id = ?", job.ID).
		Updates(updateData).Error; err != nil {
		logs.Logger.Println("Error updating job:", err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logs.Logger.Println("Error committing transaction:", err)
		tx.Rollback()
		return nil, err
	}

	return job, nil
}

// DeleteJob deletes a job from the database
func (repo *jobRepository) DeleteJob(id string) (int64, error) {
	tx := repo.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete targets associated with the job
	if err := tx.Debug().Where("job_id = ?", id).Delete(&models.Target{}).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// Delete job
	result := tx.Debug().Where("id = ?", id).Delete(&models.Job{})
	if result.Error != nil {
		tx.Rollback()
		return 0, result.Error
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

// FindJobByUUID finds a job by its UUID
func (repo *jobRepository) FindJobByUUID(id string) (*models.Job, error) {
	var job models.Job
	err := repo.db.Debug().
		Model(models.Job{}).
		Where("id = ?", id).
		Preload(clause.Associations).
		Preload("Target").
		Preload("Resource").
		Preload("Resource.Conditions").
		Preload("Resource.Remediations").
		Preload("Instruction").
		Preload("Instruction.Contents").
		Take(&job).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job not found")
		}
		return nil, err
	}
	return &job, nil
}

// FindJobByResourceUUID finds a job by its resource UUID
func (repo *jobRepository) FindJobByResourceUUID(id string) (*models.Job, error) {
	var job models.Job

	err := repo.db.Debug().
		Model(&models.Job{}).
		Joins("JOIN resources ON resources.job_id = jobs.id").
		Where("resources.id = ?", id).
		Preload("Target").
		Preload("Resource").
		First(&job).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("job not found")
		}
		return nil, err
	}

	return &job, nil
}

// FindAllJobs retrieves all jobs from the database
func (repo *jobRepository) FindAllJobs() (*[]models.Job, error) {
	jobs := []models.Job{}
	err := repo.db.Debug().Model(&models.Job{}).Preload(clause.Associations).Preload("Target").Preload("Resource").Find(&jobs).Error
	if err != nil {
		return nil, err
	}
	return &jobs, nil
}

// FindJobsByState retrieves jobs by their state
func (repo *jobRepository) FindJobsByState(state models.JobState) (*[]models.Job, error) {
	var jobs []models.Job
	err := repo.db.Debug().Model(&models.Job{}).Where("state = ?", string(state)).Preload(clause.Associations).Preload("Target").Find(&jobs).Error
	if err != nil {
		return nil, err
	}
	return &jobs, nil
}

func (repo *jobRepository) FindJobsToExecute(orchestratorType, ownerID string) (*[]models.Job, error) {
	var jobs []models.Job
	err := repo.db.Debug().Preload(clause.Associations).Preload("Target").Preload("Resource").Preload("Instruction").
		Preload("Instruction.Contents").Preload("Resource.Remediations").Preload("Resource.Remediations.RemediationTarget").
		Find(&jobs,
			"((type = ?) AND state = ? AND (owner_id = '' OR owner_id IS NULL) AND orchestrator = ?) OR "+
				"((type = ?) AND (state = ? OR state = ?) AND owner_id != ? AND updated_at < ? AND orchestrator = ?) OR "+
				"(state = ? AND owner_id = ? AND orchestrator = ?) OR "+
				"(type = ? AND state = ? AND owner_id = ? AND orchestrator = ?)",
			// (1) CreateDeployment, JobCreated, owner_id = nil
			models.CreateDeployment, models.Created, orchestratorType,
			// (2) CreateDeployment, JobProgressing or JobDegraded, owner_id != ownerID, updated_at < now - 300s
			models.CreateDeployment, models.Progressing, models.Degraded, ownerID, time.Now().Local().Add(time.Second*time.Duration(-300)), orchestratorType,
			// (3) UpdateDeployment/CreateDeployment(Start), JobCreated, owner_id = ownerID
			models.Created, ownerID, orchestratorType,
			// (4) DeleteDeployment, JobCreated, owner_id = ownerID
			models.DeleteDeployment, models.Created, ownerID, orchestratorType).Error
	if err != nil {
		return nil, err
	}
	return &jobs, nil
}

// JobLocker updates the lock state of a job
func (repo *jobRepository) JobPromote(job *models.Job) (*models.Job, error) {
	tx := repo.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// log.Println("Setting new TTL for the Job before update: " + job.ID)
	err := tx.Debug().Model(&models.Job{}).Where("id = ?", job.ID).Updates(
		models.Job{OwnerID: job.OwnerID, State: job.State}).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Display the updated Job
	err = tx.Debug().Model(models.Job{}).Where("id = ?", job.ID).Preload("Target").Preload("Resource").Take(job).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return job, nil
}
