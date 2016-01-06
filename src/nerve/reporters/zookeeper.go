package reporters

const REPORTER_ZOOKEEPER_TYPE string = "ZOOKEEPER"

type ZookeeperReporter struct {
	Reporter
	ZKHost string
	ZKPort string
	ZKPath string
	ZKKey string
}

func(x ZookeeperReporter) Initialize() error {
	x.ZKHost = "localhost"
	x.ZKPort = "443"
	x.ZKPath = "nerve"
	x.ZKKey = "reporter"
	x._type = REPORTER_ZOOKEEPER_TYPE
	return nil
}

func(x ZookeeperReporter) Report(IP string, Port string, Host string, Status int) error {
	return nil
}

func(x ZookeeperReporter) GetType() string {
	return x._type
}
