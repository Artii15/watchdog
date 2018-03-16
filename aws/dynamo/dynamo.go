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
	loggersObject *loggers.Loggers
}

func New(awsSession client.ConfigProvider, tableName, primaryKey string, loggersObject *loggers.Loggers) *ConfigLoader {
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
		loader.loggersObject.Info.Println("New configuration fetched from dynamoDb")
		var checkerConfig checker.Config
		dynamodbattribute.UnmarshalMap(result.Item, &checkerConfig)
		return &checkerConfig, nil
	} else {
		loader.loggersObject.Error.Println("Could not load new configuration from dynamo db", err)
		return nil, err
	}
}

