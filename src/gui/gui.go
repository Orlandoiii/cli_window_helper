package gui

import (
	"bufio"
	"cli_window_helper/src/app_log"
	"cli_window_helper/src/appsettings_simf"
	"cli_window_helper/src/argus"
	"cli_window_helper/src/config"
	"cli_window_helper/src/dns_simf"
	"cli_window_helper/src/update_services"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/huh"
)

const DNS_MANAGER = "dns"
const RECOLECTAR_INFO = "recolectar_info"
const FIX_ARGUS = "fix_argus"
const DEPLOYMENT_MANAGER = "deployment_manager"

var GlobalTheme = huh.ThemeBase16()

func clearConsole() {
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		// // Enable virtual terminal processing
		// kernel32 := syscall.NewLazyDLL("kernel32.dll")
		// proc := kernel32.NewProc("SetConsoleMode")
		// handle, _, _ := kernel32.NewProc("GetStdHandle").Call(^uintptr(10)) // STD_OUTPUT_HANDLE = -11

		// var mode uint32
		// syscall.GetConsoleMode(syscall.Handle(handle), &mode)
		// mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
		// proc.Call(handle, uintptr(mode))

		fmt.Print("\033[H\033[2J")
	default:
		fmt.Print("\n\n\n\n\n")
	}
}

func EnviromentOptionsForm(all bool, actualEnvironment string) []huh.Option[string] {
	options := []huh.Option[string]{}

	if all {
		options = append(options, huh.NewOption("Desarrollo", config.AMBIENTE_DESARROLLO))
		options = append(options, huh.NewOption("Certificacion", config.AMBIENTE_CERTIFICACION))
		options = append(options, huh.NewOption("Produccion", config.AMBIENTE_PRODUCCION))
	} else {
		if actualEnvironment != config.AMBIENTE_DESARROLLO {
			options = append(options, huh.NewOption("Desarrollo", config.AMBIENTE_DESARROLLO))
		}
		if actualEnvironment != config.AMBIENTE_CERTIFICACION {
			options = append(options, huh.NewOption("Certificacion", config.AMBIENTE_CERTIFICACION))
		}
		if actualEnvironment != config.AMBIENTE_PRODUCCION {
			options = append(options, huh.NewOption("Produccion", config.AMBIENTE_PRODUCCION))
		}
	}

	return options
}

func EnviromentBaseForm(all bool, actualEnvironment string) (string, error) {
	clearConsole()

	var environment string
	confirmacion := false

	form := huh.NewForm(
		huh.NewGroup(

			huh.NewSelect[string]().
				Title("Seleccione el Ambiente a Usar").
				Options(
					EnviromentOptionsForm(all, actualEnvironment)...,
				).
				Value(&environment),

			huh.NewConfirm().
				TitleFunc(func() string {
					return fmt.Sprintf("¿Estas Seguro de usar el Ambiente => %s", environment)
				}, &environment).Affirmative("Si").Negative("No").Value(&confirmacion),
		),
	).WithTheme(GlobalTheme)

	err := form.Run()
	if err != nil {
		app_log.Get().Printf("Error al Seleccionar Ambiente => %v", err)
		return "", err
	}

	if confirmacion {
		//config.SetAmbiente(environment)
		return environment, nil
	}

	return "", nil
}

func AmbienteForm() (string, error) {

	environment, moreThanOne := dns_simf.DetectEnvironment()
	confirmacion := false

	if !moreThanOne && environment != "" {

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().Title(fmt.Sprintf("Ambiente Detectado => %s", environment)),

				huh.NewConfirm().
					TitleFunc(func() string {
						return fmt.Sprintf("¿Estas Seguro de usar el Ambiente => %s", environment)
					}, &environment).Affirmative("Si").Negative("No").Value(&confirmacion),
			),
		).WithTheme(GlobalTheme)

		err := form.Run()

		if err != nil {
			app_log.Get().Printf("Error al Seleccionar Ambiente => %v", err)
			return "", err
		}

		if confirmacion {
			config.SetAmbiente(environment)
			return environment, nil
		}
	}

	if moreThanOne || !confirmacion || environment == "" {

		app_log.Get().Printf("Mas de un Ambiente Detectado => %s", environment)

		environment, err := EnviromentBaseForm(true, "")
		if err != nil {
			app_log.Get().Printf("Error al Seleccionar Ambiente => %v", err)
			return "", err
		}

		return environment, nil
	}

	return "", nil
}

func init() {

}

func MainForm() {

	environment, err := AmbienteForm()

	if err != nil {

		app_log.Get().Printf("Error al Seleccionar Ambiente => %v", err)

		panic(err)
	}

	config.SetAmbiente(environment)

	_ = config.GetAmbiente()

	var opcionSeleccionada string

	mainForm := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("Elija una Opcion"),
			huh.NewSelect[string]().
				Title("Seleccione una Opcion").
				Options(
					huh.NewOption("Control de Cambio CC27", DEPLOYMENT_MANAGER),
					huh.NewOption("Recolectar Info", RECOLECTAR_INFO),
					huh.NewOption("Argus Ajustar DNS y AppSettings", FIX_ARGUS),
				).
				Value(&opcionSeleccionada),
		),
	).WithTheme(GlobalTheme)

	clearConsole()

	err = mainForm.Run()
	if err != nil {
		app_log.Get().Printf("Error al Mostrar Menu Principal => %v", err)
		panic(err)
	}

	switch opcionSeleccionada {
	case DEPLOYMENT_MANAGER:
		option, err := DeploymentManagerForm()
		if err != nil {
			app_log.Get().Printf("Error al Mostrar Menu de Control de Cambio CC27 => %v", err)
			panic(err)
		}

		switch option {

		case UPDATE_WINDOW_EXPORTER:
			err = update_services.UpdateWindowExporter()
			if err != nil {
				app_log.Get().Printf("Error al Actualizar Servicio Windows Exporter => %v", err)
				panic(err)
			}
		case ADJUST_DNS:
			err = ManageDeploymentDNS(config.GetAmbiente())
			if err != nil {
				app_log.Get().Printf("Error al Ajustar DNS => %v", err)
				panic(err)
			}

		case ADJUST_DNS_APPSETTINGS:
			err = appsettings_simf.ChangeAppSettings(config.GetAmbiente())
			if err != nil {
				app_log.Get().Printf("Error al Ajustar AppSettings => %v", err)
				panic(err)
			}
		case CHANGE_DNS_ENVIRONMENT:
			newEnvironment, err := EnviromentBaseForm(false, config.GetAmbiente())

			if err != nil {
				app_log.Get().Printf("Error al Seleccionar Nuevo Ambiente => %v", err)
				panic(err)
			}

			err = dns_simf.ChangeEnvironment(config.GetAmbiente(), newEnvironment)
			if err != nil {
				app_log.Get().Printf("Error al Cambiar Ambiente DNS => %v", err)
				panic(err)
			}

			// err = appsettings_simf.ChangeAppSettingsEnviroment(newEnvironment)
			// if err != nil {
			// 	app_log.Get().Printf("Error al Cambiar Ambiente AppSettings => %v", err)
			// 	panic(err)
			// }
		}

	case DNS_MANAGER:
		DnsManagerForm()
	case RECOLECTAR_INFO:
		RecolectarInfoForm()
	case FIX_ARGUS:
		argus.ArgusDeployDnsAndAppSettings(config.GetAmbiente())
	}
}
func dns_test() {
	fmt.Println("Cli_Window_Helper")

	modifications := make(map[string]string)
	modifications["simf"] = "todo.simf"

	entries, err := dns_simf.LoadHostsFile("./hosts", []string{"simf", "grafana", "lbtr"})
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {

		entry.PrintInfo()
		fmt.Println("--------------------------------")
	}
}

func WaitForEnter() {
	fmt.Print("\nPresione Enter para continuar...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
