package reporters

const REPORTER_ZOOKEEPER_TYPE string = "ZOOKEEPER"

type ZookeeperReporter struct {
	Reporter
	ZKHost string
	ZKPort string
	ZKPath string
	ZKKey string
}


func(x *ZookeeperReporter) Initialize(IP string, Port int, Rise int, Fall int) error {
	x.ZKHost = "localhost"
	x.ZKPort = "443"
	x.ZKPath = "nerve"
	x.ZKKey = "reporter"
	x._type = REPORTER_ZOOKEEPER_TYPE
	x.SetBaseConfiguration(IP,Port,Rise,Fall)
	return nil
}

func(x ZookeeperReporter) Report(Status int) error {
	return nil
}

func(x ZookeeperReporter) GetType() string {
	return x._type
}
