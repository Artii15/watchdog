package main

import (
	"watchdog/config"
	"log"
	"os"
	"io"
)

const errorLogPrefix  = "ERROR: "
const warningLogPrefix  = "WARNING: "
const infoLogPrefix  = "INFO: "

var errorLogger *log.Logger
var warningLogger *log.Logger
var infoLogger *log.Logger
var programSettings *config.ProgramSettings

func init() {
	programSettings = config.LoadProgramSettings()

	logFile, logError := os.OpenFile(programSettings.LogfileLocation, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if logError != nil {
		errorLogger = createLogger(os.Stderr, errorLogPrefix)
		warningLogger = createLogger(os.Stdout, warningLogPrefix)
		infoLogger = createLogger(os.Stdout, infoLogPrefix)
		errorLogger.Println("Failed to open ", programSettings.LogfileLocation, ". Using stderr and stdout instead")
	} else {
		errorLogger = createLogger(logFile, errorLogPrefix)
		warningLogger = createLogger(logFile, warningLogPrefix)
		infoLogger = createLogger(logFile, infoLogPrefix)
	}
}

func createLogger(writer io.Writer, prefix string) *log.Logger {
	return log.New(writer, prefix, log.Ldate|log.Ltime)
}

func main()  {


}
