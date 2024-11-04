package appsettings_simf

import (
	"cli_window_helper/src/config"
	"cli_window_helper/src/dns_simf"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type MainEndPointCoreSIMF struct {
	EndPointCredito      string
	EndPointDebito       string
	EndPointDebitoCheque string
	EndPointReverso      string
	EndPointSolicitudOtp string
}

func (m *MainEndPointCoreSIMF) PrintInfo() {

	endpoints := fmt.Sprintf(`EndPoints Core SIMF:
	  EndPointCredito:      %s
	  EndPointDebito:       %s
	  EndPointDebitoCheque: %s
	  EndPointReverso:      %s
	  EndPointSolicitudOtp: %s`,
		m.EndPointCredito,
		m.EndPointDebito,
		m.EndPointDebitoCheque,
		m.EndPointReverso,
		m.EndPointSolicitudOtp)

	fmt.Println(endpoints)
}

type MainEndPointCoreLBTR struct {
	EndPointInter    string
	EndPointIntra    string
	EndPointEfectivo string
}

func (m *MainEndPointCoreLBTR) PrintInfo() {
	endpoint := fmt.Sprintf(`EndPoints Core LBTR:
	  EndPointInter:    %s
	  EndPointIntra:    %s
	  EndPointEfectivo: %s`,
		m.EndPointInter, m.EndPointIntra, m.EndPointEfectivo)

	fmt.Println(endpoint)
}

func (a *AppSettingsJson) ChangeMultipleFields(changes map[string]string) error {
	modifiedJSON := a.Content

	for path, value := range changes {
		var err error
		modifiedJSON, err = sjson.Set(modifiedJSON, path, value)
		if err != nil {
			return fmt.Errorf("error modifying JSON for path %s: %w", path, err)
		}
	}

	// Parse the modified JSON
	var jsonData interface{}
	err := json.Unmarshal([]byte(modifiedJSON), &jsonData)
	if err != nil {
		return fmt.Errorf("error parsing modified JSON: %w", err)
	}

	// Pretty print the JSON
	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("error prettifying JSON: %w", err)
	}

	// Write the prettified JSON back to the file
	err = os.WriteFile(a.Path, prettyJSON, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	// Update the Content field with the new prettified JSON
	a.Content = string(prettyJSON)

	return nil
}

func (a *AppSettingsJson) ChangeMultipleFieldsInterface(changes map[string]interface{}) error {
	modifiedJSON := a.Content

	for path, value := range changes {
		var err error
		modifiedJSON, err = sjson.Set(modifiedJSON, path, value)
		if err != nil {
			return fmt.Errorf("error modifying JSON for path %s: %w", path, err)
		}
	}

	// Parse the modified JSON
	var jsonData interface{}
	err := json.Unmarshal([]byte(modifiedJSON), &jsonData)
	if err != nil {
		return fmt.Errorf("error parsing modified JSON: %w", err)
	}

	// Pretty print the JSON
	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("error prettifying JSON: %w", err)
	}

	// Write the prettified JSON back to the file
	err = os.WriteFile(a.Path, prettyJSON, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	// Update the Content field with the new prettified JSON
	a.Content = string(prettyJSON)

	return nil
}

func (a *AppSettingsJson) ChangeField(path string, value string) error {
	return a.ChangeMultipleFields(map[string]string{path: value})
}

func readMSWSAppSettings(msPath string) ([]AppSettingsJson, error) {
	var appSettings []AppSettingsJson

	subfolders, err := os.ReadDir(msPath)
	if err != nil {
		return nil, fmt.Errorf("error reading MSWS directory: %w", err)
	}

	for _, subfolder := range subfolders {
		if subfolder.IsDir() {

			product := subfolder.Name()

			appsettingsPath := filepath.Join(msPath, product, "appsettings.json")

			content, err := os.ReadFile(appsettingsPath)

			if err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("No Se Encontro El Archivo appsettings.json para el producto %s", product)
					panic(err) // Skip if appsettings.json doesn't exist
				}
				fmt.Printf("Error Al Leer El Archivo appsettings.json para el producto %s", product)
				panic(err)
			}

			appSettings = append(appSettings, AppSettingsJson{
				Path:    appsettingsPath,
				Content: string(content),
				Product: product,
			})
		}
	}

	return appSettings, nil
}

func loadAppSettings(disco string, tipo string) ([]AppSettingsJson, error) {

	mainFolder := tipo

	if tipo == "LBTR" {
		mainFolder = "SGLBTR"
	}

	folderPath := fmt.Sprintf("%s:\\%s", disco, mainFolder)
	msPath := filepath.Join(folderPath, "MSWS")
	restApiPath := filepath.Join(folderPath, "RestApi_"+tipo, "appsettings.json")

	var allAppSettings []AppSettingsJson

	msSettings, err := readMSWSAppSettings(msPath)
	if err != nil {
		return nil, fmt.Errorf("error reading %s MSWS settings: %w", tipo, err)
	}
	allAppSettings = append(allAppSettings, msSettings...)

	restApiContent, err := os.ReadFile(restApiPath)
	if err != nil {
		return nil, fmt.Errorf("error reading %s RestApi settings: %w", tipo, err)
	}
	allAppSettings = append(allAppSettings, AppSettingsJson{
		Path:    restApiPath,
		Content: string(restApiContent),
		Product: "RestApi_" + tipo,
	})

	return allAppSettings, nil
}

func LoadSimfAppSettings(disco string) ([]AppSettingsJson, error) {
	return loadAppSettings(disco, "SIMF")
}

func LoadLbtrAppSettings(disco string) ([]AppSettingsJson, error) {
	return loadAppSettings(disco, "LBTR")
}

func LoadAllAppSettings() ([]AppSettingsJson, error) {
	disco := config.GetDisco()

	simfSettings, err := LoadSimfAppSettings(disco)
	if err != nil {
		return nil, fmt.Errorf("error loading SIMF settings: %w", err)
	}

	lbtrSettings, err := LoadLbtrAppSettings(disco)
	if err != nil {
		return nil, fmt.Errorf("error loading LBTR settings: %w", err)
	}

	return append(simfSettings, lbtrSettings...), nil
}

func (a *AppSettingsJson) ChangeBootstrap() error {

	bootstrapServers := gjson.Get(a.Content, "KafkaSettings.BootstrapServers").String()

	newBootstrapServers, err := changeKafkaBootstrap(bootstrapServers)

	if err != nil {
		return fmt.Errorf("error changing kafka bootstrap: %v", err)
	}

	return a.ChangeField("KafkaSettings.BootstrapServers", newBootstrapServers)

}

func (a *AppSettingsJson) ChangePostgresHost() error {
	postgresHost := gjson.Get(a.Content, "DataBaseSettings.ConnectionString").String()

	newPostgresHost, err := changePostgresHost(postgresHost)

	if err != nil {
		return fmt.Errorf("error changing postgres host: %v", err)
	}

	return a.ChangeField("DataBaseSettings.ConnectionString", newPostgresHost)
}

func (a *AppSettingsJson) ChangeBootstrapEnviroment(newEnvironment string) error {

	bootstrapServers := gjson.Get(a.Content, "KafkaSettings.BootstrapServers").String()

	newBootstrapServers, err := changeKafkaBootstrapEnvironment(bootstrapServers, newEnvironment)

	if err != nil {
		return fmt.Errorf("error changing kafka bootstrap: %v", err)
	}

	return a.ChangeField("KafkaSettings.BootstrapServers", newBootstrapServers)

}

func (a *AppSettingsJson) ChangePostgresHostEnviroment(newEnvironment string) error {
	postgresHost := gjson.Get(a.Content, "DataBaseSettings.ConnectionString").String()

	newPostgresHost, err := changePostgresHostEnvironment(postgresHost, newEnvironment)

	if err != nil {
		return fmt.Errorf("error changing postgres host: %v", err)
	}

	return a.ChangeField("DataBaseSettings.ConnectionString", newPostgresHost)
}

func FixArgumentAppSettings(appSettings []AppSettingsJson, ambiente string) error {

	endpointServiceDiscovery := "http://server.discovery.desarrollo.simf:5500/services"

	endpointPrometheus := "http://server.prometheus.desarrollo.simf:9090/api/v1/"

	switch ambiente {

	case DESARROLLO:
		endpointServiceDiscovery = "http://server.discovery.desarrollo.simf:5500/services"
		endpointPrometheus = "http://server.prometheus.desarrollo.simf:9090/api/v1/"
	case CERTIFICACION:
		endpointServiceDiscovery = "http://server.discovery.certificacion.simf:5500/services"
		endpointPrometheus = "http://server.prometheus.certificacion.simf:9090/api/v1/"
	case PRODUCCION:
		endpointServiceDiscovery = "http://server.discovery.simf:5500/services"
		endpointPrometheus = "http://server.prometheus.simf:9090/api/v1/"
	}

	mapFields := map[string]string{
		"HTTPSettings.EndPointServiceDiscovery": endpointServiceDiscovery,
		"HTTPSettings.EndPointPrometheus":       endpointPrometheus,
	}

	for _, appSetting := range appSettings {
		err := appSetting.ChangeMultipleFields(mapFields)
		if err != nil {
			return fmt.Errorf("error fixing argument app settings for %s: %w", appSetting.Product, err)
		}
	}

	return nil
}

func ArgusWindowFixAppSettings(ambiente string) error {
	appSettings, err := LoadAllAppSettings()
	if err != nil {
		return err
	}
	err = FixArgumentAppSettings(appSettings, ambiente)
	if err != nil {
		fmt.Printf("error fixing argument app settings: %v\n", err)
		return err
	}
	return nil
}

func GetMainEndPointCoreSIMF(ambiente string) (*MainEndPointCoreSIMF, error) {
	appSettings, err := LoadSimfAppSettings(config.GetDisco())
	if err != nil {
		return nil, err
	}

	var mainEndPointCoreSIMF MainEndPointCoreSIMF
	for _, appSetting := range appSettings {

		switch appSetting.Product {
		case "Credito":
			mainEndPointCoreSIMF.EndPointCredito = gjson.Get(appSetting.Content, "HTTPSettings.EndPointCoreCredito").String()
		case "Debito":
			mainEndPointCoreSIMF.EndPointDebito = gjson.Get(appSetting.Content, "HTTPSettings.EndPointCoreDebito").String()
			mainEndPointCoreSIMF.EndPointDebitoCheque = gjson.Get(appSetting.Content, "HTTPSettings.EndPointCoreDebitoCheque").String()
		case "Reverso":
			mainEndPointCoreSIMF.EndPointReverso = gjson.Get(appSetting.Content, "HTTPSettings.EndPointCoreReverso").String()
		case "Comunes_Simf":
			mainEndPointCoreSIMF.EndPointSolicitudOtp = gjson.Get(appSetting.Content, "HTTPSettings.EndPointCoreSolicitudOtp").String()

		}

	}

	if mainEndPointCoreSIMF.EndPointDebitoCheque == "" {
		mainEndPointCoreSIMF.EndPointDebitoCheque = mainEndPointCoreSIMF.EndPointDebito
	}

	return &mainEndPointCoreSIMF, nil
}

func GetMainEndPointCoreLBTR(ambiente string) (*MainEndPointCoreLBTR, error) {
	appSettings, err := LoadLbtrAppSettings(config.GetDisco())
	if err != nil {
		return nil, err
	}

	var mainEndPointCoreLBTR MainEndPointCoreLBTR

	for _, appSetting := range appSettings {
		switch appSetting.Product {

		case "Inter":
			mainEndPointCoreLBTR.EndPointInter = gjson.Get(appSetting.Content, "HTTPSettings.EndPointCoreInter").String()
		case "Intra":
			mainEndPointCoreLBTR.EndPointIntra = gjson.Get(appSetting.Content, "HTTPSettings.EndPointCoreIntra").String()
		case "Efectivo":
			mainEndPointCoreLBTR.EndPointEfectivo = gjson.Get(appSetting.Content, "HTTPSettings.EndPointCoreEfectivo").String()
		}

	}

	return &mainEndPointCoreLBTR, nil
}

type EndPointData struct {
	Protocol            string
	Host                string
	NewHost             string
	Product             string
	Port                string
	Path                string
	Bussines            string
	Ip                  string
	AppSettingsNamePath string
}

func (e *EndPointData) String() string {
	return fmt.Sprintf("%s://%s:%s%s", e.Protocol, e.Host, e.Port, e.Path)
}

func (e *EndPointData) GetNewString() string {
	if e.NewHost == "" {
		panic("NewHost is empty")
	}
	return fmt.Sprintf("%s://%s:%s%s", e.Protocol, e.NewHost, e.Port, e.Path)
}

// ... existing code ...

func (e *EndPointData) PrintInfo() {
	fmt.Printf("EndPoint Details:\n")
	fmt.Printf("  Protocol: %s\n", e.Protocol)
	fmt.Printf("  Host:     %s\n", e.Host)
	fmt.Printf("  NewHost:  %s\n", e.NewHost)
	fmt.Printf("  Ip:       %s\n", e.Ip)
	fmt.Printf("  Port:     %s\n", e.Port)
	fmt.Printf("  Path:     %s\n", e.Path)
	fmt.Printf("  Product:  %s\n", e.Product)
	fmt.Printf("  AppSettingsNamePath:     %s\n", e.AppSettingsNamePath)
	fmt.Printf("  Bussines: %s\n", e.Bussines)
	fmt.Printf("  Full URL: %s\n", e.String())
}

func (e *EndPointData) HostIsIP() bool {
	return net.ParseIP(e.Host) != nil
}

func (e *EndPointData) LoadNewHost(ambiente string) {

	if ambiente == "produccion" {
		e.NewHost = fmt.Sprintf("core.%s.%s", e.Product, e.Bussines)
	} else {
		e.NewHost = fmt.Sprintf("core.%s.%s.%s", e.Product, ambiente, e.Bussines)
	}

}

type EndPointDataList struct {
	Bussines  string
	EndPoints []EndPointData
}

// ... rest of the file ...
func GetEndPointData(endpoint string) (*EndPointData, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error parsing endpoint: %w", err)
	}

	endPointData := &EndPointData{
		Protocol: url.Scheme,
		Host:     strings.Split(url.Host, ":")[0],
		Port:     url.Port(),
		Path:     url.Path,
	}

	if endPointData.HostIsIP() {
		endPointData.Ip = endPointData.Host
	}

	return endPointData, nil
}

func (e *MainEndPointCoreSIMF) GetEndPointDataListSimf() ([]EndPointData, error) {

	const bussines = "simf"

	endPointDataList := []EndPointData{}

	endPointData, err := GetEndPointData(e.EndPointCredito)
	if err != nil {
		return nil, err
	}
	endPointData.AppSettingsNamePath = "HTTPSettings.EndPointCoreCredito"
	endPointData.Product = "credito"
	endPointData.Bussines = bussines
	endPointDataList = append(endPointDataList, *endPointData)

	endPointData, err = GetEndPointData(e.EndPointDebito)
	if err != nil {
		return nil, err
	}
	endPointData.AppSettingsNamePath = "HTTPSettings.EndPointCoreDebito"
	endPointData.Product = "debito"
	endPointData.Bussines = bussines
	endPointDataList = append(endPointDataList, *endPointData)

	endPointData, err = GetEndPointData(e.EndPointDebitoCheque)
	if err != nil {
		return nil, err
	}

	endPointData.AppSettingsNamePath = "HTTPSettings.EndPointCoreDebitoCheque"
	endPointData.Product = "debito.cheque"
	endPointData.Bussines = bussines
	endPointDataList = append(endPointDataList, *endPointData)

	endPointData, err = GetEndPointData(e.EndPointSolicitudOtp)
	if err != nil {
		return nil, err
	}
	endPointData.AppSettingsNamePath = "HTTPSettings.EndPointCoreSolicitudOtp"
	endPointData.Product = "debito.otp"
	endPointData.Bussines = bussines
	endPointDataList = append(endPointDataList, *endPointData)

	endPointData, err = GetEndPointData(e.EndPointReverso)
	if err != nil {
		return nil, err
	}
	endPointData.AppSettingsNamePath = "HTTPSettings.EndPointCoreReverso"
	endPointData.Product = "reverso"
	endPointData.Bussines = bussines
	endPointDataList = append(endPointDataList, *endPointData)

	return endPointDataList, nil

}

func (e *MainEndPointCoreLBTR) GetEndPointDataListLBTR() ([]EndPointData, error) {

	const bussines = "lbtr"

	endPointDataList := []EndPointData{}

	endPointData, err := GetEndPointData(e.EndPointInter)
	if err != nil {
		return nil, err
	}
	endPointData.AppSettingsNamePath = "HTTPSettings.EndPointCoreInter"
	endPointData.Product = "inter"
	endPointData.Bussines = bussines

	endPointDataList = append(endPointDataList, *endPointData)

	endPointData, err = GetEndPointData(e.EndPointIntra)
	if err != nil {
		return nil, err
	}
	endPointData.AppSettingsNamePath = "HTTPSettings.EndPointCoreIntra"
	endPointData.Product = "intra"
	endPointData.Bussines = bussines
	endPointDataList = append(endPointDataList, *endPointData)

	endPointData, err = GetEndPointData(e.EndPointEfectivo)
	if err != nil {
		return nil, err
	}
	endPointData.AppSettingsNamePath = "HTTPSettings.EndPointCoreEfectivo"
	endPointData.Product = "efectivo"
	endPointData.Bussines = bussines
	endPointDataList = append(endPointDataList, *endPointData)

	return endPointDataList, nil
}

func GetDnsAppSettings(ambiente string) ([]AppSettingsJson, error) {
	return LoadSimfAppSettings(config.GetDisco())
}

func GetNewDnsCore(ambiente string) ([]dns_simf.DnsEntryHost, error) {
	mainEndPointCoreSIMF, err := GetMainEndPointCoreSIMF(ambiente)
	if err != nil {
		return nil, err
	}

	endPointSimfDataList, err := mainEndPointCoreSIMF.GetEndPointDataListSimf()
	if err != nil {
		return nil, err
	}

	mainEndPointCoreLBTR, err := GetMainEndPointCoreLBTR(ambiente)
	if err != nil {
		return nil, err
	}

	endPointLbtrDataList, err := mainEndPointCoreLBTR.GetEndPointDataListLBTR()
	if err != nil {
		return nil, err
	}

	endPointDataList := append(endPointSimfDataList, endPointLbtrDataList...)

	dnsEntries := []dns_simf.DnsEntryHost{}

	for _, endPointData := range endPointDataList {
		endPointData.LoadNewHost(ambiente)

		if endPointData.Ip == "" {
			dns, err := dns_simf.GetDnsEntryByHostname(endPointData.Host)
			if err != nil {
				return nil, err
			}
			endPointData.Ip = dns.IP
		}

		dnsEntry := dns_simf.DnsEntryHost{
			Hostname:        endPointData.NewHost,
			IP:              endPointData.Ip,
			Environment:     ambiente,
			TrailingComment: fmt.Sprintf("#Puerto %s", endPointData.Port),
			LineNumber:      0,
		}

		dnsEntry.LoadFullLine()

		dnsEntries = append(dnsEntries, dnsEntry)
	}

	return dnsEntries, nil
}

func ChangeAppSettings(ambiente string) error {

	hostPath := config.GetHostPath()

	discoveryDns := fmt.Sprintf("http://discovery.%s.simf:5500/services", ambiente)

	prometheusDns := fmt.Sprintf("http://prometheus.%s.simf:9090/api/v1/", ambiente)

	if ambiente == config.AMBIENTE_PRODUCCION {
		discoveryDns = "http://discovery.simf:5500/services"
		prometheusDns = "http://prometheus.simf:9090/api/v1/"
	}

	content, err := os.ReadFile(hostPath)
	if err != nil {
		return err
	}

	if !strings.Contains(string(content), "NUEVOS DNS SIMF CORE CC27") {
		return fmt.Errorf("no se han agregado los nuevos dns core CC27")
	}

	mainEndPointCoreSIMF, err := GetMainEndPointCoreSIMF(ambiente)
	if err != nil {
		return err
	}

	endPointSimfDataList, err := mainEndPointCoreSIMF.GetEndPointDataListSimf()
	if err != nil {
		return err
	}

	simfNewAppSettings := map[string]string{}

	for _, endPointData := range endPointSimfDataList {
		endPointData.LoadNewHost(ambiente)

		simfNewAppSettings[endPointData.AppSettingsNamePath] =
			endPointData.GetNewString()
	}

	simfNewAppSettings["HTTPSettings.EndPointPrometheus"] = prometheusDns
	simfNewAppSettings["HTTPSettings.EndPointServiceDiscovery"] = discoveryDns

	mainEndPointCoreLBTR, err := GetMainEndPointCoreLBTR(ambiente)
	if err != nil {
		return err
	}

	endPointLbtrDataList, err := mainEndPointCoreLBTR.GetEndPointDataListLBTR()
	if err != nil {
		return err
	}

	lbtrNewAppSettings := map[string]string{}

	for _, endPointData := range endPointLbtrDataList {
		endPointData.LoadNewHost(ambiente)
		lbtrNewAppSettings[endPointData.AppSettingsNamePath] = endPointData.GetNewString()
	}

	lbtrNewAppSettings["HTTPSettings.EndPointPrometheus"] = prometheusDns
	lbtrNewAppSettings["HTTPSettings.EndPointServiceDiscovery"] = discoveryDns

	allAppSettings, err := LoadAllAppSettings()
	if err != nil {
		return err
	}

	for _, appSetting := range allAppSettings {

		appSetting.ChangeBootstrap()
		appSetting.ChangePostgresHost()

		if strings.Contains(appSetting.Path, "SIMF") {
			err = appSetting.ChangeMultipleFields(simfNewAppSettings)
			if err != nil {
				return err
			}
		} else if strings.Contains(appSetting.Path, "LBTR") {
			err = appSetting.ChangeMultipleFields(lbtrNewAppSettings)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ChangeAppSettingsEnviroment(ambiente string) error {

	discoveryDns := fmt.Sprintf("http://server.discovery.%s.simf:5500/services", ambiente)

	prometheusDns := fmt.Sprintf("http://server.prometheus.%s.simf:9090/api/v1/", ambiente)

	if ambiente == config.AMBIENTE_PRODUCCION {
		discoveryDns = "http://server.discovery.simf:5500/services"
		prometheusDns = "http://server.prometheus.simf:9090/api/v1/"
	}

	mainEndPointCoreSIMF, err := GetMainEndPointCoreSIMF(ambiente)
	if err != nil {
		return err
	}

	endPointSimfDataList, err := mainEndPointCoreSIMF.GetEndPointDataListSimf()
	if err != nil {
		return err
	}

	simfNewAppSettings := map[string]string{}

	for _, endPointData := range endPointSimfDataList {

		simfNewAppSettings[endPointData.AppSettingsNamePath] =
			endPointData.String()
	}

	simfNewAppSettings["HTTPSettings.EndPointPrometheus"] = prometheusDns
	simfNewAppSettings["HTTPSettings.EndPointServiceDiscovery"] = discoveryDns

	mainEndPointCoreLBTR, err := GetMainEndPointCoreLBTR(ambiente)
	if err != nil {
		return err
	}

	endPointLbtrDataList, err := mainEndPointCoreLBTR.GetEndPointDataListLBTR()
	if err != nil {
		return err
	}

	lbtrNewAppSettings := map[string]string{}

	for _, endPointData := range endPointLbtrDataList {
		//endPointData.LoadNewHost(ambiente)
		lbtrNewAppSettings[endPointData.AppSettingsNamePath] = endPointData.String()
	}

	lbtrNewAppSettings["HTTPSettings.EndPointPrometheus"] = prometheusDns
	lbtrNewAppSettings["HTTPSettings.EndPointServiceDiscovery"] = discoveryDns

	allAppSettings, err := LoadAllAppSettings()
	if err != nil {
		return err
	}

	for _, appSetting := range allAppSettings {

		appSetting.ChangeBootstrapEnviroment(ambiente)
		appSetting.ChangePostgresHostEnviroment(ambiente)

		if strings.Contains(appSetting.Path, "SIMF") {
			err = appSetting.ChangeMultipleFields(simfNewAppSettings)
			if err != nil {
				return err
			}
		} else if strings.Contains(appSetting.Path, "LBTR") {
			err = appSetting.ChangeMultipleFields(lbtrNewAppSettings)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func GetLogZipperSettingsWindowExporter(path string) (*AppSettingsJson, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	appSetting := &AppSettingsJson{
		Path:    path,
		Content: string(content),
		Product: "window_exporter",
	}

	return appSetting, nil
}

func UpdateLogZipperSettingsWindowExporter(path string, simfDisk string) error {
	logZipperConfig, err := GetLogZipperSettingsWindowExporter(path)
	if err != nil {
		return err
	}

	logZipperPath := filepath.Join(fmt.Sprintf("%s:", simfDisk), "SYCOM", "LOGS_SIMF_EXPORTER", "log.logs")

	changeValues := map[string]interface{}{
		"DeleteWrongFileOrDirectory":  true,
		"DisableLogCompressor":        false,
		"LogsPath":                    logZipperPath,
		"RotationIntervalInSeconds":   86400,
		"EnableLogs":                  false,
		"MonitoringIntervalInSeconds": 300,
	}

	err = logZipperConfig.ChangeMultipleFieldsInterface(changeValues)
	if err != nil {
		return err
	}

	return nil
}
