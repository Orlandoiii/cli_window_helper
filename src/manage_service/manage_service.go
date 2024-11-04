package manage_service

import (
	"cli_window_helper/src/app_log"
	"fmt"
	"path/filepath"
	"time"

	"strings"

	"github.com/yusufpapurcu/wmi"
)

type services_bin string

const SIMF_SERVICES_BIN services_bin = "simf_ws.exe"
const SGLBTR_SERVICES_BIN services_bin = "sglbtr_ws.exe"
const WINDOWS_EXPORTER_SERVICES_BIN services_bin = "windows_exporter"
const POSTGRESQL_EXPORTER_SERVICES_BIN services_bin = "postgres_exporter"

type Win32_Service struct {
	DisplayName string
	Name        string
	ProcessId   uint32
	State       string
	Status      string
	StartMode   string
	PathName    string
}

func (service *Win32_Service) SycomService(executableName services_bin) bool {
	path := strings.TrimSpace(strings.ToLower(service.PathName))
	return strings.Contains(path, string(executableName))
}

func GetCurrentDisk() (string, error) {
	var servicios []Win32_Service
	q := wmi.CreateQuery(&servicios, "")
	err := wmi.Query(q, &servicios)
	if err != nil {
		return "", err
	}
	if len(servicios) == 0 {
		app_log.Get().Println("No Se Encontraron Servicios")

		return "", fmt.Errorf("no se ejecuto el query de los servicios ")
	}

	for _, service := range servicios {
		app_log.Get().Println(service)
		if service.SycomService(SIMF_SERVICES_BIN) ||
			service.SycomService(SGLBTR_SERVICES_BIN) {

			app_log.Get().Println(service)
			binPath := service.PathName
			app_log.Get().Println(binPath)
			disco := strings.Split(binPath, ":")[0]
			app_log.Get().Println(disco)
			return strings.TrimSpace(disco), nil
		}
	}
	return "", fmt.Errorf("no se conseguio ningun servicio SIMF instalado")
}
func GetWindowsExporterServiceExecutablePath() (string, error) {

	var servicios []Win32_Service
	q := wmi.CreateQuery(&servicios, "")
	err := wmi.Query(q, &servicios)
	if err != nil {
		return "", err
	}
	for _, service := range servicios {
		if service.SycomService(WINDOWS_EXPORTER_SERVICES_BIN) {
			return service.PathName, nil
		}
	}
	return "", fmt.Errorf("no se encontro el servicio windows exporter")
}

func GetPostgresqlExporterServiceExecutablePath() (string, error) {
	var servicios []Win32_Service
	q := wmi.CreateQuery(&servicios, "")
	err := wmi.Query(q, &servicios)
	if err != nil {
		return "", err
	}
	for _, service := range servicios {
		if service.SycomService(POSTGRESQL_EXPORTER_SERVICES_BIN) {
			return service.PathName, nil
		}
	}
	return "", fmt.Errorf("no se encontro el servicio postgresql exporter")
}

func GetDiscoWinwdowExporter() (string, error) {
	executablePath, err := GetWindowsExporterServiceExecutablePath()
	if err != nil {
		return "", err
	}
	return strings.Split(executablePath, ":")[0], nil
}

func GetDiscoPostgresqlExporter() (string, error) {
	executablePath, err := GetPostgresqlExporterServiceExecutablePath()
	if err != nil {
		return "", err
	}
	return strings.Split(executablePath, ":")[0], nil
}

func StopService(serviceName string) error {
	// Get the service object
	var services []Win32_Service
	query := fmt.Sprintf("WHERE Name = '%s'", serviceName)
	q := wmi.CreateQuery(&services, query)

	err := wmi.Query(q, &services)
	if err != nil {
		return fmt.Errorf("failed to query service: %v", err)
	}

	if len(services) == 0 {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	// Check service state before attempting to stop
	service := services[0]
	if service.State == "Stopped" {
		return fmt.Errorf("service '%s' is already stopped", serviceName)
	}

	// Get the specific service instance path
	servicePath := fmt.Sprintf("Win32_Service.Name='%s'", serviceName)

	// Execute the StopService method using WMI
	returnValue, err := wmi.CallMethod(nil, servicePath, "StopService", nil)
	if err != nil {
		return fmt.Errorf("failed to stop service: %v", err)
	}

	switch returnValue {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("service not found")
	case 2:
		return fmt.Errorf("service cannot be stopped (access denied or invalid service state)")
	case 3:
		return fmt.Errorf("service cannot be stopped due to dependent services running")
	case 4:
		return fmt.Errorf("service already stopped")
	default:
		return fmt.Errorf("failed to stop service, unknown return code: %d", returnValue)
	}
}
func StopServiceWithRetry(serviceName string, maxRetries int) error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Query current service state
		var services []Win32_Service
		query := fmt.Sprintf("WHERE Name = '%s'", serviceName)
		q := wmi.CreateQuery(&services, query)

		err := wmi.Query(q, &services)
		if err != nil {
			return fmt.Errorf("failed to query service: %v", err)
		}

		if len(services) == 0 {
			return fmt.Errorf("service '%s' not found", serviceName)
		}

		// Check if service is already stopped
		service := services[0]
		if service.State == "Stopped" {
			return nil // Service is already in desired state
		}

		// Try to stop the service
		servicePath := fmt.Sprintf("Win32_Service.Name='%s'", serviceName)
		returnValue, err := wmi.CallMethod(nil, servicePath, "StopService", nil)
		if err != nil {
			return fmt.Errorf("failed to stop service: %v", err)
		}

		switch returnValue {
		case 0:
			// Wait a bit for the service to actually stop
			time.Sleep(750 * time.Millisecond)
			continue // Check state again
		case 1:
			return fmt.Errorf("service not found")
		case 2:
			if attempt == maxRetries-1 {
				return fmt.Errorf("service cannot be stopped (access denied or invalid service state)")
			}
			time.Sleep(750 * time.Millisecond)
			continue
		case 3:
			if attempt == maxRetries-1 {
				return fmt.Errorf("service cannot be stopped due to dependent services running")
			}
			time.Sleep(750 * time.Millisecond)
			continue
		case 4:
			return nil // Service is already stopped
		default:
			if attempt == maxRetries-1 {
				return fmt.Errorf("failed to stop service, unknown return code: %d", returnValue)
			}
			time.Sleep(750 * time.Millisecond)
			continue
		}
	}

	// Final state check
	var services []Win32_Service
	query := fmt.Sprintf("WHERE Name = '%s'", serviceName)
	q := wmi.CreateQuery(&services, query)

	err := wmi.Query(q, &services)
	if err != nil {
		return fmt.Errorf("failed to query final service state: %v", err)
	}

	if len(services) == 0 {
		return fmt.Errorf("service '%s' not found in final check", serviceName)
	}

	if services[0].State != "Stopped" {
		return fmt.Errorf("service failed to stop after %d attempts. Current state: %s", maxRetries, services[0].State)
	}

	return nil
}
func StartService(serviceName string) error {
	// Get the service object
	var services []Win32_Service
	query := fmt.Sprintf("WHERE Name = '%s'", serviceName)
	q := wmi.CreateQuery(&services, query)

	err := wmi.Query(q, &services)
	if err != nil {
		return fmt.Errorf("failed to query service: %v", err)
	}

	if len(services) == 0 {
		return fmt.Errorf("service '%s' not found", serviceName)
	}

	// Check service state before attempting to start
	service := services[0]
	if service.State == "Running" {
		return fmt.Errorf("service '%s' is already running", serviceName)
	}

	// Get the specific service instance path
	servicePath := fmt.Sprintf("Win32_Service.Name='%s'", serviceName)

	// Execute the StartService method using WMI
	returnValue, err := wmi.CallMethod(nil, servicePath, "StartService", nil)
	if err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	switch returnValue {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("service not found")
	case 2:
		return fmt.Errorf("service cannot be started (access denied or invalid service state)")
	case 3:
		return fmt.Errorf("service cannot be started due to service dependencies")
	case 4:
		return fmt.Errorf("service already running")
	case 5:
		return fmt.Errorf("service marked for deletion")
	case 6:
		return fmt.Errorf("service has no start configuration")
	case 7:
		return fmt.Errorf("service is disabled")
	case 8:
		return fmt.Errorf("service logon failed")
	case 9:
		return fmt.Errorf("service request timeout")
	case 10:
		return fmt.Errorf("service failed to start due to unknown error")
	default:
		return fmt.Errorf("failed to start service, unknown return code: %d", returnValue)
	}
}
func StartServiceWithRetry(serviceName string, maxRetries int) error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Query current service state
		var services []Win32_Service
		query := fmt.Sprintf("WHERE Name = '%s'", serviceName)
		q := wmi.CreateQuery(&services, query)

		err := wmi.Query(q, &services)
		if err != nil {
			return fmt.Errorf("failed to query service: %v", err)
		}

		if len(services) == 0 {
			return fmt.Errorf("service '%s' not found", serviceName)
		}

		// Check if service is already running
		service := services[0]
		if service.State == "Running" {
			return nil // Service is already in desired state
		}

		// Try to start the service
		servicePath := fmt.Sprintf("Win32_Service.Name='%s'", serviceName)
		returnValue, err := wmi.CallMethod(nil, servicePath, "StartService", nil)
		if err != nil {
			return fmt.Errorf("failed to start service: %v", err)
		}

		switch returnValue {
		case 0:
			// Wait a bit for the service to actually start
			time.Sleep(750 * time.Millisecond)
			continue // Check state again
		case 1:
			return fmt.Errorf("service not found")
		case 2:
			if attempt == maxRetries-1 {
				return fmt.Errorf("service cannot be started (access denied or invalid service state)")
			}
			time.Sleep(750 * time.Millisecond)
			continue
		case 3:
			if attempt == maxRetries-1 {
				return fmt.Errorf("service cannot be started due to service dependencies")
			}
			time.Sleep(750 * time.Millisecond)
			continue
		case 4:
			return nil // Service is already running
		case 5:
			return fmt.Errorf("service marked for deletion")
		case 6:
			return fmt.Errorf("service has no start configuration")
		case 7:
			return fmt.Errorf("service is disabled")
		case 8:
			return fmt.Errorf("service logon failed")
		case 9:
			if attempt == maxRetries-1 {
				return fmt.Errorf("service request timeout")
			}
			time.Sleep(750 * time.Millisecond)
			continue
		case 10:
			if attempt == maxRetries-1 {
				return fmt.Errorf("service failed to start due to unknown error")
			}
			time.Sleep(750 * time.Millisecond)
			continue
		default:
			if attempt == maxRetries-1 {
				return fmt.Errorf("failed to start service, unknown return code: %d", returnValue)
			}
			time.Sleep(750 * time.Millisecond)
			continue
		}
	}

	// Final state check
	var services []Win32_Service
	query := fmt.Sprintf("WHERE Name = '%s'", serviceName)
	q := wmi.CreateQuery(&services, query)

	err := wmi.Query(q, &services)
	if err != nil {
		return fmt.Errorf("failed to query final service state: %v", err)
	}

	if len(services) == 0 {
		return fmt.Errorf("service '%s' not found in final check", serviceName)
	}

	if services[0].State != "Running" {
		return fmt.Errorf("service failed to start after %d attempts. Current state: %s", maxRetries, services[0].State)
	}

	return nil
}
func GetDirectoryService(serviceName string) (string, error) {
	// Query the specific service
	var services []Win32_Service
	query := fmt.Sprintf("WHERE Name = '%s'", serviceName)
	q := wmi.CreateQuery(&services, query)

	err := wmi.Query(q, &services)
	if err != nil {
		return "", fmt.Errorf("failed to query service: %v", err)
	}

	if len(services) == 0 {
		return "", fmt.Errorf("service '%s' not found", serviceName)
	}

	// Get the executable path
	execPath := services[0].PathName

	// Clean the path (remove quotes if present)
	execPath = strings.Trim(execPath, "\"")

	// Get the directory by removing the executable name
	directory := filepath.Dir(execPath)

	return directory, nil
}
