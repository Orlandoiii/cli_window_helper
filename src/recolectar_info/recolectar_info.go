package recolectar_info

import (
	"cli_window_helper/src/app_log"
	"cli_window_helper/src/config"
	"cli_window_helper/src/diagnostico"
	"cli_window_helper/src/file_management"
	"cli_window_helper/src/informe"
	"cli_window_helper/src/postgres_client"
	"cli_window_helper/src/prometheus_client"
	"fmt"
	"sync"
	"time"
)

const DESPLIEGUE = "despliegue"
const PRE_DESPLIEGUE = "pre_despliegue"

func generarJson(rutaPrincipal string, results map[string]string, extraName string) {
	for key, value := range results {
		name := fmt.Sprintf("%s_%s", extraName, key)
		if extraName == "" {
			name = key
		}
		app_log.Get().Printf("Nombre Generado: %s KEY:%s\n", name, key)
		app_log.Get().Println("Datos: ", value)

		err := informe.GenerarArchivoJson(rutaPrincipal, name, []byte(value))
		if err != nil {
			app_log.Get().Printf("Error Escribiendo Json =>%v\n", err)
			app_log.Get().Println("Key ", key)
			app_log.Get().Println("Error ", err)
			informe.GenerarArchivoTxt(rutaPrincipal, key, []byte(value))
		}
	}
}

func generarTxt(rutaPrincipal string, results map[string]string, extraName string) {
	for key, value := range results {

		name := fmt.Sprintf("%s_%s", extraName, key)
		if extraName == "" {
			name = key
		}
		app_log.Get().Printf("Nombre Generado: %s KEY:%s\n", name, key)
		app_log.Get().Println("Datos: ", value)

		err := informe.GenerarArchivoTxt(rutaPrincipal, name, []byte(value))
		if err != nil {
			app_log.Get().Printf("Error Escribiendo TXT =>%v\n", err)
			app_log.Get().Println("Key ", key)
			app_log.Get().Println("Error ", err)
		}
	}
}

func RecolectarInfo(momentoDespliegue string) {

	altoValor := true

	directorioPrincipal := fmt.Sprintf("./Evidencias_%s/Pre-Despliegue-%d",
		time.Now().Format("02-01-2006"), time.Now().UnixMilli())

	if momentoDespliegue == DESPLIEGUE {
		directorioPrincipal = fmt.Sprintf("./Evidencias_%s/Post-Despliegue-%d",
			time.Now().Format("02-01-2006"), time.Now().UnixMilli())
	}

	err := file_management.CrearDirectorio(directorioPrincipal)

	if err != nil {
		app_log.Get().Fatalf("No se Pudo Crear el Directorio Principal =>%v", err)
	}

	config.SetDirectorioPrincipal(directorioPrincipal)

	app_log.Get().Println("Ejecutando Diagnosticos")
	var grupoDeEspera sync.WaitGroup

	grupoDeEspera.Add(1)
	go func(waitgroup *sync.WaitGroup) {
		defer waitgroup.Done()
		diagnosticoResult := diagnostico.GetAll(altoValor)
		generarJson(config.GetPathDirectorioPrincipal(), diagnosticoResult, "")
	}(&grupoDeEspera)

	disco := fmt.Sprintf("%s:\\", config.GetDisco())

	app_log.Get().Println("Ejecutando Busqueda De Archivos Config AppSettings")

	fmt.Println(disco)
	resultAppSettings := file_management.GetConfigFiles(disco, file_management.AppFile)
	generarJson(config.GetPathDirectorioPrincipal(), resultAppSettings, "appsettings")

	app_log.Get().Println("Ejecutando Busqueda De Archivos Config NlogSettings")

	resultNlogSettings := file_management.GetConfigFiles(disco, file_management.NlogFile)
	generarTxt(config.GetPathDirectorioPrincipal(), resultNlogSettings, "nlog")

	app_log.Get().Println("Ejecutando Config Iniciales")

	configsIniciales, err := postgres_client.GetConfigInicial()
	if err != nil {
		app_log.Get().Printf("Error Ejecutando Diagnostico => %v\n", err)
		fmt.Printf("Error Ejecutando Diagnostico => %v\n", err)
	} else {
		generarJson(config.GetPathDirectorioPrincipal(), configsIniciales, "")
	}

	instancias_prometheus := prometheus_client.GetAll()
	generarJson(config.GetPathDirectorioPrincipal(), instancias_prometheus, "")

	err = file_management.GuardarHostFile(config.GetPathDirectorioPrincipal())
	if err != nil {
		app_log.Get().Printf("Error Respaldando hosts => %v\n", err)
	}

	err = file_management.GuardarInformeCLI(config.Get().RutaInformeCLI, config.GetPathDirectorioPrincipal())

	if err != nil {
		app_log.Get().Printf("Error Obteniendo Informe  => %v\n", err)
	}

	app_log.Get().Println("Ejecutando Busqueda de Logs")

	logsResutl := file_management.GetLastLogs(disco)
	generarTxt(config.GetPathDirectorioPrincipal(), logsResutl, "")

	grupoDeEspera.Wait()

	// app_log.Get().Println("generando ZIP ")

	// file_management.ZipSource(config.GetPathDirectorioPrincipal(), fmt.Sprintf("%s.zip", config.GetPathDirectorioPrincipal()))

	app_log.Get().Println("Finalizacion Del Recolector De Evidencias")

}
