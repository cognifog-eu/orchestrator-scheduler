package logs

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {

	Logger = log.New(os.Stdout, "[JOB-MANAGER] ", log.Ldate|log.Ltime)

}
