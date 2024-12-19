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
)

// TODO: Refactor the DTOs once other services change their models to fit the new Job Manager model

// AllocationDTO represents the allocation data transfer object used for matchmaking
type AllocationDTO struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Components  []ComponentDTO           `json:"components"`
	Policies    []Policy                 `json:"policies"`
	Manifests   []map[string]interface{} `json:"manifests"`
}

// TargetDTO represents the target information for deployment
type TargetDTO struct {
	ClusterName  string `json:"clustername" yaml:"clustername"`
	NodeName     string `json:"nodename" yaml:"nodename"`
	Orchestrator string `json:"orchestrator" yaml:"orchestrator"`
}

// AllocationOptionDTO defines a function type used for applying options to an AllocationDTO
type AllocationOptionDTO func(*AllocationDTO)

// ComponentDTO represents a component in an allocation
type ComponentDTO struct {
	Name        string           `json:"name" yaml:"name"`
	Type        string           `json:"type" yaml:"type"`
	Requirement Requirement      `json:"requirements,omitempty" yaml:"requirements,omitempty"`
	Policies    []Policy         `json:"policies,omitempty" yaml:"policies,omitempty"`
	Manifests   []ManifestRefDTO `json:"manifests" yaml:"manifests"`
	Target      interface{}      `json:"targets,omitempty" yaml:"targets,omitempty"`
}

// JobOwnershipDTO represents the ownership details of a job
type JobOwnershipDTO struct {
	OwnerID string `json:"owner_id"`
}

// ManifestRefDTO represents a reference to a manifest with its name
type ManifestRefDTO struct {
	Name string `gorm:"type:text" json:"name" yaml:"name"`
}

// Notification is a DTO for sending notifications to the Police Manager
type Notification struct {
	AppInstance  string `json:"app_instance"`
	CommonAction Action `json:"common_action"`
	Service      string `json:"service"`
	Manifest     string `json:"app_descriptor"`
}

type Action struct {
	URI                string            `json:"uri"`
	HTTPMethod         string            `json:"http_method"`
	IncludeAccessToken bool              `json:"include_access_token"`
	ExtraParameters    map[string]string `json:"extra_parameters"`
}

type ExtraParameters struct {
	JobGroupId uuid.UUID `json:"job_group_id"`
}
