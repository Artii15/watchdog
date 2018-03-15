package dynamo

import (
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"watchdog/checker"
)

type ConfigLoader struct {
	dynamoService *dynamodb.DynamoDB
	tableName string
	primaryKey string
}

func New(awsSession client.ConfigProvider, tableName, primaryKey string) *ConfigLoader {
	dynamoDbService := dynamodb.New(awsSession)
	return &ConfigLoader{dynamoService:dynamoDbService, tableName:tableName, primaryKey:primaryKey}
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
		var checkerConfig checker.Config
		dynamodbattribute.UnmarshalMap(result.Item, &checkerConfig)
		return &checkerConfig, nil
	} else {
		return nil, err
	}
}

