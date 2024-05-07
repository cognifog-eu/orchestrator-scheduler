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
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
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
	CreateNamespace
)

var JobTypeFromString = map[string]JobType{
	"CreateDeployment": CreateDeployment,
	"GetDeployment":    GetDeployment,
	"DeleteDeployment": DeleteDeployment,
	"RecoveryJob":      RecoveryJob,
	"CreateNamespace":  CreateNamespace,
}

type JobGroupHeader struct {
	Name        string      `json:"name"`
	Namespace   string      `json:"namespace"`
	Description string      `json:"description"`
	Components  []Component `json:"components"`
}

// type Manifest struct {
// 	Name string `json:"name"`
// }

// hold information that N jobs share (N jobs needed to provide application x)
type JobGroup struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	AppName        string    `json:"appName"`
	AppDescription string    `json:"appDescription"`
	Jobs           []Job     `json:"jobs"`
}

// TODO: this Job is pulled by the drivers, we should agree on Jobs model
type Job struct {
	// gorm.Model
	ID                  uuid.UUID        `gorm:"type:char(36);primaryKey"`      // unique across all ecosystem
	UUID                uuid.UUID        `gorm:"type:text" json:"uuid"`         // optional and unique across all ecosystem
	JobGroupID          uuid.UUID        `gorm:"type:text" json:"job_group_id"` // unique across all ecosystem
	JobGroupName        string           `gorm:"type:text" json:"job_group_name"`
	JobGroupDescription string           `gorm:"type:text" json:"job_group_description,omitempty"`
	Type                JobType          `gorm:"type:text" json:"type"`
	State               JobState         `gorm:"type:text" json:"state"`
	Manifests           []PlainManifests `json:"manifests"`
	Manifest            string           `gorm:"type:text" json:"manifest"`
	Targets             []Target         `json:"targets"` // array of targets where the Manifest is applied
	Locker              *bool            `json:"locker"`
	Orchestrator        OrchestratorType `gorm:"type:text" json:"orchestrator"` // identifies the orchestrator that can execute the job based on target provided by MM
	UpdatedAt           time.Time        `json:"updated_at"`
	Resource            Resource         `gorm:"foreignkey:JobID;" json:"resource,omitempty"`
	Namespace           string           `gorm:"type:text" json:"namespace"`
}

// type MatchmakingResWrapper struct {
// 	Components []Component `json:"components"`
// }

type Component struct {
	Name      string     `json:"name"`
	Type      string     `json:"type,omitempty"`
	Manifests []Manifest `json:"manifests,omitempty"`
	Targets   []Target   `json:"targets,omitempty"`
}

type PlainManifests struct {
	ID         uint32    `gorm:"primaryKey" json:"-"`
	JobID      uuid.UUID `json:"-"`
	YamlString string    `json:"yamlString"`
}

type Manifest struct {
	Name string `json:"name"`
}

type K8sMapper struct {
	ApiVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   K8sMapperMetadata `json:"metadata"`
}

type K8sMapperMetadata struct {
	Name      string            `json:"name,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels"`
}

type Target struct {
	ID           uint32           `gorm:"primaryKey" json:"-"`
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

// func StateIsValid(value int) bool {
// 	return int(JobCreated) >= value && value <= int(Degraded)
// }

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

func (j *Job) FindJobsToExecute(db *gorm.DB, orch string) (*[]Job, error) {
	var err error
	jobs := []Job{}
	err = db.Debug().Preload("Targets").Preload("Resource").Find(&jobs, "state =? AND locker = FALSE AND orchestrator =? OR state =? AND locker = TRUE AND orchestrator =? AND updated_at < ?", int(JobCreated), orch, Progressing, orch, time.Now().Local().Add(time.Second*time.Duration(-300))).Error
	db.Logger.LogMode(logger.Info)
	// err = db.Debug().Model(&Job{}).Where(db.Where("state = ?", int(Created)).Where("locker = ?", false)).
	// 	Or(db.Where("state = ?", int(Progressing)).Where("locker = ?", true)).Where("updated_at < ?", time.Now().Local().Add(time.Second*time.Duration(-300))).
	// 	Preload("Targets").Find(&jobs).Error
	if err != nil {
		return &[]Job{}, err
	}
	return &jobs, err
}

func (j *Job) UpdateAJob(db *gorm.DB) (*Job, error) {
	// trigger TTL ticker on each writing access except the CreateJob
	// logs.Logger.Println("Setting new TTL for the Job before update: " + j.ID.String())
	// j.NewJobTTL()
	db = db.Debug().Model(&Job{}).Where("id = ?", j.ID).Updates(
		Job{UUID: j.UUID, State: j.State, Manifest: j.Manifest, Orchestrator: j.Orchestrator, UpdatedAt: time.Now()})
	if db.Error != nil {
		return &Job{}, db.Error
	}

	// This is the display the updated Job
	err := db.Debug().Model(Job{}).Where("id = ?", j.ID).Preload("Targets").Preload("Resource").Take(&j).Error
	if err != nil {
		return &Job{}, err
	}
	return j, nil
}

func (j *Job) JobLocker(db *gorm.DB) (*Job, error) {
	// trigger TTL ticker on each writing access except the CreateJob
	logs.Logger.Println("Setting new TTL for the Job before update: " + j.ID.String())
	// j.NewJobTTL()
	db = db.Debug().Model(&Job{}).Where("id = ?", j.ID).Updates(
		Job{Locker: j.Locker, UpdatedAt: time.Now()})
	if db.Error != nil {
		return &Job{}, db.Error
	}

	// This is the display the updated Job
	err := db.Debug().Model(Job{}).Where("id = ?", j.ID).Preload("Targets").Preload("Resource").Take(&j).Error
	if err != nil {
		return &Job{}, err
	}
	return j, nil
}

func (j *Job) DeleteAJob(db *gorm.DB) (int64, error) {

	// db = db.Debug().Model(&Job{}).Where("id = ?", uid).Take(&Job{}).Delete(&Job{}) // debug only
	// delete targets first
	// db = db.Select(j.Targets).Delete(&Job{ID: uuid})
	// delete targets first
	db = db.Model(&Target{}).Where("job_id = ?", j.ID).Delete(&Target{})
	// delete job
	db = db.Model(&Job{}).Where("id = ?", j.ID).Take(&Job{}).Delete(&Job{})
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
	var app *JobGroup
	logs.Logger.Println("Querying for JobGroup's ID: " + uuid.String())
	// err := db.Preload("Jobs.Resource.Conditions").Find(&app).Where("id = ?", uuid).Error
	err := db.Preload(clause.Associations).Preload("Jobs."+clause.Associations).Preload("Jobs.Resource.Conditions").Find(&app).Where("id = ?", uuid).Error
	db.Logger.LogMode(logger.Info)
	if err != nil {
		return &JobGroup{}, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &JobGroup{}, errors.New("JobGroup Not Found")
	}
	return app, err
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

func (jg *JobGroup) FindAllJobGroups(db *gorm.DB) (*[]JobGroup, error) {
	var err error
	jobGroups := []JobGroup{}
	// err = db.Preload("Jobs.Resources.Conditions").Preload().Find(&jobGroups).Error
	err = db.Preload(clause.Associations).Preload("Jobs." + clause.Associations).Preload("Jobs.Resource.Conditions").Find(&jobGroups).Error
	// err = db.Model(&Menu{}).Where("pid = ?", 0).Preload(clause.Associations, preload).Find(&views).Error
	if err != nil {
		return &[]JobGroup{}, err
	}
	return &jobGroups, err
}

func (jg *JobGroup) UpdateJobGroupState(db *gorm.DB, uuid uuid.UUID) (*JobGroup, error) {
	db = db.Debug().Model(&Job{}).Where("id = ?", uuid).Updates(JobGroup{Jobs: jg.Jobs})
	if db.Error != nil {
		return &JobGroup{}, db.Error
	}
	return jg, nil
}
