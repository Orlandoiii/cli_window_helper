package main

import (
	"cli_window_helper/src/appsettings_simf"
	"cli_window_helper/src/gui"
	"fmt"
)

func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
			gui.WaitForEnter()

		}
	}()
	// gui.MainForm()
	// gui.WaitForEnter()

	// appPathsSIMF, err := appsettings_simf.GetWindowsAppPaths(appsettings_simf.SIMF_BUSINESS)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	// for key, value := range appPathsSIMF {
	// 	fmt.Println(key, value)

	// 	// appSettings, err := appsettings_simf.LoadAppSettingsJson(value)
	// 	// if err != nil {
	// 	// 	fmt.Println("Error:", err)
	// 	// 	continue
	// 	// }
	// 	// appSettings.PrintInfo()

	// }

	// appPathsLBTR, err := appsettings_simf.GetWindowsAppPaths(appsettings_simf.LBTR_BUSINESS)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	// for key, value := range appPathsLBTR {
	// 	fmt.Println(key, value)
	// 	// appSettings, err := appsettings_simf.LoadAppSettingsJson(value)
	// 	// if err != nil {
	// 	// 	fmt.Println("Error:", err)
	// 	// 	continue
	// 	// }
	// 	// appSettings.PrintInfo()
	// }

	// windowExportersPaths, err := appsettings_simf.GetWindowExportersPaths()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	// for key, value := range windowExportersPaths {
	// 	fmt.Println(key, value)
	// }

	kafkaStr := "Host=server.postgres.desarrollo.simf;Port=5432;Username=simf_admin_user;Database=simf;Minimum Pool Size=4;Maximum Pool Size=16;Timeout=10;CommandTimeout=15"

	newKafkaStr, err := appsettings_simf.ChangePostgresHostWithEnvironment(kafkaStr, appsettings_simf.CERTIFICACION, false)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println(newKafkaStr)

}
