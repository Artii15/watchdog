package checker

import (
	"testing"
)

func TestChecker_isServiceInstalled(t *testing.T) {
	if !isServiceInstalled("udev") {
		t.Error("Existing service not found")
	}
}