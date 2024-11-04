package config

import (
	"cli_window_helper/src/app_log"
	"cli_window_helper/src/manage_service"
	"encoding/json"
	"os"
	"runtime"
)

var configuracion *Settings
var discoDeLosEjecutables string
var directorioPrincipal string
var directorioCargado bool = false

var ambiente string = ""

const AMBIENTE_DESARROLLO = "desarrollo"
const AMBIENTE_CERTIFICACION = "certificacion"
const AMBIENTE_PRODUCCION = "produccion"

type Settings struct {
	DireccionPostgres   string `json:"Postgres"`
	DireccionPrometheus string `json:"Prometheus"`
	RestApisIp          string `json:"RestApisIp"`
	RutaInformeCLI      string `json:"RutaInformeCLI"`
}

func newSettings() *Settings {
	return &Settings{}
}

func init() {
	app_log.Get().Println("Iniciando Carga Del Settings")
	settings := newSettings()

	configFile, err := os.ReadFile("./settings.json")
	app_log.Get().Println("Config File: ", string(configFile))

	if err != nil {
		app_log.Get().Fatal(err)
	}
	err = json.Unmarshal(configFile, settings)
	if err != nil {
		app_log.Get().Fatal(err)
	}
	configuracion = settings
	discoDeLosEjecutables, err = manage_service.GetCurrentDisk()
	if err != nil {
		app_log.Get().Fatal(err)
	}
	app_log.Get().Println("Disco De Los Ejecutables: ", discoDeLosEjecutables)

	app_log.Get().Println("FIN Carga Del Settings")
}

func GetDisco() string {
	return discoDeLosEjecutables
}

func Get() *Settings {
	return configuracion
}

func SetDirectorioPrincipal(direccion string) {
	if !directorioCargado {
		directorioPrincipal = direccion
		directorioCargado = true
	}
}

func GetPathDirectorioPrincipal() string {
	if !directorioCargado {
		app_log.Get().Fatal("se quiere obtener el directorio principal y aun no ha sido cargado")
	}
	return directorioPrincipal
}

func GetHostPath() string {
	switch runtime.GOOS {
	case "linux", "darwin":
		return "/etc/hosts"
	case "windows":
		return "C:\\Windows\\System32\\drivers\\etc\\hosts"
	default:
		panic("Unsupported OS")
	}
}

func SetAmbiente(amb string) {
	if amb != AMBIENTE_DESARROLLO &&
		amb != AMBIENTE_CERTIFICACION &&
		amb != AMBIENTE_PRODUCCION {
		panic("Ambiente Invalido")
	}
	ambiente = amb
}

func GetAmbiente() string {
	if ambiente == "" {
		panic("Ambiente No Cargado")
	}
	return ambiente
}
