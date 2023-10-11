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
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;primary_key"` // lets abstract this id from the shell user -> TODO: should be uuid
	UUID     uuid.UUID `gorm:"type:text" json:"uuid"` // optional and unique across all icos
	Type     JobType   `gorm:"type:text" json:"type"`
	State    State     `gorm:"type:text" json:"state"`
	Manifest string    `gorm:"type:text" json:"manifest"`
	// // Manifest *workv1.Manifest `json:"manifest"` // Can be used instead
	// Manifest struct {
	// APIVersion string `json:"apiVersion"`
	// Kind       string `json:"kind"`
	// Metadata   struct {
	// 	Name   string            `json:"name"`
	// 	Labels map[string]string `json:"labels"`
	// } `gorm:"type:text" json:"metadata"`
	// Spec struct {
	// 	Replicas int `json:"replicas"`
	// 	Selector struct {
	// 		MatchLabels map[string]string `json:"matchLabels"`
	// 	} `gorm:"type:text" json:"selector"`
	// 	Template struct {
	// 		Metadata struct {
	// 			Name   string            `json:"name"`
	// 			Labels map[string]string `json:"labels"`
	// 		} `gorm:"type:text" json:"metadata"`
	// 		TemplateSpec struct {
	// 			Container []struct {
	// 				Name      string   `json:"name"`
	// 				Image     string   `json:"image"`
	// 				Command   []string `json:"command"`
	// 				Args      []string `json:"args"`
	// 				Resources struct {
	// 					Requests map[string]string `gorm:"type:text"  json:"requests"`
	// 					Limits   map[string]string `gorm:"type:text"  json:"limits"`
	// 				} `gorm:"type:text" json:"resources"`
	// 			} `gorm:"type:text" json:"containers"`
	// 		} `gorm:"type:text"`
	// 	} `gorm:"type:text" json:"template"`
	// } `gorm:"type:text" json:"spec"`
	// } `gorm:"type:text" json:"manifest"` // will be an array in the future
	Targets []Target `gorm:"type:text" json:"targets"` // array of targets where the Manifest is applied
	// Policies?
	// Requirements?
	Locker bool `gorm:"default:false" json:"locker"`
}

// hold information that N jobs share (N jobs needed to provide application x)
type JobGroup struct {
	AppName        string `json:"appName"`
	AppDescription string `json:"appDescription"`
}

func (job *Job) BeforeCreate(tx *gorm.DB) (err error) {
	// UUID version 4
	job.ID = uuid.New()
	return
}

// type Manifest struct {
// 	gorm.Model
// 	ID         uint32 `gorm:"primary_key"`
// 	APIVersion string `json:"apiVersion"`
// 	Kind       string `json:"kind"`
// 	Metadata   struct {
// 		Name   string            `json:"name"`
// 		Labels map[string]string `json:"labels"`
// 	} `json:"metadata"`
// 	Spec struct {
// 		Replicas int `json:"replicas"`
// 		Selector struct {
// 			gorm.Model
// 			ID          uint32            `gorm:"primary_key"`
// 			MatchLabels map[string]string `json:"matchLabels"`
// 		} `json:"selector"`
// 		Template struct {
// 			gorm.Model
// 			ID       uint32 `gorm:"primary_key"`
// 			Metadata struct {
// 				Name   string            `json:"name"`
// 				Labels map[string]string `json:"labels"`
// 			} `json:"metadata"`
// 			TemplateSpec struct {
// 				Containers []Container `json:"containers"`
// 			}
// 		} `json:"template"`
// 	} `json:"spec"`
// }

// type Metadata struct {
// 	gorm.Model
// 	ID     uint32            `gorm:"primary_key"`
// 	Name   string            `json:"name"`
// 	Labels map[string]string `json:"labels"`
// }

// type Spec struct {
// 	gorm.Model
// 	ID       uint32   `gorm:"primary_key"`
// 	Replicas int      `json:"replicas"`
// 	Selector Selector `json:"selector"`
// 	Template Template `json:"template"`
// }

// type Selector struct {
// 	gorm.Model
// 	ID          uint32            `gorm:"primary_key"`
// 	MatchLabels map[string]string `json:"matchLabels"`
// }

// type Template struct {
// 	gorm.Model
// 	ID           uint32 `gorm:"primary_key"`
// 	Metadata     Metadata
// 	TemplateSpec TemplateSpec
// }

type Target struct {
	gorm.Model
	ID          uint32 `gorm:"primary_key" json:"id"`
	ClusterName string `json:"cluster_name"`
	Hostname    string `json:"node_name"`
	// what we need to know about targets -> ocm: cluster-id; nuvla: infra-service-uuid
	// at least:
	// cluster-id: string. Represents a cluster
	// infra-service-uuid: string.  Represents a cluster
	//

	// what we need to know about peripherals
	// TODO UPC
}

// type TemplateSpec struct {
// 	gorm.Model
// 	ID         uint32      `gorm:"primary_key"`
// 	Containers []Container `json:"containers"`
// }

// type Container struct {
// 	gorm.Model
// 	ID        uint32    `gorm:"primary_key"`
// 	Name      string    `json:"name"`
// 	Image     string    `json:"image"`
// 	Command   []string  `json:"command"`
// 	Args      []string  `json:"args"`
// 	Resources Resources `json:"resources"`
// }

// type Resources struct {
// 	gorm.Model
// 	ID       uint32            `gorm:"primary_key"`
// 	Requests map[string]string `gorm:"type:text"  json:"requests"`
// 	Limits   map[string]string `gorm:"type:text"  json:"limits"`
// }

func StateIsValid(value int) bool {
	return int(Created) >= value && value <= int(Failed)
}

func JobTypeIsValid(value int) bool {
	return int(CreateDeployment) >= value && value <= int(DeleteDeployment)
}

func (j *Job) NewJobTTL() {
	if time.Now().Unix()-j.UpdatedAt.Unix() > int64(300) {
		j.Locker = false
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

func (j *Job) FindJobsByState(db *gorm.DB, state int) (*[]Job, error) {
	var err error
	jobs := []Job{}
	err = db.Debug().Model(&Job{}).Where("state = ?", state).Limit(100).Find(&jobs).Error
	if err != nil {
		return &[]Job{}, err
	}
	return &jobs, err
}

func (j *Job) FindJobsToExecute(db *gorm.DB) (*[]Job, error) {
	var err error
	jobs := []Job{}
	err = db.Debug().Model(&Job{}).Where(db.Where("state = ?", int(Created)).Where("locker = ?", false)).
		Or(db.Where("state = ?", int(Progressing)).Where("locker = ?", true)).Where("updated_at < ?", time.Now().Local().Add(time.Second*time.Duration(-300))).
		Limit(100).Find(&jobs).Error
	if err != nil {
		return &[]Job{}, err
	}
	return &jobs, err
}

func (j *Job) UpdateAJob(db *gorm.DB, uuid uuid.UUID) (*Job, error) {
	// trigger TTL ticker on each writing access except the CreateJob
	j.NewJobTTL()
	db = db.Debug().Model(&Job{}).Where("id = ?", uuid).Take(&Job{}).UpdateColumns(
		map[string]interface{}{
			"state":      j.State,
			"updated_at": time.Now(),
			"locker":     j.Locker, // TODO: this is not OK! - idempotency? how many times can I unlock/lock a job?
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
