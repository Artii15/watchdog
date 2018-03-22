package dynamo

import (
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/Artii15/watchdog/loggers"
	"github.com/Artii15/watchdog/checker"
)

type ConfigLoader struct {
	dynamoService *dynamodb.DynamoDB
	tableName string
	primaryKey string
	loggersObject *loggers.Logs
}

func New(awsSession client.ConfigProvider, tableName, primaryKey string, loggersObject *loggers.Logs) *ConfigLoader {
	dynamoDbService := dynamodb.New(awsSession)
	return &ConfigLoader{dynamoService:dynamoDbService, tableName:tableName, primaryKey:primaryKey, loggersObject:loggersObject}
}

func (loader *ConfigLoader) ReloadConfig() (*checker.Config, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(loader.primaryKey),
			},
		},
		TableName: aws.String(loader.tableName),
	}

	result,err := loader.dynamoService.GetItem(input)
	if err == nil {
		loader.loggersObject.Info("New configuration fetched from dynamoDb")
		var checkerConfig checker.Config
		err = dynamodbattribute.UnmarshalMap(result.Item, &checkerConfig)
		if err != nil {
			return nil, err
		}

		return &checkerConfig, nil
	} else {
		loader.loggersObject.Error("Could not load new configuration from dynamo db", err)
		return nil, err
	}
}

