package prometheus_client

import (
	"cli_window_helper/src/app_log"
	"cli_window_helper/src/config"
	"cli_window_helper/src/http_client"
	"fmt"
	"io"

	"sync"
)

type TipoDeArchivo string

const (
	UpInstances   = "prometheus_up_instances"
	DownInstances = "prometheus_down_instances"
)

const (
	USERNAME = "admin"
	PASSWORD = "$Yc0m-2023"
)

func ExecuteQuery(query string) (string, error) {
	urlProm := fmt.Sprintf("%squery?%s", config.Get().DireccionPrometheus, query)
	req := http_client.NewRequest([]byte(""), urlProm, http_client.GET, http_client.JsonContent)
	response, err := http_client.SendRequestWithBasicAuth(req, USERNAME, PASSWORD)
	if err != nil {
		return "", err
	}
	if response.StatusCode == 599 {
		return "", fmt.Errorf("TimeOut De La Peticion HTTP")
	}
	defer response.Body.Close()
	datos, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(datos), err
}

func GetUpInstances() (string, error) {
	return ExecuteQuery("query=up")
}
func GetDownInstances() (string, error) {
	return ExecuteQuery("query=down")
}

func wrapperAsync(result *string, waitGroup *sync.WaitGroup, lambda func() (string, error)) {
	waitGroup.Add(1)
	go func(r *string) {
		defer waitGroup.Done()
		var err error
		*result, err = lambda()
		if err != nil {
			app_log.Get().Printf("Error Ejecutando Obtencion de Instancias Prometheus => %v", err)
		}
	}(result)
}

func GetAll() map[string]string {
	resultados := make(map[string]string)

	var grupoDeEspera sync.WaitGroup

	upResult := ""
	wrapperAsync(&upResult, &grupoDeEspera, GetUpInstances)

	downResult := ""
	wrapperAsync(&downResult, &grupoDeEspera, GetDownInstances)

	grupoDeEspera.Wait()

	resultados[UpInstances] = upResult
	resultados[DownInstances] = downResult

	return resultados
}
