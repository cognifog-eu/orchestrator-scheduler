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

// Enum-like Types
type (
	RemediationType   string
	ResourceState     string
	ConditionStatus   string
	JobState          string
	JobType           string
	OrchestratorType  string
	Operation         string
	RemediationStatus string
)

// RemediationType Enum
const (
	ScaleUp    RemediationType = "scale-up"
	ScaleDown  RemediationType = "scale-down"
	ScaleOut   RemediationType = "scale-out"
	ScaleIn    RemediationType = "scale-in"
	Patch      RemediationType = "patch"
	Reallocate RemediationType = "reallocate"
	Replace    RemediationType = "replace"
	Secure     RemediationType = "secure"
)

// ConditionStatus Enum
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// OrchestratorType Enum
const (
	OCM   OrchestratorType = "ocm"
	Nuvla OrchestratorType = "nuvla"
	None  OrchestratorType = ""
)

// JobState Enum
const (
	Created     JobState = "Created"
	Progressing JobState = "Progressing"
	Finished    JobState = "Finished"
	Degraded    JobState = "Degraded"
)

// JobType Enum
const (
	CreateDeployment JobType = "CreateDeployment"
	DeleteDeployment JobType = "DeleteDeployment"
	UpdateDeployment JobType = "UpdateDeployment"
	//ReplaceDeployment JobType = "ReplaceDeployment"
)

// ResourceState Enum
const (
	ResourceProgressing ResourceState = "Progressing"
	ResourceApplied     ResourceState = "Applied"
	ResourceAvailable   ResourceState = "Available"
	ResourceDegraded    ResourceState = "Degraded"
)

const (
	Pending    RemediationStatus = "Pending"
	Failed     RemediationStatus = "Failed"
	Remediated RemediationStatus = "Remediated"
)

// OrchestratorType Enum Mapper
func OrchestratorTypeMapper(orchestratorType string) OrchestratorType {
	switch orchestratorType {
	case string(Nuvla):
		return Nuvla
	case string(OCM):
		return OCM
	default:
		return None
	}
}
