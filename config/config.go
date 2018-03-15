package config

import "flag"

const defaultLogFileLocation = "/var/log/watchdog"

type ProgramSettings struct {
	LogfileLocation string
	DynamoDbTableName string
	DynamoDbPrimaryKey string
	SnsTopic string
}

func LoadProgramSettings() *ProgramSettings {
	logfileLocation := flag.String("l", defaultLogFileLocation, "path to logfile")
	dynamoDbTableName := flag.String("t", "", "Dynamo DB table name")
	dynamoDbPrimaryKey := flag.String("p", "", "Primary key of configuration in Dynamo DB")
	snsTopic := flag.String("s", "", "SnS topic name")

	return &ProgramSettings{
		LogfileLocation: *logfileLocation,
		DynamoDbTableName: *dynamoDbTableName,
		DynamoDbPrimaryKey: *dynamoDbPrimaryKey,
		SnsTopic: *snsTopic,
	}
}
