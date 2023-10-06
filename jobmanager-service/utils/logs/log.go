package logs

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {

	Logger = log.New(os.Stdout, "[OTPAAS-BACKEND] ", log.Ldate|log.Ltime)

}
