package nerve

import (
	"strings"
	"nerve/reporters"
)

type ReporterI interface {
	Initialize(IP string, Port int, Rise int, Fall int) error
	Report(Status int) error
	GetType() string
}

func CreateReporter(IP string, Port int, config NerveReporterConfiguration, ipv6 bool) (reporter ReporterI, err error) {
	switch (strings.ToUpper(config.Type)) {
		case reporters.REPORTER_ZOOKEEPER_TYPE:
			reporter = new(reporters.ZookeeperReporter)
		case reporters.REPORTER_FILE_TYPE:
			reporter = new(reporters.FileReporter)
		default:
			reporter = new(reporters.ConsoleReporter)
	}
	reporter.Initialize(IP,Port,config.Rise,config.Fall)
	return reporter, nil
}
