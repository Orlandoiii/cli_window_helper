package appsettings_simf

const DATA_BASE_FIELD = "DATABASE"
const KAFKA_FIELD = "KAFKA"

const CORE_DEBITO_FIELD = "CORE_DEBITO"
const CORE_CREDITO_FIELD = "CORE_CREDITO"
const CORE_REVERSO_FIELD = "CORE_REVERSO"

const CORE_INTER_FIELD = "CORE_INTER"
const CORE_INTRA_FIELD = "CORE_INTRA"
const CORE_EFECTIVO_FIELD = "CORE_EFECTIVO"

const DESARROLLO = "desarrollo"
const PRODUCCION = "produccion"
const CERTIFICACION = "certificacion"

var ENDPOINTS_CORE = []string{
	CORE_DEBITO_FIELD,
	CORE_CREDITO_FIELD,
	CORE_REVERSO_FIELD,

	CORE_INTER_FIELD,
	CORE_INTRA_FIELD,
	CORE_EFECTIVO_FIELD,
}

var ChangeDnsFields = map[string]string{
	DATA_BASE_FIELD: "DataBaseSettings.ConnectionString",
	KAFKA_FIELD:     "KafkaSettings.BootstrapServers",

	CORE_DEBITO_FIELD:  "HTTPSettings.EndPointCoreDebito",
	CORE_CREDITO_FIELD: "HTTPSettings.EndPointCoreCredito",
	CORE_REVERSO_FIELD: "HTTPSettings.EndPointCoreReversos",

	CORE_INTER_FIELD:    "HTTPSettings.EndPointCoreInter",
	CORE_INTRA_FIELD:    "HTTPSettings.EndPointCoreIntra",
	CORE_EFECTIVO_FIELD: "HTTPSettings.EndPointCoreEfectivo",
}
