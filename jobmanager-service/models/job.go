package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type State int
type JobType int

const (
	Created State = iota + 1
	Started
	Progressing
	Finished
	Failed

	CreateDeployment JobType = iota + 1
	GetDeployment
	DeleteDeployment
)

// TODO: this Job is pulled by the drivers, we should agree on Jobs model
type Job struct {
	UUID           uuid.UUID      `json:"uuid"`
	Type           JobType        `json:"type"`
	State          State          `json:"state"`
	AppDescription AppDescription `json:"component"` // will be an array in the future
	Targets        []Target       // array of targets where the AppDescription is applied
	// Policies?
	// Requirements?
}

type Target struct {
	// TODO UPC
}

type Metadata struct {
	Name   string            `yaml:"name"`
	Labels map[string]string `yaml:"labels"`
}

type Selector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}

type Container struct {
	Name      string    `yaml:"name"`
	Image     string    `yaml:"image"`
	Command   []string  `yaml:"command"`
	Args      []string  `yaml:"args"`
	Resources Resources `yaml:"resources"`
}

type Resources struct {
	Requests map[string]string `yaml:"requests"`
	Limits   map[string]string `yaml:"limits"`
}

type TemplateSpec struct {
	Containers []Container `yaml:"containers"`
}

type Template struct {
	Metadata     Metadata
	TemplateSpec TemplateSpec
}

type Spec struct {
	Replicas int      `yaml:"replicas"`
	Selector Selector `yaml:"selector"`
	Template Template `yaml:"template"`
}

type AppDescription struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Jobs []struct {
	Job Job
}

func StateIsValid(value int) bool {
	return int(Created) >= value && value <= int(Failed)
}

func JobTypeIsValid(value int) bool {
	return int(CreateDeployment) >= value && value <= int(DeleteDeployment)
}

func (j *Job) SaveJob(db *gorm.DB) (*Job, error) {

	err := db.Debug().Create(&j).Error
	if err != nil {
		return &Job{}, err
	}
	return j, nil
}

func (j *Job) FindJobByUUID(db *gorm.DB, uuid uuid.UUID) (*Job, error) {
	err := db.Debug().Model(Job{}).Where("id = ?", uuid).Take(&j).Error
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
	err = db.Debug().Model(&Job{}).Limit(100).Find(&jobs).Error
	if err != nil {
		return &[]Job{}, err
	}
	return &jobs, err
}

func (u *Job) FindJobsByState(db *gorm.DB, state int) (*[]Job, error) {
	var err error
	jobs := []Job{}
	err = db.Debug().Model(&Job{}).Where("state = ?", state).Limit(100).Find(&jobs).Error
	if err != nil {
		return &[]Job{}, err
	}
	return &jobs, err
}

func (j *Job) UpdateAJob(db *gorm.DB, uuid uuid.UUID) (*Job, error) {
	db = db.Debug().Model(&Job{}).Where("id = ?", uuid).Take(&Job{}).UpdateColumns(
		map[string]interface{}{
			"state":      j.State,
			"created_at": time.Now(),
		},
	)
	if db.Error != nil {
		return &Job{}, db.Error
	}

	// This is the display the updated Job
	err := db.Debug().Model(&Job{}).Where("id = ?", uuid).Take(&j).Error
	if err != nil {
		return &Job{}, err
	}
	return j, nil
}

func (j *Job) DeleteAJob(db *gorm.DB, uuid uuid.UUID) (int64, error) {

	// db = db.Debug().Model(&Job{}).Where("id = ?", uid).Take(&Job{}).Delete(&Job{}) // debug only
	db = db.Model(&Job{}).Where("id = ?", uuid).Take(&Job{}).Delete(&Job{})

	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
