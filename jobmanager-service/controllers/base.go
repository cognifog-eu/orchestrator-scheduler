package controllers

import (
	"fmt"
	"icos/server/jobmanager-service/models"
	"net/http"

	"log"

	go_driver "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Server struct {
	DB     *gorm.DB
	Router *mux.Router
}

func (server *Server) Init() {
	server.Router = mux.NewRouter()
	server.initializeRoutes()
}

func (server *Server) Initialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName string) {

	var err error

	if Dbdriver == "mysql" {
		DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)
		config := go_driver.Config{
			AllowNativePasswords: true, // deprecate in the future
		}
		server.DB, err = gorm.Open(mysql.New(
			mysql.Config{
				DSN:       DBURL,
				DSNConfig: &config,
			}), &gorm.Config{})

		if err != nil {
			fmt.Printf("Cannot connect to %s database", Dbdriver)
			log.Fatal("This is the error:", err)
		} else {
			fmt.Printf("We are connected to the %s database", Dbdriver)
		}
		// first time schema creation
		server.DB.Exec("CREATE DATABASE IF NOT EXISTS " + DbName + ";")
	}

	server.DB.Debug().AutoMigrate(&models.Job{}) //database migration

	server.Router = mux.NewRouter()

	server.initializeRoutes()
}

func (server *Server) Run(addr string) {
	fmt.Printf("Listening to port %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
