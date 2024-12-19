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

// Target entity
type Target struct {
	BaseUINT
	JobID        string           `gorm:"type:char(36);index;not null" json:"-" validate:"omitempty,uuid4"`
	ClusterName  string           `json:"cluster_name" validate:"required"`
	NodeName     string           `json:"node_name,omitempty" validate:"omitempty"`
	Orchestrator OrchestratorType `gorm:"type:text" json:"orchestrator" validate:"required"`
}
