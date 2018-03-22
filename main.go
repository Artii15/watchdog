package main

import (
	"time"
	"github.com/Artii15/watchdog/checker"
	"github.com/Artii15/watchdog/config"
	"github.com/Artii15/watchdog/loggers"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/Artii15/watchdog/aws/s3"
	"github.com/Artii15/watchdog/aws/dynamo"
	"github.com/Artii15/watchdog/aws/sns"
	"github.com/Artii15/watchdog/config/validator"
)

const configRefreshingIntervalInMinutes = 15

func main()  {
	programSettings := config.LoadProgramSettings()
	validatorObject := validator.New()
	validatorObject.Validate(programSettings)
	if validatorObject.HasErrors() {
		fmt.Println("Invalid config. Check program argunents")
	}

	logs, err := loggers.New(programSettings.LoggersConfig)
	if err != nil {
		fmt.Println("Could not create logs. Exiting")
		return
	}

	awsSession, sessionError := session.NewSession()
	var snsNotifier *sns.Notifier
	var dynamoLoader *dynamo.ConfigLoader
	var s3Uploader *s3.Uploader
	if sessionError != nil {
		logs.Error("Could not create aws session", sessionError)
		return
	} else {
		dynamoLoader = dynamo.New(awsSession, programSettings.DynamoDbTableName, programSettings.DynamoDbPrimaryKey, logs)
		snsNotifier = sns.New(awsSession, programSettings.SnsTopic)
		s3Uploader = s3.New(awsSession, programSettings.S3BucketName)
		logs.SetUploader(s3Uploader)
	}

	logs.Info("Fetching config from dynamoDb")
	checkerConfig, dynamoErr := dynamoLoader.ReloadConfig()
	if dynamoErr != nil {
		logs.Error("Could not fetch config from dynamoDb", dynamoErr)
		return
	}

	servicesCheckingTicker := time.NewTicker(time.Duration(checkerConfig.NumOfSecCheck) * time.Second)
	defer servicesCheckingTicker.Stop()

	configCheckingTicker := time.NewTicker(configRefreshingIntervalInMinutes * time.Minute)
	defer servicesCheckingTicker.Stop()

	configChannel := make(chan *checker.Config)
	checkerChannel := make(chan bool)

	servicesChecker := checker.New(snsNotifier, logs, checkerChannel)
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
