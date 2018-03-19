package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"time"
	"os"
	"github.com/Artii15/watchdog/checker"
	"github.com/Artii15/watchdog/config"
	"github.com/Artii15/watchdog/loggers"
	"github.com/Artii15/watchdog/aws/dynamo"
	"github.com/Artii15/watchdog/aws/sns"
)

const configRefreshingIntervalInMinutes = 15

func main()  {
	programSettings := config.LoadProgramSettings()

	var logfile, logfileErr = os.OpenFile(programSettings.LogfileLocation, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
	var loggersObject *loggers.Logs
	if logfileErr == nil {
		loggersObject = loggers.New(logfile)
		defer logfile.Close()
	} else {
		loggersObject = loggers.New(os.Stdout)
		loggersObject.Warning.Println("Could not open", logfile, "using stdout instead")
	}

	awsSession, sessionError := session.NewSession()
	var snsNotifier *sns.Notifier
	var dynamoLoader *dynamo.ConfigLoader
	if sessionError != nil {
		loggersObject.Error.Println("Could not create aws session", sessionError)
		panic("Exiting. No AWS session")
	} else {
		dynamoLoader = dynamo.New(awsSession, programSettings.DynamoDbTableName, programSettings.DynamoDbPrimaryKey, loggersObject)
		snsNotifier = sns.New(awsSession, programSettings.SnsTopic)
	}

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
	checkerChannel := make(chan bool)

	servicesChecker := checker.New(snsNotifier, loggersObject, checkerChannel)
	go servicesChecker.Check(checkerConfig)
	areServicesBeingChecked := true
	for {
		select {
		case <-servicesCheckingTicker.C:
			if !areServicesBeingChecked {
				go servicesChecker.Check(checkerConfig)
				areServicesBeingChecked = true
			}
		case <-configCheckingTicker.C:
			go func() {
				reloadedConfig, err := dynamoLoader.ReloadConfig()
				if err == nil {
					configChannel<- reloadedConfig
				}
			}()
		case newConfig := <-configChannel:
			checkerConfig = newConfig
		case <-checkerChannel:
			areServicesBeingChecked = false
		}
	}
}
