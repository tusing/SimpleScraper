package main

import (
	"github.com/jessevdk/go-flags"
	"regexp"
)

var commandLineOpts struct {
	LogLevel              string `short:"l" default:"error" long:"log-level" description:"Log level: <info|notice|warning|error|critical>"`
	URL                   string `short:"u" required:"true" long:"url" description:"URL to scrape, provides shard REQUESTED_URL"`
	ConfigPath            string `short:"c" default:"config.toml" long:"config" description:"Location of config file"`
	InterpolationContains string `short:"i" default:"" default-mask:"-" long:"contains" description:"If multiple interpolations are compatible with your URL, filter by the interpolations containing the given string."`
}

func validateURL(url string) bool {
	validURL := regexp.MustCompile(`^https{0,1}:\/\/www\.`)
	return validURL.Match([]byte(url))
}

func main() {
	if _, err := flags.Parse(&commandLineOpts); err != nil {
		log.Fatal("Could not parse command-line flags!\n")
	}
	logLevel := commandLineOpts.LogLevel
	url := commandLineOpts.URL
	configPath := commandLineOpts.ConfigPath
	interpContains := commandLineOpts.InterpolationContains

	log = makeLogger("loggerName", parseLogLevel(logLevel))
	config := mustConstructConfig(configPath)
	config.initialize()

	if validateURL(url) {
		providedShard := map[string]shard{"REQUESTED_URL": {Override: url, Name: "REQUESTED_URL"}}
		print(constructFromURL(url, config, providedShard, interpContains))
	} else {
		log.Fatalf(`Provided URL is not well formed (i.e. beginning with "http[s]://www.")`)
	}
}
