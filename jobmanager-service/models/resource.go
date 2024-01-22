package models

import (
	"errors"
	"icos/server/jobmanager-service/utils/logs"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ResourceState int
type ConditionStatus string

const (
	Progressing ResourceState = iota + 1
	Applied
	Available
	Degraded
)

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type Resource struct {
	// gorm.Model
	ID           uuid.UUID `gorm:"type:char(36);primary_key"` // unique across all icos
	JobID        uuid.UUID `json:"job_id"`
	ResourceUUID uuid.UUID `gorm:"type:char(36)" json:"resource_uuid,omitempty"`
	ResourceName string    `gorm:"type:text" json:"resource_name"`
	// Target       Target    `json:"node_target"`
	// Status    Status    `gorm:"foreignkey:ResourceID;" json:"status"`
	Conditions []Condition `gorm:"foreignkey:ResourceID;" json:"conditions,omitempty"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}

func (Resource *Resource) BeforeCreate(tx *gorm.DB) (err error) {
	// UUID version 4
	Resource.ID = uuid.New()
	return
}

// type Status struct {
// 	ID         uint32 `gorm:"primary_key" json:"id"`
// 	ResourceID uuid.UUID
// 	Conditions []Condition `json:"conditions,omitempty"`
// }

type Condition struct {
	ID         uint32 `gorm:"primary_key" json:"id"`
	ResourceID uuid.UUID
	// type of condition in CamelCase or in foo.example.com/CamelCase.
	// ---
	// Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be
	// useful (see .node.status.conditions), the ability to deconflict is important.
	// The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)
	// +required
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// status of the condition, one of True, False, Unknown.
	// +required
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status"`
	// observedGeneration represents the .metadata.generation that the condition was set based upon.
	// For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
	// with respect to the current state of the instance.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,3,opt,name=observedGeneration"`
	// lastTransitionTime is the last time the condition transitioned from one status to another.
	// This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.
	// +required
	LastTransitionTime time.Time `json:"lastTransitionTime" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// reason contains a programmatic identifier indicating the reason for the condition's last transition.
	// Producers of specific condition types may define expected values and meanings for this field,
	// and whether the values are considered a guaranteed API.
	// The value should be a CamelCase string.
	// This field may not be empty.
	// +required
	Reason string `json:"reason" protobuf:"bytes,5,opt,name=reason"`
	// message is a human readable message indicating details about the transition.
	// This may be an empty string.
	// +required
	Message string `json:"message" protobuf:"bytes,6,opt,name=message"`
}

func (r *Resource) UpdateAResource(db *gorm.DB, jobId, uuid uuid.UUID) (*Resource, error) {
	logs.Logger.Println("Updating the resource: " + r.ID.String())
	db = db.Session(&gorm.Session{FullSaveAssociations: true}).Where("job_id = ?", jobId).Updates(&Resource{ResourceUUID: r.ResourceUUID, Conditions: r.Conditions, ResourceName: r.ResourceName})
	if db.Error != nil {
		return &Resource{}, db.Error
	}

	// This is the display the updated Job
	err := db.Debug().Model(Resource{}).Where("job_id = ?", jobId).Preload("Conditions").Take(&r).Error
	if err != nil {
		return &Resource{}, err
	}
	return r, nil
}

func (r *Resource) AddCondition(db *gorm.DB, condition *Condition) (*Resource, error) {
	// create condition
	logs.Logger.Println("Updating the resource: " + r.ID.String())
	// db = db.Session(&gorm.Session{FullSaveAssociations: true}).Where("id = ?", r.ID).Updates(&Resource{Conditions: r.Conditions})
	err := db.Debug().Create(&condition)
	if err != nil {
		return &Resource{}, db.Error
	}
	// db = db.Session(&gorm.Session{FullSaveAssociations: true}).Where("id = ?", r.ID).Updates(Resource{Status: r.Status})

	// This is the display the updated Job
	err = db.Debug().Model(Resource{}).Where("resource_uuid =?", r.ResourceUUID).Preload("Conditions").Take(&r)
	return r, err.Error
}

func (r *Resource) RemoveConditions(db *gorm.DB) (*Resource, error) {
	// create condition
	logs.Logger.Println("Updating the resource: " + r.ID.String())
	// db = db.Session(&gorm.Session{FullSaveAssociations: true}).Where("id = ?", r.ID).Updates(&Resource{Conditions: r.Conditions})
	// remove existing status first
	logs.Logger.Println("Removing old status of the resource: " + r.ID.String())
	conds := []Condition{}
	err := db.Debug().Model(&Condition{}).Where("resource_id =?", r.ID).Find(&conds).Error
	err = db.Delete(&conds).Error
	if err != nil {
		return &Resource{}, db.Error
	}
	return r, err
}

func (r *Resource) SaveResource(db *gorm.DB) (*Resource, error) {
	err := db.Debug().Create(&r).Error
	if err != nil {
		return &Resource{}, err
	}
	return r, nil
}

func (r *Resource) FindResourceByUUID(db *gorm.DB, uuid uuid.UUID) (*Resource, error) {
	err := db.Debug().Model(Resource{}).Where("job_id = ?", uuid).Preload("Conditions").Take(&r).Error
	if err != nil {
		return &Resource{}, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &Resource{}, errors.New("Job Not Found")
	}
	return r, err
}
