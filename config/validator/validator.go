package validator

import (
	"errors"
	"github.com/Artii15/watchdog/config"
)

var (
	EmptyDynamoDbTableName = errors.New("dynamo DB table name is missing")
	EmptyDynamoDbPrimaryKey = errors.New("dynamo DB primary key is missing")
	EmptySnSTopic = errors.New("SnS topic is missing")
	EmptyS3BucketName = errors.New("s3 bucket name is missing")
	EmptyLogsDirPath = errors.New("logs dir path is missing")
	EmptyLogfileSplitThreshold = errors.New("log file split threshold is missing")
)

type Validator struct {
	errors []error
}

func New() *Validator {
	return new(Validator)
}

func (validator *Validator) notEmpty(string string, error error) {
	if string == "" {
		validator.errors = append(validator.errors, error)
	}
}

func (validator *Validator) positive(value int64, error error) {
	if value <= 0 {
		validator.errors = append(validator.errors, error)
	}
}

func (validator *Validator) Validate(settings config.ProgramSettings) {
	validator.errors = nil
	validator.notEmpty(settings.DynamoDbTableName, EmptyDynamoDbTableName)
	validator.notEmpty(settings.DynamoDbPrimaryKey, EmptyDynamoDbPrimaryKey)
	validator.notEmpty(settings.SnsTopic, EmptySnSTopic)
	validator.notEmpty(settings.S3BucketName, EmptyS3BucketName)
	validator.notEmpty(settings.LoggersConfig.LogsDirPath, EmptyLogsDirPath)
	validator.positive(settings.LoggersConfig.LogfileSplitThreshold, EmptyLogfileSplitThreshold)
}

func (validator *Validator) HasErrors() bool {
	return len(validator.errors) > 0
}
