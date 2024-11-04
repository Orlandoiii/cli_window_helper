package gui

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
)

const ADJUST_DNS = "AJUSTAR_DNS"
const ADJUST_DNS_APPSETTINGS = "AJUSTAR_DNS_APPSETTINGS"
const CHANGE_DNS_ENVIRONMENT = "CAMBIAR_ENTORNO_DNS"
const UPDATE_WINDOW_EXPORTER = "ACTUALIZAR_SERVICIO_WINDOWS_EXPORTER"

func DeploymentManagerForm() (string, error) {

	var opcionSeleccionada string

	confirm := false

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("CC27"),
			huh.NewSelect[string]().
				Title("Seleccione una Opcion").
				Options(
					huh.NewOption("Cambiar Ambiente DNS", CHANGE_DNS_ENVIRONMENT),
					huh.NewOption("Ajustar DNS", ADJUST_DNS),
					huh.NewOption("Actualizar Servicio Windows Exporter", UPDATE_WINDOW_EXPORTER),
					huh.NewOption("Ajustar DNS en AppSettings", ADJUST_DNS_APPSETTINGS),
				).
				Value(&opcionSeleccionada),
			huh.NewConfirm().TitleFunc(func() string {
				return fmt.Sprintf("¿Desea continuar con la opcion %s?", opcionSeleccionada)
			}, &opcionSeleccionada).Value(&confirm),
		),
	).WithTheme(huh.ThemeBase16())

	err := form.Run()
	if err != nil {
		return "", err
	}

	if !confirm {
		return "", errors.New("no se desea continuar")
	}

	return opcionSeleccionada, nil
}
