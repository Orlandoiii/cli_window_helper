package informe

import (
	"bytes"
	"cli_window_helper/src/app_log"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"
)

func GenerarArchivoJson(direccion string, nombre string, jsonByte []byte) error {
	direccion = path.Join(direccion, fmt.Sprintf("%s_%d.json", nombre, time.Now().UnixMilli()))
	app_log.Get().Println("Direccion Para Generar Archivo JSON: ", direccion)
	file, err := os.Create(direccion)
	if err != nil {
		app_log.Get().Println("Error Para Generar Archivo JSON: ", direccion, err)
		return err
	}
	defer file.Close()
	var prettyJson bytes.Buffer
	err = json.Indent(&prettyJson, jsonByte, "", "    ")
	if err != nil {
		app_log.Get().Println("Error Para Generar Archivo JSON: ", err)
		return err
	}
	_, err = file.Write(prettyJson.Bytes())
	if err != nil {
		app_log.Get().Println("Error Para Generar Archivo JSON: ", err)
		return err
	}
	return nil
}

func GenerarArchivoTxt(direccion string, nombre string, datos []byte) error {
	direccion = path.Join(direccion, fmt.Sprintf("%s_%d.txt", nombre, time.Now().UnixMilli()))
	app_log.Get().Println("Direccion Para Generar Archivo TXT: ", direccion)

	file, err := os.Create(direccion)
	if err != nil {
		app_log.Get().Println("Error Para Generar Archivo TXT: ", err)
		return err
	}
	defer file.Close()
	_, err = file.Write(datos)
	if err != nil {
		app_log.Get().Println("Error Para Generar Archivo TXT: ", err)
		return err
	}
	return nil
}
