package reporters

import "strings"

const (
	StatusOK int = 0
	StatusKO int = 1
	StatusUnknown int = 2
)

type Reporter struct {
	Error error
	_type string
}

type ReporterI interface {
	Initialize() error
	Report(IP string, Port string, Host string, Status int) error
	GetType() string
}

func CreateReporter(_type string, config []string) (reporter ReporterI, err error) {
	switch (strings.ToUpper(_type)) {
		case REPORTER_ZOOKEEPER_TYPE:
			reporter = new(zookeeperReporter)
		case REPORTER_FILE_TYPE:
			reporter = new(fileReporter)
		default:
			reporter = new(consoleReporter)
			reporter.Initialize()
	}
	return reporter, nil
}
