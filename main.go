package main

import (
	"github.com/jessevdk/go-flags"
	"regexp"
)

var commandLineOpts struct {
	LogLevel              string `short:"l" default:"error" long:"log-level" description:"Log level: <info|notice|warning|error|critical>"`
	URL                   string `short:"u" required:"true" long:"url" description:"URL to scrape, provides shard REQUESTED_URL"`
	ConfigPath            string `short:"c" default:"config.toml" long:"config" description:"Location of config file"`
	InterpolationContains string `short:"i" default:"" default-mask:"-" long:"contains" description:"If multiple interpolations are compatible with your URL, filter by the interpolations containig the given string."`
}

func validateURL(url string) bool {
	validURL, _ := regexp.Compile(`^https{0,1}:\/\/www\.`)
	if !validURL.Match([]byte(url)) {
		return false
	}
	return true
}

func main() {
	flags.Parse(&commandLineOpts)
	logLevel := commandLineOpts.LogLevel
	url := commandLineOpts.URL
	configPath := commandLineOpts.ConfigPath
	interpContains := commandLineOpts.InterpolationContains

	log = makeLogger("loggerName", parseLogLevel(logLevel))
	config := constructConfig(configPath)
	config.initialize()

	if validateURL(url) {
		providedShard := map[string]shard{"REQUESTED_URL": {Override: url, Name: "REQUESTED_URL"}}
		print(constructFromURL(url, config, providedShard, interpContains))
	} else {
		log.Fatalf(`Provided URL is not well formed (i.e. beginning with "http[s]://www.")`)
	}
}
