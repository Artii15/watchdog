package main

import (
	"watchdog/config"
	"watchdog/loggers"
)

var programSettings *config.ProgramSettings
var loggersObject *loggers.Loggers

func init() {
	programSettings = config.LoadProgramSettings()
	loggersObject = loggers.New(programSettings.LogfileLocation)
}

func main()  {

}
