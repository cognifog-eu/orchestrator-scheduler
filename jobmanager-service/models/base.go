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

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Metadata for the database entities
type Metadata struct {
	CreatedAt time.Time `json:"-" yaml:"-"`
	UpdatedAt time.Time `json:"-" yaml:"-"`
}

// Base entities with UUID and UINT
type BaseUUID struct {
	Metadata
	ID string `gorm:"type:char(36);primary_key;" yaml:"-"`
}

type BaseUINT struct {
	Metadata
	ID uint32 `gorm:"primary_key;autoIncrement" json:"-" yaml:"-"`
}

// GORM hooks for BaseUUID
func (baseUUID *BaseUUID) BeforeCreate(tx *gorm.DB) (err error) {
	if baseUUID.ID == "" {
		baseUUID.ID = uuid.New().String()
	}
	return nil
}
