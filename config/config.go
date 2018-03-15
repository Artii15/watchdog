package config

import "flag"

const defaultLogFileLocation = "/var/log/watchdog"

type ProgramSettings struct {
	logfileLocation string
	dynamoDbTableName string
	dynamoDbPrimaryKey string
	snsTopic string
}

func LoadProgramSettings() *ProgramSettings {
	logfileLocation := flag.String("l", defaultLogFileLocation, "path to logfile")
	dynamoDbTableName := flag.String("t", "", "Dynamo DB table name")
	dynamoDbPrimaryKey := flag.String("p", "", "Primary key of configuration in Dynamo DB")
	snsTopic := flag.String("s", "", "SnS topic name")

	return &ProgramSettings{
		logfileLocation: *logfileLocation,
		dynamoDbTableName: *dynamoDbTableName,
		dynamoDbPrimaryKey: *dynamoDbPrimaryKey,
		snsTopic: *snsTopic,
	}
}
