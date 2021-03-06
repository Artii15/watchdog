package validator

import (
	"testing"
	"github.com/Artii15/watchdog/loggers"
	"github.com/Artii15/watchdog/config"
	"errors"
)

var invalidSettings = config.ProgramSettings{
	S3BucketName: "",
	SnsTopic: "",
	DynamoDbPrimaryKey: "",
	DynamoDbTableName: "",
	LoggersConfig: loggers.Config{LogfileSplitThreshold: 0, LogsDirPath: ""},
}

func TestValidator_Validate(t *testing.T) {
	const expectedErrorsCount = 6
	validator := New()

	validator.Validate(invalidSettings)
	if !validator.HasErrors() || len(validator.errors) != expectedErrorsCount {
		t.Error("Validator should have errors")
	}
}

func TestValidator_positive(t *testing.T) {
	validator := New()

	expectedError := errors.New("value not positive")
	validator.positive(0, expectedError)

	if len(validator.errors) != 1 || validator.errors[0] != expectedError {
		t.Error("Expected error", expectedError)
	}
}

func TestValidator_notEmpty(t *testing.T) {
	validator := New()

	expectedError := errors.New("value is empty")
	validator.notEmpty("", expectedError)

	if len(validator.errors) != 1 || validator.errors[0] != expectedError {
		t.Error("Expected error", expectedError)
	}
}
