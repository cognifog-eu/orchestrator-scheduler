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
	"gorm.io/gorm"
)

// Job entity
type Job struct {
	BaseUUID
	JobGroupID   string           `gorm:"type:char(36);not null" json:"job_group_id" validate:"omitempty,uuid4"`
	OwnerID      string           `gorm:"type:char(36);default:''" json:"owner_id" validate:"omitempty,uuid4"`
	Type         JobType          `gorm:"type:text" json:"type"`
	SubType      RemediationType  `gorm:"type:text" json:"sub_type,omitempty"`
	State        JobState         `gorm:"type:text" json:"state"`
	Target       Target           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"targets,omitempty" validate:"omitempty"`
	Orchestrator OrchestratorType `gorm:"type:text" json:"orchestrator"`
	Instruction  *Instruction     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"instruction,omitempty"`
	Resource     *Resource        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"resource,omitempty"`
	Namespace    string           `gorm:"type:text" json:"namespace,omitempty" validate:"omitempty"`
}

func (j *Job) Validate() error {
	return validate.Struct(j)
}

func (j *Job) AfterCreate(tx *gorm.DB) (err error) {
	return j.Validate()
}

func (j *Job) BeforeUpdate(tx *gorm.DB) (err error) {
	return j.Validate()
}
