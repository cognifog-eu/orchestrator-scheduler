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
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Incompliance entity
type Remediation struct {
	BaseUUID
	RemediationType   RemediationType    `gorm:"type:text" json:"remediationType" validate:"required"`
	Status            RemediationStatus  `gorm:"type:text" default:"Pending" json:"remediationStatus" validate:"required"`
	RemediationTarget *RemediationTarget `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"remediationTarget,omitempty" validate:"omitempty"`
	ResourceID        string             `gorm:"type:char(36);not null" json:"resource_id" validate:"omitempty,uuid4"`
}

type RemediationTarget struct {
	BaseUUID
	RemediationID string `gorm:"type:char(36);not null" json:"remediation_id" validate:"omitempty,uuid4"`
	Container     string `gorm:"type:text" json:"container,omitempty" validate:"omitempty"`
	PodUID        string `gorm:"type:text" json:"pod_uid,omitempty" validate:"omitempty"`
	Pod           string `gorm:"type:text" json:"pod,omitempty" validate:"omitempty"`
	Node          string `gorm:"type:text" json:"node,omitempty" validate:"omitempty"`
	Namespace     string `gorm:"type:text" json:"namespace,omitempty" validate:"omitempty"`
	Command       string `gorm:"type:text" json:"command,omitempty" validate:"omitempty"`
}

type StringMap map[string]string

// ExtraLabels Mapper
func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}
