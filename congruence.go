package pifra

func getConfigurationKey(conf Configuration) string {
	return prettyPrintRegister(conf.Register) + PrettyPrintAst(conf.Process)
}
