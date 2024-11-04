package diagnostico

import (
	"cli_window_helper/src/app_log"
	"cli_window_helper/src/config"
	"cli_window_helper/src/http_client"
	"fmt"
	"io"

	"sync"
)

const (
	DiagnosticoSimf        = "diagnostico_simf"
	DiagnosticoLbtr        = "diagnostico_lbtr"
	DiagnosticoMonitorJson = "monitor_json"
	IncompletasSimfLabel   = "incompletas_simf"
	IncompletasLbtrLabel   = "incompletas_lbtr"
)

func getDiagnostico(url string) (string, error) {
	req := http_client.NewRequest([]byte(""), url, http_client.GET, http_client.JsonContent)
	app_log.Get().Println("Request De Diagnostico: ", req)
	response, err := http_client.SendRequest(req)
	if err != nil {
		return "", err
	}
	if response.StatusCode == 599 {
		app_log.Get().Println("Timeout en la peticion HTTP", req)
		return "", fmt.Errorf("TimeOut De La Peticion HTTP")
	}
	defer response.Body.Close()
	datos, err := io.ReadAll(response.Body)
	if err != nil {
		app_log.Get().Println("Error Ejecutando Request", req, err)
		return "", err
	}
	return string(datos), err
}

func Simf() (string, error) {
	urlSimf := fmt.Sprintf("%s:8082/simf/diagnostico/all", config.Get().RestApisIp)
	return getDiagnostico(urlSimf)
}
func Sglpar() (string, error) {
	urlLbtr := fmt.Sprintf("%s:8083/lbtr/diagnostico/all", config.Get().RestApisIp)
	return getDiagnostico(urlLbtr)
}
func MonitorJson() (string, error) {
	urlMonitor := fmt.Sprintf("%s:8082/simf/estatus/monitor/json", config.Get().RestApisIp)
	return getDiagnostico(urlMonitor)
}
func IncompletasSimf() (string, error) {
	urlSimf := fmt.Sprintf("%s:8082/simf/estatus/monitor/incompletas", config.Get().RestApisIp)
	return getDiagnostico(urlSimf)
}

func IncompletasLbtr() (string, error) {
	urlLbtr := fmt.Sprintf("%s:8083/simf/estatus/monitor/incompletas", config.Get().RestApisIp)
	return getDiagnostico(urlLbtr)
}

func wrapperAsync(result *string, waitGroup *sync.WaitGroup, lambda func() (string, error)) {
	waitGroup.Add(1)
	go func(r *string) {
		defer waitGroup.Done()
		var err error
		*result, err = lambda()
		if err != nil {
			app_log.Get().Printf("Error Ejecutando Request => %v", err)
		}
	}(result)
}

func GetAll(altoValor bool) map[string]string {
	resultados := make(map[string]string)

	var grupoDeEspera sync.WaitGroup

	simfResult := ""
	wrapperAsync(&simfResult, &grupoDeEspera, Simf)

	lbtrResult := ""
	if altoValor {
		wrapperAsync(&lbtrResult, &grupoDeEspera, Sglpar)
	}

	monitorResult := ""
	wrapperAsync(&monitorResult, &grupoDeEspera, MonitorJson)

	monitorIncompletasSimf := ""
	wrapperAsync(&monitorIncompletasSimf, &grupoDeEspera, IncompletasSimf)

	monitorIncompletasLbtr := ""
	wrapperAsync(&monitorIncompletasLbtr, &grupoDeEspera, IncompletasLbtr)

	grupoDeEspera.Wait()

	resultados[DiagnosticoSimf] = simfResult

	if altoValor {
		resultados[DiagnosticoLbtr] = lbtrResult
	}
	resultados[DiagnosticoMonitorJson] = monitorResult

	resultados[IncompletasSimfLabel] = monitorIncompletasSimf

	resultados[IncompletasLbtrLabel] = monitorIncompletasLbtr

	return resultados
}
