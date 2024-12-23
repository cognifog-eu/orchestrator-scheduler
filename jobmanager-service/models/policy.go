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

// Policy-related entities
type Policy struct {
	BaseUINT
	InstructionID string          `gorm:"type:char(36);not null" json:"-" validate:"omitempty,uuid4"`
	Name          string          `gorm:"-" json:"name"`
	Component     string          `gorm:"-" json:"component"`
	FromTemplate  string          `gorm:"-" json:"fromTemplate,omitempty"`
	Spec          *PolicySpec     `gorm:"-" json:"spec,omitempty"`
	Remediation   string          `gorm:"-" json:"remediation,omitempty"`
	Variables     PolicyVariables `gorm:"-" json:"variables"`
}

type PolicySpec struct {
	Expr       string     `json:"expr"`
	Thresholds Thresholds `json:"thresholds"`
}

type Thresholds struct {
	Warning  int `json:"warning"`
	Critical int `json:"critical"`
}

type PolicyVariables struct {
	ThresholdTimeSeconds int    `json:"thresholdTimeSeconds"`
	CompssTask           string `json:"compssTask"`
}
