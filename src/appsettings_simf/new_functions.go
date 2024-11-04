package appsettings_simf

import (
	"cli_window_helper/src/config"
	"cli_window_helper/src/manage_service"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/tidwall/gjson"
)

type AppSettingsJson struct {
	Path         string
	Content      string
	Product      string
	CurrentValue gjson.Result
}

type AppSettingsXml struct {
	Path         string
	Content      string
	Product      string
	CurrentValue *etree.Document
}

var VALID_SUBFOLDERS_MSWS_SIMF = []string{
	"Credito",
	"Debito",
	"Reverso",
	"Comunes_Simf",
}

var VALID_SUBFOLDERS_MSWS_LBTR = []string{
	"Inter",
	"Intra",
	"Efectivo",
	"Comunes_Sglpar",
}

const SIMF_BUSINESS = "SIMF"
const LBTR_BUSINESS = "SGLBTR"

func LoadAppSettingsJson(path string) (*AppSettingsJson, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("error leyendo el archivo: %s error: %v", path, err)
	}

	jsonContent := string(content)

	if !gjson.Valid(jsonContent) {
		return nil, fmt.Errorf("contenido json invalido en el path: %s", path)
	}

	jsonParsed := gjson.Parse(jsonContent)

	return &AppSettingsJson{Path: path, Content: jsonContent, CurrentValue: jsonParsed}, nil
}

func (a *AppSettingsJson) PrintInfo() {
	fmt.Println("Path:", a.Path)
	fmt.Println("CurrentValue:", a.CurrentValue.String())
}

func ChangeKafkaBootstrapWithEnviroment(bootstrap string, enviroment string, trimServer bool) (string, error) {
	servers := strings.Split(bootstrap, ",")
	var newServers []string

	for _, server := range servers {
		server = strings.TrimSpace(server)
		parts := strings.SplitN(server, ":", 2)
		hostParts := strings.Split(parts[0], ".")

		newHostParts := []string{}

		// Handle server prefix
		if !trimServer {
			newHostParts = append(newHostParts, "server")
		}

		// Process the middle parts (kafka1, kafka2, etc.)
		for _, part := range hostParts {
			if part != "server" && part != "desarrollo" &&
				part != "certificacion" && part != "simf" {
				newHostParts = append(newHostParts, part)
			}
		}

		// Add environment (unless it's production) and simf
		if enviroment != PRODUCCION {
			newHostParts = append(newHostParts, enviroment)
		}
		newHostParts = append(newHostParts, "simf")

		// Reconstruct the host with port
		host := strings.Join(newHostParts, ".")
		if len(parts) > 1 {
			newServers = append(newServers, host+":"+strings.TrimSpace(parts[1]))
		} else {
			newServers = append(newServers, host)
		}
	}

	return strings.Join(newServers, ","), nil
}
func ChangePostgresHostWithEnvironment(connectionString string, environment string, trimServer bool) (string, error) {
	// Split connection string into parts
	parts := strings.Split(connectionString, ";")
	var newParts []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "Host=") {
			// Extract host value
			host := strings.TrimPrefix(part, "Host=")
			hostParts := strings.Split(host, ".")

			newHostParts := []string{}

			// Handle server prefix
			if !trimServer {
				newHostParts = append(newHostParts, "server")
			}

			// Process the middle parts (postgres)
			for _, hostPart := range hostParts {
				if hostPart != "server" && hostPart != "desarrollo" &&
					hostPart != "certificacion" && hostPart != "simf" {
					newHostParts = append(newHostParts, hostPart)
				}
			}

			// Add environment (unless it's production) and simf
			if environment != PRODUCCION {
				newHostParts = append(newHostParts, environment)
			}
			newHostParts = append(newHostParts, "simf")

			// Reconstruct the host part
			newHost := strings.Join(newHostParts, ".")
			newParts = append(newParts, "Host="+newHost)
		} else {
			// Keep other parts unchanged
			newParts = append(newParts, part)
		}
	}

	return strings.Join(newParts, ";"), nil
}
func ChangeServerDiscoveryWithEnvironment(enviroment string, trimServer bool, addEndPoint bool) string {
	endPoint := ""
	result := ""
	if addEndPoint {
		endPoint = "/services"
	}

	if enviroment == PRODUCCION {
		if trimServer {
			result = "http://disovery.simf:5500"
		} else {
			result = "http://server.discovery.simf:5500"
		}
	} else {
		if trimServer {
			result = fmt.Sprintf("http://discovery.%s.simf:5500", enviroment)
		} else {
			result = fmt.Sprintf("http://server.discovery.%s.simf:5500", enviroment)
		}
	}
	return fmt.Sprintf("%s%s", result, endPoint)
}
func ChangePrometheusHostWithEnvironment(enviroment string, trimServer bool, addEndPoint bool) string {
	endPoint := ""
	result := ""
	if addEndPoint {
		endPoint = "/api/v1/"
	}

	if enviroment == PRODUCCION {
		if trimServer {
			result = "http://prometheus.simf:9090"
		} else {
			result = "http://server.prometheus.simf:9090"
		}
	} else {
		if trimServer {
			result = fmt.Sprintf("http://prometheus.%s.simf:9090", enviroment)
		} else {
			result = fmt.Sprintf("http://server.prometheus.%s.simf:9090", enviroment)
		}
	}
	return fmt.Sprintf("%s%s", result, endPoint)
}
func ChangeRestApiWithEnvironment(enviroment string, trimServer bool, port string) string {

	result := ""

	if enviroment == PRODUCCION {
		if trimServer {
			result = fmt.Sprintf("http://restapi.simf:%s", port)
		} else {
			result = fmt.Sprintf("http://server.restapi.simf:%s", port)
		}
	} else {
		if trimServer {
			result = fmt.Sprintf("http://restapi.%s.simf:%s", enviroment, port)
		} else {
			result = fmt.Sprintf("http://server.restapi.%s.simf:%s", enviroment, port)
		}
	}
	return result
}
func ChangeRestApiDashboardWithEnvironment(enviroment string, trimServer bool, port string, addEndPoint bool) string {
	endPoint := ""
	result := ""
	if addEndPoint {
		endPoint = "lbtr/api/v1/"
	}

	if enviroment == PRODUCCION {
		if trimServer {
			result = fmt.Sprintf("http://rest_api.simf:%s/", port)
		} else {
			result = fmt.Sprintf("http://server.rest_api.simf:%s/", port)
		}
	} else {
		if trimServer {
			result = fmt.Sprintf("http://rest_api.%s.simf:%s/", enviroment, port)
		} else {
			result = fmt.Sprintf("http://server.rest_api.%s.simf:%s/", enviroment, port)
		}
	}
	return fmt.Sprintf("%s%s", result, endPoint)
}

func (a *AppSettingsJson) ChangeBootstrapKafka(path string, enviroment string, trimServer bool) error {

	bootstrapServers := gjson.Get(a.Content, path).String()

	newBootstrapServers, err := ChangeKafkaBootstrapWithEnviroment(bootstrapServers, enviroment, trimServer)

	if err != nil {
		return fmt.Errorf("error changing kafka bootstrap: %v", err)
	}

	return a.ChangeField("KafkaSettings.BootstrapServers", newBootstrapServers)
}
func (a *AppSettingsJson) ChangeConnectionHost(path string, enviroment string, trimServer bool) error {
	postgresHost := gjson.Get(a.Content, path).String()

	newPostgresHost, err := ChangePostgresHostWithEnvironment(postgresHost, enviroment, trimServer)

	if err != nil {
		return fmt.Errorf("error changing postgres host: %v", err)
	}

	return a.ChangeField("DataBaseSettings.ConnectionString", newPostgresHost)
}

func LoadAppSettingsXml(path string) (*AppSettingsXml, error) {

	content, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("error leyendo el archivo: %s error: %v", path, err)
	}

	xmlContent := string(content)

	xmlParsed := etree.NewDocument()

	xmlParsed.ReadFromString(xmlContent)

	return &AppSettingsXml{Path: path, Content: xmlContent, CurrentValue: xmlParsed}, nil
}

func (a *AppSettingsXml) PrintInfo() {
	fmt.Println("Path:", a.Path)
	currentValue, err := a.CurrentValue.WriteToString()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("CurrentValue:", currentValue)
}

func (a *AppSettingsXml) ChangeAppSettingValue(key string, value string) error {
	return a.ChangeMultipleAppSettingValues(map[string]string{key: value})
}

func (a *AppSettingsXml) GetAppSettingValue(key string) (string, error) {
	return a.CurrentValue.FindElement("//appSettings/add[@key='" + key + "']").SelectAttr("value").Value, nil
}

// Multiple key-value changes using a dictionary
func (a *AppSettingsXml) ChangeMultipleAppSettingValues(changes map[string]string) error {
	// Find all add elements under appSettings
	appSettings := a.CurrentValue.FindElement("//appSettings")
	if appSettings == nil {
		return fmt.Errorf("appSettings section not found")
	}

	// Track which keys we've found and updated
	foundKeys := make(map[string]bool)

	// Update all matching keys
	for _, element := range appSettings.SelectElements("add") {
		key := element.SelectAttr("key").Value
		if newValue, exists := changes[key]; exists {
			element.SelectAttr("value").Value = newValue
			foundKeys[key] = true
		}
	}

	// Add any keys that weren't found
	for key, value := range changes {
		if !foundKeys[key] {
			// Create new element
			newElement := appSettings.CreateElement("add")
			newElement.CreateAttr("key", key)
			newElement.CreateAttr("value", value)
		}
	}

	// Save the changes back to the file
	err := a.CurrentValue.WriteToFile(a.Path)
	if err != nil {
		return fmt.Errorf("error writing XML file: %w", err)
	}

	return nil
}

func GetWindowsAppPaths(bussines string) (map[string]string, error) {
	disco := config.GetDisco()

	appPaths := make(map[string]string)

	mainPath := filepath.Join(fmt.Sprintf("%s:\\", disco), bussines)

	msPath := filepath.Join(mainPath, "MSWS")

	subfolders, err := os.ReadDir(msPath)

	if err != nil {
		return nil, fmt.Errorf("error leyendo el directorio: %s error: %v", msPath, err)
	}

	for _, subfolder := range subfolders {
		if !subfolder.IsDir() {
			continue
		}
		if bussines == SIMF_BUSINESS && !contains(VALID_SUBFOLDERS_MSWS_SIMF, subfolder.Name()) {
			continue
		}
		if bussines == LBTR_BUSINESS && !contains(VALID_SUBFOLDERS_MSWS_LBTR, subfolder.Name()) {
			continue
		}

		appPaths[subfolder.Name()] = filepath.Join(msPath, subfolder.Name(), "appsettings.json")
	}

	mainRestPath := "SIMF"
	if bussines == LBTR_BUSINESS {
		mainRestPath = "LBTR"
	}

	restPath := filepath.Join(mainPath, fmt.Sprintf("RestApi_%s", mainRestPath), "appsettings.json")

	appPaths[fmt.Sprintf("RestApi_%s", mainRestPath)] = restPath

	if bussines == SIMF_BUSINESS {
		appPaths["RestApi"] = filepath.Join(mainPath, "RestApi", "appsettings.json")
		appPaths["RestApiHistorico"] = filepath.Join(mainPath, "RestApi_Histo", "appsettings.json")

		// lastVersionDashboard, err := GetLatestVersionDashboardFolder(filepath.Join(mainPath, "Dashboard", "Application Files"))

		// if err == nil {
		// 	dashboardPath := filepath.Join(lastVersionDashboard, "dashboard.exe.config.deploy")
		// 	appPaths["Dashboard"] = dashboardPath
		// } else {
		// 	fmt.Printf("error obteniendo la carpeta de Dashboard: %v", err)
		// }

	}

	return appPaths, nil
}

func GetWindowExportersPaths() (map[string]string, error) {
	appPaths := make(map[string]string)

	postgresExporterPath, err := manage_service.GetPostgresqlExporterServiceExecutablePath()

	if err == nil {
		postgresExporterPath = filepath.Dir(postgresExporterPath)

		appPaths["PostgresqlExporter"] = filepath.Join(postgresExporterPath, "backup_zipper_config.json")
	} else {
		fmt.Printf("error obteniendo el path del servicio de PostgresqlExporter: %v", err)
	}

	windowExporterPath, err := manage_service.GetWindowsExporterServiceExecutablePath()

	if err == nil {
		windowExporterPath = filepath.Dir(windowExporterPath)

		appPaths["WindowExporter"] = filepath.Join(windowExporterPath, "appsettings.json")
	} else {
		fmt.Printf("error obteniendo el path del servicio de WindowExporter: %v", err)
	}

	return appPaths, nil
}

func ReadMSWSAppSettings(msPath string) ([]AppSettingsJson, error) {
	var appSettings []AppSettingsJson

	subfolders, err := os.ReadDir(msPath)

	if err != nil {
		return nil, fmt.Errorf("error reading MSWS directory: %v", err)
	}

	for _, subfolder := range subfolders {

		if subfolder.IsDir() {

			if !contains(VALID_SUBFOLDERS_MSWS_SIMF, subfolder.Name()) &&
				!contains(VALID_SUBFOLDERS_MSWS_LBTR, subfolder.Name()) {
				continue
			}

			product := subfolder.Name()

			appsettingsPath := filepath.Join(msPath, product, "appsettings.json")

			appSetting, err := LoadAppSettingsJson(appsettingsPath)

			if err != nil {
				fmt.Printf("error cargando el archivo appsettings.json en el path: %s error: %v", appsettingsPath, err)
				continue
			}

			appSetting.Product = product

			appSettings = append(appSettings, *appSetting)
		}
	}

	return appSettings, nil
}

func GetLatestVersionDashboardFolder(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %v", err)
	}

	var latestVersion string
	var latestParts []int

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Split name like "dashboard_1_6_8_0" into parts
		parts := strings.Split(entry.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		// Get version numbers and convert to integers
		versionParts := strings.Split(strings.Join(parts[1:], "_"), "_")
		var numbers []int
		for _, p := range versionParts {
			num, err := strconv.Atoi(p)
			if err != nil {
				continue
			}
			numbers = append(numbers, num)
		}

		// Compare versions
		if latestVersion == "" || compareVersions(numbers, latestParts) > 0 {
			latestVersion = entry.Name()
			latestParts = numbers
		}
	}

	if latestVersion == "" {
		return "", fmt.Errorf("no valid version folders found in %s", path)
	}

	return filepath.Join(path, latestVersion), nil
}

// compareVersions compares two version number arrays
// returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 []int) int {
	// If v2 is empty and v1 has values, v1 is greater
	if len(v2) == 0 && len(v1) > 0 {
		return 1
	}

	// Compare each number
	for i := 0; i < len(v1) && i < len(v2); i++ {
		if v1[i] > v2[i] {
			return 1
		}
		if v1[i] < v2[i] {
			return -1
		}
	}

	// If all numbers match, longer version is greater
	if len(v1) > len(v2) {
		return 1
	}
	if len(v1) < len(v2) {
		return -1
	}

	return 0
}
