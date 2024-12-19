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
package controllers

import (
	"context"
	"etsn/server/jobmanager-service/models"
	"etsn/server/jobmanager-service/service"
	"etsn/server/jobmanager-service/utils/logs"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"etsn/server/jobmanager-service/repository"
	"log"

	go_driver "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Server struct {
	DB               *gorm.DB
	Router           *mux.Router
	JobService       service.JobService
	JobGroupService  service.JobGroupService
	PolicyService    service.PolicyService
	ResourceService  service.ResourceService
	AllocatorService service.AllocatorService
	ManifestService  service.ManifestService
}

func (server *Server) Initialize(dbdriver, dbUser, dbPassword, dbPort, dbHost, dbName string) {

	var err error

	if dbdriver == "mysql" {
		DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
		config := go_driver.Config{
			AllowNativePasswords: true, // deprecate in the future
		}
		server.DB, err = gorm.Open(mysql.New(
			mysql.Config{
				DSN:       DBURL,
				DSNConfig: &config,
			}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

		if err != nil {
			fmt.Printf("Cannot connect to %s database", dbdriver)
			log.Fatal("This is the error:", err)
			DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort)
			config := go_driver.Config{
				AllowNativePasswords: true, // deprecate in the future
			}
			server.DB, err = gorm.Open(mysql.New(
				mysql.Config{
					DSN:       DBURL,
					DSNConfig: &config,
				}), &gorm.Config{})
			server.DB.Exec("USE " + dbName)
		} else {
			fmt.Printf("We are connected to the %s database", dbdriver)
		}
		// first time schema creation
		server.DB.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + ";")
		server.DB.Exec("USE " + dbName)
	}

	server.DB.Debug().
		AutoMigrate(
			&models.JobGroup{},
			&models.Job{},
			&models.Instruction{},
			&models.Content{},
			&models.Policy{},
			&models.Requirement{},
			//	&models.ManifestRef{},
			&models.Target{},
			&models.Resource{},
			&models.Condition{},
			&models.Remediation{},
			&models.RemediationTarget{},
		)

	server.Router = mux.NewRouter()

	// Initialize repositories
	jobRepo := repository.NewJobRepository(server.DB)
	jobGroupRepo := repository.NewJobGroupRepository(server.DB)
	policyRepo := repository.NewPolicyRepository(server.DB)
	resourceRepo := repository.NewResourceRepository(server.DB)
	httpClient := &http.Client{}

	// Initialize services
	server.AllocatorService = service.NewAllocatorService()
	server.ManifestService = service.NewManifestService()
	server.ResourceService = service.NewResourceService(resourceRepo, jobRepo)
	server.JobGroupService = service.NewJobGroupService(jobGroupRepo, server.AllocatorService, server.ManifestService)
	server.JobService = service.NewJobService(jobRepo, server.ResourceService, server.AllocatorService, server.JobGroupService)
	// TODO: we should reference a single httpclient for all services
	server.PolicyService = service.NewPolicyService(policyRepo, server.JobService, httpClient)

	// swagger
	server.Router.PathPrefix("/jobmanager/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

	// enable JWT middleware
	enableJWT := os.Getenv("ENVIRONMENT") != "development"

	server.initializeRoutes(enableJWT)
}

func (server *Server) Run(addr string) {
	logs.Logger.Println("Listening on port " + addr + " ...")
	handler := cors.AllowAll().Handler(server.Router)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,

		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	defer signal.Stop(stop)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Logger.Fatalf("Could not listen on %s: %v\n", addr, err)
		}
	}()

	logs.Logger.Println("Server is ready to handle requests")

	<-stop

	logs.Logger.Println("Shutdown signal received")
	logs.Logger.Println("Shutting down server gracefully...")

	shutdownTimeout := flag.Duration("shutdown-timeout", 10*time.Second, "shutdown timeout (5s,5m,5h) before connections are cancelled")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logs.Logger.Fatalf("Server Shutdown Failed:%+v", err)
	}

	logs.Logger.Println("Server gracefully stopped")
}
