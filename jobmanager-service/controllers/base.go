package controllers

import (
	"context"
	"flag"
	"fmt"
	"icos/server/jobmanager-service/models"
	"icos/server/jobmanager-service/utils/logs"
	"net/http"
	"os"
	"os/signal"
	"time"

	"log"

	go_driver "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
			}), &gorm.Config{})

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

	server.DB.Debug().AutoMigrate(&models.Job{}, &models.Target{}) //database migration

	server.Router = mux.NewRouter()

	server.initializeRoutes()
}

func (server *Server) Run(addr string) {
	logs.Logger.Println("Listening to port " + addr + " ...")
	handler := cors.AllowAll().Handler(server.Router)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	go func() {
		// init server
		if err := http.ListenAndServe(addr, handler); err != nil {
			if err != http.ErrServerClosed {
				logs.Logger.Fatal(err)
			}
		}
	}()

	<-stop

	// after stopping server
	logs.Logger.Println("Closing connections ...")

	var shutdownTimeout = flag.Duration("shutdown-timeout", 10*time.Second, "shutdown timeout (5s,5m,5h) before connections are cancelled")
	_, cancel := context.WithTimeout(context.Background(), *shutdownTimeout)
	defer cancel()
}
