package loggers

import (
	"log"
	"io"
	"os"
)

const errorLogPrefix  = "ERROR: "
const warningLogPrefix  = "WARNING: "
const infoLogPrefix  = "INFO: "

type Loggers struct {
	Info *log.Logger
	Warning *log.Logger
	Error *log.Logger
}

func New(logfileLocation string) *Loggers {
	logFile, logError := os.OpenFile(logfileLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	var loggers Loggers
	if logError != nil {
		loggers.Error = createLogger(os.Stderr, errorLogPrefix)
		loggers.Warning = createLogger(os.Stdout, warningLogPrefix)
		loggers.Info = createLogger(os.Stdout, infoLogPrefix)
		loggers.Error.Println("Failed to open", logfileLocation + ".", "Using stderr and stdout instead")
	} else {
		loggers.Error = createLogger(logFile, errorLogPrefix)
		loggers.Warning = createLogger(logFile, warningLogPrefix)
		loggers.Info = createLogger(logFile, infoLogPrefix)
	}
	return &loggers
}

func createLogger(writer io.Writer, prefix string) *log.Logger {
	return log.New(writer, prefix, log.Ldate|log.Ltime)
}
