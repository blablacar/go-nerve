package nerve

import (
	"strings"
	"nerve/reporters"
)

type ReporterI interface {
	Initialize(IP string, Port int, Rise int, Fall int, Weight int, InstanceID string) error
	Report(Status int) error
	Destroy() error
	GetType() string
}

func CreateReporter(IP string, Port int, ServiceName string, InstanceID string, config NerveReporterConfiguration, ipv6 bool) (reporter ReporterI, err error) {
	switch (strings.ToUpper(config.Type)) {
		case reporters.REPORTER_ZOOKEEPER_TYPE:
			_reporter := new(reporters.ZookeeperReporter)
			_reporter.SetZKConfiguration(config.Hosts,config.Path)
			_reporter.SetServiceName(ServiceName)
			reporter = _reporter
		case reporters.REPORTER_FILE_TYPE:
			reporter = new(reporters.FileReporter)
		default:
			reporter = new(reporters.ConsoleReporter)
	}
	reporter.Initialize(IP,Port,config.Rise,config.Fall,config.Weight,InstanceID)
	return reporter, nil
}
