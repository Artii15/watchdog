package config

import (
	"flag"
	"github.com/Artii15/watchdog/loggers"
)

type ProgramSettings struct {
	DynamoDbTableName string
	DynamoDbPrimaryKey string
	SnsTopic string
	LoggersConfig loggers.Config
}

func LoadProgramSettings() *ProgramSettings {
	dynamoDbTableName := flag.String("dynamo-table", "", "Dynamo DB table name")
	dynamoDbPrimaryKey := flag.String("dynamo-key", "", "Primary key of configuration in Dynamo DB")
	snsTopic := flag.String("sns", "", "SnS topic name")
	logsDirPath := flag.String("logs-dir", "", "path to directory storing log files")
	logfileSplitThreshold := flag.Int64("logfile-split-threshold", 0, "logfile size at which log file gonna be split")
	flag.Parse()

	return &ProgramSettings{
		DynamoDbTableName: *dynamoDbTableName,
		DynamoDbPrimaryKey: *dynamoDbPrimaryKey,
		SnsTopic: *snsTopic,
		LoggersConfig: loggers.Config{LogsDirPath: *logsDirPath, LogfileSplitThreshold: *logfileSplitThreshold},
	}
}
