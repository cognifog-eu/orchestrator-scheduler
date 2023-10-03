package server

import (
	"fmt"
	"icos/server/jobmanager-service/controllers"
	"os"
)

var server = controllers.Server{}

func Init() {
	// loads values from .env into the system
	// if err := godotenv.Load(); err != nil {
	// 	log.Print("sad .env file found")
	// }

}

func Run() {
	server.Initialize(os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	addr := fmt.Sprintf(":%s", os.Getenv(("SERVER_PORT")))
	server.Run(addr)

}
