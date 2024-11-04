package app_log

import (
	"fmt"
	"log"
	"os"
	"time"
)

var logger *log.Logger
var nombreDelArchivoLog string

func init() {
	ruta := fmt.Sprintf("./Evidencias_%s", time.Now().Format("02-01-2006"))
	_, err := os.Stat(ruta)

	if err != nil {
		if !os.IsExist(err) {
			errDir := os.MkdirAll(ruta, 0777)
			if errDir != nil {
				log.Panic(errDir)
			}
		} else {
			log.Panic(err)
		}
	}

	nombreDelArchivoLog = fmt.Sprintf("%s/log_recuperar_archivos%d", ruta, time.Now().UnixMilli())
	archivo, err := os.Create(nombreDelArchivoLog)
	if err != nil {
		log.Panic(err)
	}
	logger = log.Default()
	logger.SetOutput(archivo)
}

func GetNombreDelArchivoLog() string {
	return nombreDelArchivoLog
}
func Get() *log.Logger {
	return logger
}
