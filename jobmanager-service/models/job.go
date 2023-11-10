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
	Progressing
	Available
	Degraded

	CreateDeployment JobType = iota + 1
	GetDeployment
	DeleteDeployment
)

// TODO: this Job is pulled by the drivers, we should agree on Jobs model
type Job struct {
	// gorm.Model
	ID        uuid.UUID `gorm:"type:char(36);primary_key"` // lets abstract this id from the shell user -> TODO: should be uuid
	UUID      uuid.UUID `gorm:"type:text" json:"uuid"`     // optional and unique across all icos
	Type      JobType   `gorm:"type:text" json:"type"`
	State     State     `gorm:"type:text" json:"state"`
	Manifest  string    `gorm:"type:text" json:"manifest"`
	Targets   []Target  `json:"targets"` // array of targets where the Manifest is applied
	Locker    *bool     `json:"locker"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// hold information that N jobs share (N jobs needed to provide application x)
type JobGroup struct {
	AppName        string `json:"appName"`
	AppDescription string `json:"appDescription"`
}

func (job *Job) BeforeCreate(tx *gorm.DB) (err error) {
	// UUID version 4
	job.ID = uuid.New()
	b := new(bool)
	*b = false
	job.Locker = b
	return
}

type Target struct {
	ID          uint32 `gorm:"primary_key" json:"id"`
	JobID       uuid.UUID
	ClusterName string `json:"cluster_name"`
	NodeName    string `json:"node_name"`
	// what we need to know about peripherals
	// TODO UPC&AGGREGATOR
}

func StateIsValid(value int) bool {
	return int(Created) >= value && value <= int(Degraded)
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
	err := db.Debug().Model(Job{}).Where("id = ?", uuid).Preload("Targets").Take(&j).Error
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
	err = db.Debug().Model(&Job{}).Preload("Targets").Find(&jobs).Error
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
	err = db.Debug().Preload("Targets").Find(&jobs, "state =? AND locker = FALSE OR state =? AND locker = TRUE AND updated_at < ?", int(Created), int(Progressing), time.Now().Local().Add(time.Second*time.Duration(-300))).Error
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
	j.NewJobTTL()
	db = db.Debug().Model(&Job{}).Where("id = ?", uuid).Updates(Job{UUID: j.UUID, State: j.State, UpdatedAt: time.Now(), Locker: j.Locker})
	// db = db.Debug().Model(&Job{}).Where("id = ?", uuid).Take(&Job{}).UpdateColumns(
	// 	map[string]interface{}{
	// 		"state":      j.State,
	// 		"updated_at": time.Now(),
	// 		"locker":     j.Locker, // TODO: this is not OK! - idempotency? how many times can I unlock/lock a job?
	// 	},
	// )
	if db.Error != nil {
		return &Job{}, db.Error
	}

	// This is the display the updated Job
	err := db.Debug().Model(Job{}).Where("id = ?", uuid).Preload("Targets").Take(&j).Error
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
