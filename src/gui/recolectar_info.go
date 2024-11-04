package gui

import (
	"cli_window_helper/src/app_log"
	"cli_window_helper/src/recolectar_info"
	"fmt"

	"github.com/charmbracelet/huh"
)

func RecolectarInfoForm() {

	var momentoDespliegue string

	confirmacion := false

	recolectarInfoForm := huh.NewForm(

		huh.NewGroup(
			huh.NewNote().Title("Recolectar Info"),

			huh.NewSelect[string]().
				Title("Selecciona el Momento Despliegue").
				Options(
					huh.NewOption("Despliegue", recolectar_info.DESPLIEGUE),
					huh.NewOption("Pre-Despliegue", recolectar_info.PRE_DESPLIEGUE),
					huh.NewOption("Cancelar", "CANCELAR"),
				).
				Value(&momentoDespliegue),
			huh.NewConfirm().
				TitleFunc(func() string {
					return fmt.Sprintf("¿Estas Seguro de Recolectar Info? => %s", momentoDespliegue)
				}, &momentoDespliegue).Affirmative("Si").Negative("No").Value(&confirmacion),
		),
	).WithTheme(GlobalTheme)

	err := recolectarInfoForm.Run()

	if err != nil {
		app_log.Get().Printf("Error al Recolectar Info => %v", err)
		panic(err)
	}

	if confirmacion && momentoDespliegue != "CANCELAR" {
		recolectar_info.RecolectarInfo(momentoDespliegue)
	}

}
