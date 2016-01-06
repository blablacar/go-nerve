package nerve

import (
	"strings"
	"nerve/reporters"
)

type ReporterI interface {
	Initialize() error
	Report(IP string, Port string, Host string, Status int) error
	GetType() string
}

func CreateReporter(config NerveReporterConfiguration) (reporter ReporterI, err error) {
	switch (strings.ToUpper(config.Type)) {
		case reporters.REPORTER_ZOOKEEPER_TYPE:
			reporter = new(reporters.ZookeeperReporter)
		case reporters.REPORTER_FILE_TYPE:
			reporter = new(reporters.FileReporter)
		default:
			reporter = new(reporters.ConsoleReporter)
			reporter.Initialize()
	}
	return reporter, nil
}
