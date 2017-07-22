package main

import (
	"github.com/op/go-logging"
	"os"
)

var log logging.Logger

func makeLogger(name string, minLevel logging.Level) logging.Logger {
	log := logging.MustGetLogger(name)
	format := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.8s} %{id:03x}%{color:reset} %{message}`,
	)
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logFormattedBackend := logging.NewBackendFormatter(logBackend, format)
	logLeveledBackend := logging.AddModuleLevel(logFormattedBackend)
	logLeveledBackend.SetLevel(minLevel, "")
	logging.SetBackend(logLeveledBackend)
	return *log
}

func parseLogLevel(arg string) logging.Level {
	levels := map[string]logging.Level{
		"info":     logging.INFO,
		"notice":   logging.NOTICE,
		"warning":  logging.WARNING,
		"error":    logging.ERROR,
		"critical": logging.CRITICAL,
	}
	return levels[arg]
}
