package reporters

import (
	"testing"
)

func TestReporters(t *testing.T) {
	var reporter Reporter
	reporter.IP = "1234"
	reporter.Port = 1234
	reporter.InstanceID = "ApelyAgamakou"
	reporter.Weight = 42
	if reporter.GetJsonReporterData(false) != "{\"host\":\"1234\",\"port\":1234,\"name\":\"ApelyAgamakou\",\"weight\":42,\"maintenance\":false}" {
		t.Error("Reporter JSON serialization failed expect [{\"host\":\"1234\",\"port\":1234,\"name\":\"ApelyAgamakou\",\"weight\":42}], got ", reporter.GetJsonReporterData(false))
	}
}
