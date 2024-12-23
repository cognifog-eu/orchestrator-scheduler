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
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JobGroup entity
type JobGroup struct {
	BaseUUID
	AppName        string `gorm:"type:text" json:"appName"`
	AppDescription string `gorm:"type:text" json:"appDescription"`
	Jobs           []Job  `gorm:"constraint:OnDelete:CASCADE;" json:"jobs" validate:"dive,required"`
}

func (jg *JobGroup) Validate() error {
	return validate.Struct(jg)
}

func (jg *JobGroup) AfterCreate(tx *gorm.DB) (err error) {
	return jg.Validate()
}

func (jg *JobGroup) BeforeUpdate(tx *gorm.DB) (err error) {
	return jg.Validate()
}

func CreateJobGroupFromAllocation(allocationDTO *AllocationDTO) *JobGroup {
	jobGroup := &JobGroup{
		AppName:        allocationDTO.Name + "-" + uuid.New().String(),
		AppDescription: allocationDTO.Description,
	}

	if jobGroup.AppName == "" {
		jobGroup.AppName = uuid.New().String()
	}

	return jobGroup
}
