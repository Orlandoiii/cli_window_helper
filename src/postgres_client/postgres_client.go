package postgres_client

import (
	"cli_window_helper/src/app_log"
	"cli_window_helper/src/config"
	"context"

	"sync"

	"github.com/jackc/pgx/v5"
)

const (
	AltoValorConfig = "config_inicial_alto_valor"
	BajoValorConfig = "config_inicial_bajo_valor"
)

var queryConfigInicialSimf string = "SELECT configs FROM simf.log_insertdynamic LIMIT 1"
var queryConfigInicialLbtr string = "SELECT configs FROM sglbtr.log_insertdynamic LIMIT 1"

func Conectar(ctx context.Context) (*pgx.Conn, error) {
	config, err := pgx.ParseConfig(config.Get().DireccionPostgres)
	if err != nil {
		return nil, err
	}
	return pgx.ConnectConfig(ctx, config)
}

func Desconectar(conexion *pgx.Conn, ctx context.Context) error {
	if conexion != nil {
		return conexion.Close(context.Background())
	}
	return nil
}

func getConfigInicial(query string) (string, error) {
	conexion, err := Conectar(context.Background())
	if err != nil {
		app_log.Get().Println("Error Generando Conexion: ", err)
		return "", err
	}
	defer conexion.Close(context.Background())
	var result string
	err = conexion.QueryRow(context.Background(), query).Scan(&result)
	if err != nil {
		app_log.Get().Println("Error Ejecutando Query : ", err, query)
		return "", err
	}
	return result, nil
}
func getConfig(result *string, query string, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	var err error
	*result, err = getConfigInicial(query)
	if err != nil {
		app_log.Get().Printf("Error Obteniendo Config Inicial De Alto Valor => %v\n", err)
		*result = err.Error()
	}
}

func GetConfigInicial() (map[string]string, error) {
	result := make(map[string]string)
	var grupoDeEspera sync.WaitGroup

	configAltoValor := ""
	grupoDeEspera.Add(1)
	go getConfig(&configAltoValor, queryConfigInicialLbtr, &grupoDeEspera)

	configBajoValor := ""
	grupoDeEspera.Add(1)
	go getConfig(&configBajoValor, queryConfigInicialSimf, &grupoDeEspera)

	grupoDeEspera.Wait()

	result[AltoValorConfig] = configAltoValor
	result[BajoValorConfig] = configBajoValor

	return result, nil
}
