package update_services

import (
	"cli_window_helper/src/appsettings_simf"
	"cli_window_helper/src/config"
	"cli_window_helper/src/file_management"
	"cli_window_helper/src/manage_service"
	"path/filepath"
)

func UpdateWindowExporter() error {
	err := manage_service.StopServiceWithRetry("windows_exporter", 8)

	if err != nil {
		return err
	}

	directory, err := manage_service.GetDirectoryService("windows_exporter")

	if err != nil {
		return err
	}

	err = file_management.CopyDirectory("../Exports/windows_exporter", directory)

	if err != nil {
		return err
	}

	err = appsettings_simf.UpdateLogZipperSettingsWindowExporter(
		filepath.Join(directory, "logs_zipper_config.json"), config.GetDisco())

	if err != nil {
		return err
	}

	return manage_service.StartServiceWithRetry("windows_exporter", 8)

}
