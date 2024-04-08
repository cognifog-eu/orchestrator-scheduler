/*
Copyright 2023 Bull SAS

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
package models

import (
	"errors"
	"etsn/server/jobmanager-service/utils/logs"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobState int
type JobType int

type OrchestratorType string

const (
	OCM   OrchestratorType = "ocm"
	Nuvla OrchestratorType = "nuvla"
	None  OrchestratorType = "none"
)

const (
	JobCreated JobState = iota + 1
	JobProgressing
	JobFinished
	JobDegraded

	CreateDeployment JobType = iota + 1
	GetDeployment
	DeleteDeployment
	RecoveryJob
)

// hold information that N jobs share (N jobs needed to provide application x)
type JobGroup struct {
	ID             uuid.UUID `gorm:"type:char(36);primary_key"`
	AppName        string    `json:"appName"`
	AppDescription string    `json:"appDescription"`
	Jobs           []Job     `json:"jobs"`
}

// TODO: this Job is pulled by the drivers, we should agree on Jobs model
type Job struct {
	// gorm.Model
	ID                 uuid.UUID        `gorm:"type:char(36);primary_key"`     // unique across all ecosystem
	UUID               uuid.UUID        `gorm:"type:text" json:"uuid"`         // optional and unique across all ecosystem
	JobGroupID         uuid.UUID        `gorm:"type:text" json:"job_group_id"` // unique across all ecosystem
	JobGoupName        string           `gorm:"type:text" json:"job_group_name"`
	JobGoupDescription string           `gorm:"type:text" json:"job_group_description,omitempty"`
	Type               JobType          `gorm:"type:text" json:"type"`
	State              JobState         `gorm:"type:text" json:"state"`
	Manifest           string           `gorm:"type:text" json:"manifest"`
	Targets            []Target         `json:"targets"` // array of targets where the Manifest is applied
	Locker             *bool            `json:"locker"`
	Orchestrator       OrchestratorType `gorm:"type:text" json:"orchestrator"` // identifies the orchestrator that can execute the job based on target provided by MM
	UpdatedAt          time.Time        `json:"updated_at"`
	Resource           Resource         `gorm:"foreignkey:JobID;" json:"resource,omitempty"`
	Namespace          string           `gorm:"type:text" json:"namespace"`
}

type MMResponseMapper struct {
	Components []Component `json:"components"`
}

type Component struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind,omitempty"`
	Manifests string   `json:"manifests,omitempty"`
	Targets   []Target `json:"targets"`
}

type Target struct {
	ID           uint32           `gorm:"primary_key" json:"-"`
	JobID        uuid.UUID        `json:"-"`
	ClusterName  string           `json:"cluster_name"`
	NodeName     string           `json:"node_name,omitempty"`
	Orchestrator OrchestratorType `gorm:"type:text" json:"orchestrator"`
}

func (job *Job) BeforeCreate(tx *gorm.DB) (err error) {
	// UUID version 4
	job.ID = uuid.New()
	b := new(bool)
	*b = false
	job.Locker = b
	return
}

func (jobGroup *JobGroup) BeforeCreate(tx *gorm.DB) (err error) {
	// UUID version 4
	jobGroup.ID = uuid.New()
	return
}

func StateIsValid(value int) bool {
	return int(JobCreated) >= value && value <= int(Degraded)
}

func OrchestratorTypeMapper(orchestratorType string) OrchestratorType {

	if orchestratorType == string(Nuvla) {
		return Nuvla
	} else if orchestratorType == string(OCM) {
		return OCM
	} else {
		return None
	}
}

func JobTypeIsValid(value int) bool {
	return int(CreateDeployment) >= value && value <= int(DeleteDeployment)
}

func (j *Job) NewJobTTL() {
	if time.Now().Unix()-j.UpdatedAt.Unix() > int64(300) {
		b := new(bool)
		*b = false
		j.Locker = b
	}
}

func (j *Job) SaveJob(db *gorm.DB) (*Job, error) {
	// create UUID before saving the Job to DB
	j.BeforeCreate(db)
	err := db.Debug().Create(&j).Error
	if err != nil {
		return &Job{}, err
	}
	return j, nil
}

func (j *Job) FindJobByUUID(db *gorm.DB, uuid uuid.UUID) (*Job, error) {
	err := db.Debug().Model(Job{}).Where("id = ?", uuid).Preload("Targets").Preload("Resource").Take(&j).Error
	if err != nil {
		return &Job{}, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &Job{}, errors.New("Job Not Found")
	}
	return j, err
}

func (j *Job) FindJobByResourceUUID(db *gorm.DB, uuid uuid.UUID) (*Job, error) {
	err := db.Debug().Model(Job{}).Where("uuid = ?", uuid).Preload("Targets").Preload("Resource").Take(&j).Error
	if err != nil {
		return &Job{}, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &Job{}, errors.New("Job Not Found")
	}
	return j, err
}

func (u *Job) FindAllJobs(db *gorm.DB) (*[]Job, error) {
	var err error
	jobs := []Job{}
	err = db.Debug().Model(&Job{}).Preload("Targets").Preload("Resource").Find(&jobs).Error
	if err != nil {
		return &[]Job{}, err
	}
	return &jobs, err
}

func (j *Job) FindJobsByState(db *gorm.DB, state int) (*[]Job, error) {
	var err error
	jobs := []Job{}
	err = db.Debug().Model(&Job{}).Where("state = ?", state).Preload("Targets").Find(&jobs).Error
	if err != nil {
		return &[]Job{}, err
	}
	return &jobs, err
}

func (j *Job) FindJobsToExecute(db *gorm.DB) (*[]Job, error) {
	var err error
	jobs := []Job{}
	err = db.Debug().Preload("Targets").Preload("Resource").Find(&jobs, "state =? AND locker = FALSE AND orchestrator =? OR state =? AND locker = TRUE AND orchestrator =? AND updated_at < ?", int(JobCreated), j.Orchestrator, int(Progressing), j.Orchestrator, time.Now().Local().Add(time.Second*time.Duration(-300))).Error
	// err = db.Debug().Model(&Job{}).Where(db.Where("state = ?", int(Created)).Where("locker = ?", false)).
	// 	Or(db.Where("state = ?", int(Progressing)).Where("locker = ?", true)).Where("updated_at < ?", time.Now().Local().Add(time.Second*time.Duration(-300))).
	// 	Preload("Targets").Find(&jobs).Error
	if err != nil {
		return &[]Job{}, err
	}
	return &jobs, err
}

func (j *Job) UpdateAJob(db *gorm.DB, uuid uuid.UUID) (*Job, error) {
	// trigger TTL ticker on each writing access except the CreateJob
	logs.Logger.Println("Setting new TTL for the Job before update: " + j.ID.String())
	j.NewJobTTL()
	db = db.Debug().Model(&Job{}).Where("id = ?", uuid).Updates(Job{UUID: j.UUID, State: j.State, UpdatedAt: time.Now(), Locker: j.Locker})
	if db.Error != nil {
		return &Job{}, db.Error
	}

	// This is the display the updated Job
	err := db.Debug().Model(Job{}).Where("id = ?", uuid).Preload("Targets").Preload("Resource").Take(&j).Error
	if err != nil {
		return &Job{}, err
	}
	return j, nil
}

func (j *Job) DeleteAJob(db *gorm.DB, uuid uuid.UUID) (int64, error) {

	// db = db.Debug().Model(&Job{}).Where("id = ?", uid).Take(&Job{}).Delete(&Job{}) // debug only
	// delete targets first
	// db = db.Select(j.Targets).Delete(&Job{ID: uuid})
	// delete targets first
	db = db.Model(&Target{}).Where("job_id = ?", uuid).Delete(&Target{})
	// delete job
	db = db.Model(&Job{}).Where("id = ?", uuid).Take(&Job{}).Delete(&Job{})
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (jg *JobGroup) SaveJobGroup(db *gorm.DB) (*JobGroup, error) {
	jg.BeforeCreate(db)
	err := db.Debug().Create(&jg).Error
	if err != nil {
		return &JobGroup{}, err
	}
	return jg, nil
}

func (jg *JobGroup) UpdateAJobGroup(db *gorm.DB, uuid uuid.UUID) (*JobGroup, error) {
	db = db.Debug().Model(&Job{}).Where("id = ?", uuid).Updates(JobGroup{Jobs: jg.Jobs})
	if db.Error != nil {
		return &JobGroup{}, db.Error
	}
	return jg, nil
}

func (jg *JobGroup) FindJobGroupByUUID(db *gorm.DB, uuid uuid.UUID) (*JobGroup, error) {
	// err := db.Debug().Model(JobGroup{}).Where("id = ?", uuid).Preload("Jobs").Take(&j).Error
	err := db.Preload("Jobs.Resource.Conditions").Find(&jg).Where("id = ?", uuid).Error
	if err != nil {
		return &JobGroup{}, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &JobGroup{}, errors.New("JobGroup Not Found")
	}
	return jg, err
}

func (jg *JobGroup) DeleteAJobGroup(db *gorm.DB, uuid uuid.UUID) (int64, error) {
	err := db.Model(&JobGroup{}).Where("id = ?", uuid).Association("Jobs").Clear()
	if err != nil {
		return 0, err
	}
	// delete job
	db = db.Model(&JobGroup{}).Where("id = ?", uuid).Take(&Job{}).Delete(&JobGroup{})
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
