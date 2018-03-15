package main

import (
	"watchdog/config"
	"watchdog/loggers"
	"watchdog/aws/dynamo"
	"watchdog/aws/sns"
	"github.com/aws/aws-sdk-go/aws/session"
)

var programSettings *config.ProgramSettings
var loggersObject *loggers.Loggers
var dynamoLoader *dynamo.ConfigLoader
var snsNotifier *sns.Notifier

func init() {
	programSettings = config.LoadProgramSettings()
	loggersObject = loggers.New(programSettings.LogfileLocation)
	//TODO verify settings and log errors

	awsSession, sessionError := session.NewSession()
	if sessionError == nil {
		loggersObject.Error.Println("Could not create aws session", sessionError)
	} else {
		dynamoLoader = dynamo.New(awsSession, programSettings.DynamoDbTableName, programSettings.DynamoDbPrimaryKey)
		snsNotifier = sns.New(awsSession, programSettings.SnsTopic)
	}
}

func main()  {
	loggersObject.Info.Println("Starting watchdog")
}
