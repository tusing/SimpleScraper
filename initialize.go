package SimpleScraper

func initialize(configPath string, logLevel string) {
	log = makeLogger("SimpleScraper", parseLogLevel(logLevel))
	config = MustConstructConfig(configPath)
}
