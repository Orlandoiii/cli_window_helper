package argus

import (
	"cli_window_helper/src/appsettings_simf"
	"cli_window_helper/src/dns_simf"
)

func ArgusDeployDnsAndAppSettings(environment string) error {

	err := dns_simf.AddArgusDns(environment)

	if err != nil {

		return err
	}

	err = appsettings_simf.ArgusWindowFixAppSettings(environment)

	if err != nil {

		return err
	}

	return nil
}
