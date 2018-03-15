package main

import (
	"watchdog/config"
	"watchdog/loggers"
	"watchdog/aws/dynamo"
	"watchdog/aws/sns"
	"github.com/aws/aws-sdk-go/aws/session"
	"time"
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
	if sessionError != nil {
		loggersObject.Error.Println("Could not create aws session", sessionError)
		panic("Exiting. No AWS session")
	} else {
		dynamoLoader = dynamo.New(awsSession, programSettings.DynamoDbTableName, programSettings.DynamoDbPrimaryKey)
		snsNotifier = sns.New(awsSession, programSettings.SnsTopic)
	}
}

func main()  {
	loggersObject.Info.Println("Fetching config from dynamoDb")
	checkerConfig, dynamoErr := dynamoLoader.ReloadConfig()
	if dynamoErr != nil {
		loggersObject.Error.Println("Could not fetch config from dynamoDb", dynamoErr)
		panic("Retrieving information about configuration was not possible")
	}
	servicesCheckingTicker := time.NewTicker(time.Duration(checkerConfig.NumOfSecCheck) * time.Second)

	for {
		select {
		case <-servicesCheckingTicker.C:
			loggersObject.Info.Println("Checking services")
		}
	}
}
