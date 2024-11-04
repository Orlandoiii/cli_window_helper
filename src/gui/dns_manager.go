package gui

import (
	"cli_window_helper/src/app_log"
	"cli_window_helper/src/appsettings_simf"
	"cli_window_helper/src/dns_simf"
	"fmt"

	"github.com/charmbracelet/huh"
)

const ADD_DNS = "ADD_DNS"
const EDIT_DNS = "EDIT_DNS"
const DELETE_DNS = "DELETE_DNS"

func DnsManagerForm() {

	var tipoAccion string

	confirmacion := false

	dnsManagerForm := huh.NewForm(

		huh.NewGroup(
			huh.NewNote().Title("Gestionar DNS"),

			huh.NewSelect[string]().
				Title("Seleccione Tipo de Accion").
				Options(
					huh.NewOption("Agregar DNS", ADD_DNS),
					huh.NewOption("Editar DNS", EDIT_DNS),
					huh.NewOption("Eliminar DNS", DELETE_DNS),
				).
				Value(&tipoAccion),
			huh.NewConfirm().
				TitleFunc(func() string {
					return fmt.Sprintf("¿Estas Seguro de %s? => %s", tipoAccion, tipoAccion)
				}, &tipoAccion).Affirmative("Si").Negative("No").Value(&confirmacion),
		),
	).WithTheme(GlobalTheme)

	err := dnsManagerForm.Run()

	if err != nil {
		app_log.Get().Printf("Error al Gestionar DNS => %v", err)
		panic(err)
	}

	if confirmacion {
		switch tipoAccion {
		case ADD_DNS:
			dns_test()
		case EDIT_DNS:
			//EditDnsForm()
		case DELETE_DNS:
			//DeleteDnsForm()
		}
	}
}

func AddNewDnsCoreEntries(ambiente string) error {
	dnsEntries, err := appsettings_simf.GetNewDnsCore(ambiente)
	if err != nil {
		return err
	}

	err = dns_simf.AddNewDnsCoreEntries(dnsEntries)
	if err != nil {
		return err
	}

	return nil
}

func ManageDeploymentDNS(ambiente string) error {

	err := dns_simf.RewriteOldDNS(ambiente)
	if err != nil {
		return err
	}

	err = dns_simf.AddNewDns(ambiente)
	if err != nil {
		return err
	}

	err = AddNewDnsCoreEntries(ambiente)

	if err != nil {
		return err
	}

	return nil
}
