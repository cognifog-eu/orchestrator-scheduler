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
package models

import (
	"time"
)

// Resource entity
type Resource struct {
	ID           string `gorm:"type:char(36);primaryKey" json:"resource_uuid,omitempty"` // we set this field manually
	JobID        string `gorm:"type:char(36);not null" json:"job_id" validate:"omitempty,uuid4"`
	ResourceName string `gorm:"type:text" json:"resource_name,omitempty" validate:"omitempty"`
	//OriginalResourceName string      `gorm:"type:text" json:"-" validate:"omitempty"`
	Conditions   []Condition   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"conditions,omitempty" validate:"dive"`
	Remediations []Remediation `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"remediations,omitempty" validate:"omitempty"`
}

// Condition entity
type Condition struct {
	BaseUINT           `json:"-"`
	ResourceID         string          `gorm:"type:char(36);index" json:"-" validate:"omitempty,uuid4"`
	Type               ResourceState   `gorm:"type:text" json:"type" validate:"required"`
	Status             ConditionStatus `gorm:"type:text" json:"status" validate:"required"`
	ObservedGeneration int64           `gorm:"type:bigint" json:"observedGeneration,omitempty" validate:"omitempty"`
	LastTransitionTime time.Time       `gorm:"type:timestamp" json:"lastTransitionTime" validate:"required"`
	Reason             string          `gorm:"type:text" json:"reason" validate:"required"`
	Message            string          `gorm:"type:text" json:"message" validate:"required"`
}
