package main

import (
	"watchdog/config"
	"watchdog/loggers"
	"watchdog/aws/dynamo"
	"watchdog/aws/sns"
	"github.com/aws/aws-sdk-go/aws/session"
	"time"
	"watchdog/checker"
	"os"
	"os/signal"
	"syscall"
)

var programSettings *config.ProgramSettings
var loggersObject *loggers.Loggers
var dynamoLoader *dynamo.ConfigLoader
var snsNotifier *sns.Notifier

const configRefreshingIntervalInMinutes = 15

func init() {
	programSettings = config.LoadProgramSettings()
	loggersObject = loggers.New(programSettings.LogfileLocation)

	awsSession, sessionError := session.NewSession()
	if sessionError != nil {
		loggersObject.Error.Println("Could not create aws session", sessionError)
		panic("Exiting. No AWS session")
	} else {
		dynamoLoader = dynamo.New(awsSession, programSettings.DynamoDbTableName, programSettings.DynamoDbPrimaryKey)
		snsNotifier = sns.New(awsSession, programSettings.SnsTopic)
	}
}

func reloadConfig(configChannel chan<- *checker.Config)  {
	newConfig, dynamoError := dynamoLoader.ReloadConfig()
	if dynamoError == nil {
		loggersObject.Error.Println("Could not load new configuration from dynamo db", dynamoError)
	} else {
		loggersObject.Info.Println("New configuration fetched from dynamoDb")
		configChannel<- newConfig
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
	defer servicesCheckingTicker.Stop()

	configCheckingTicker := time.NewTicker(configRefreshingIntervalInMinutes * time.Minute)
	defer servicesCheckingTicker.Stop()

	configChannel := make(chan *checker.Config)
	signalsChannel := make(chan os.Signal, 1)
    signal.Notify(signalsChannel, syscall.SIGINT, syscall.SIGTERM)

	servicesChecker := checker.New(snsNotifier, loggersObject)
	stopProgram := false
	for !stopProgram {
		select {
		case <-servicesCheckingTicker.C:
			go servicesChecker.Check(checkerConfig)
		case <-configCheckingTicker.C:
			go reloadConfig(configChannel)
		case newConfig := <-configChannel:
			checkerConfig = newConfig
		case <-signalsChannel:
			stopProgram = true
		}
	}
}
