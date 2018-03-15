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
	logfileLocation := flag.String("logfile", defaultLogFileLocation, "path to logfile")
	dynamoDbTableName := flag.String("dynamo-table", "", "Dynamo DB table name")
	dynamoDbPrimaryKey := flag.String("dynamo-key", "", "Primary key of configuration in Dynamo DB")
	snsTopic := flag.String("sns", "", "SnS topic name")

	return &ProgramSettings{
		LogfileLocation: *logfileLocation,
		DynamoDbTableName: *dynamoDbTableName,
		DynamoDbPrimaryKey: *dynamoDbPrimaryKey,
		SnsTopic: *snsTopic,
	}
}
