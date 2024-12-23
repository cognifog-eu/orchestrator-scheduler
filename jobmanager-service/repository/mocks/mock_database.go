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
package repository

import (
	"etsn/server/jobmanager-service/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type RepoInitializer func(db *gorm.DB) interface{}

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		assert.FailNow(nil, "Error connecting to the database")
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&models.JobGroup{},
		&models.Job{},
		&models.Instruction{},
		&models.Content{},
		&models.Policy{},
		&models.Requirement{},
		&models.Target{},
		&models.Resource{},
		&models.Condition{},
		&models.Remediation{},
		&models.RemediationTarget{},
	)

	if err != nil {
		assert.FailNow(t, "Error migrating the database schema")
	}

	// Teardown function to clean up the database
	teardown := func() {
		sqlDB, err := db.DB()
		if err != nil {
			assert.FailNow(t, "Error getting db instance")
		}
		sqlDB.Close()
	}

	return db, teardown
}

func SetupTest(t *testing.T, initializer RepoInitializer) interface{} {
	db, teardown := setupTestDB(t)
	repo := initializer(db)

	// Teardown to cleanup database after each test
	t.Cleanup(teardown)

	return repo
}
