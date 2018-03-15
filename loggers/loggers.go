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

func New(logFile *os.File) *Loggers {
	var loggers Loggers
	loggers.Error = createLogger(logFile, errorLogPrefix)
	loggers.Warning = createLogger(logFile, warningLogPrefix)
	loggers.Info = createLogger(logFile, infoLogPrefix)
	return &loggers
}

func createLogger(writer io.Writer, prefix string) *log.Logger {
	return log.New(writer, prefix, log.Ldate|log.Ltime)
}
