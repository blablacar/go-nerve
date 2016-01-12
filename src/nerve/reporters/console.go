package reporters

import (
	"fmt"
)

const REPORTER_CONSOLE_TYPE = "CONSOLE"

type ConsoleReporter struct {
	Reporter
}

func(cr *ConsoleReporter) Initialize(IP string, Port int, Rise int, Fall int, Weight int, ServiceName string, InstanceID string, HAProxyServerOptions string) error {
	cr._type = REPORTER_CONSOLE_TYPE
	cr.SetBaseConfiguration(IP,Port,Rise,Fall,Weight,ServiceName,InstanceID,HAProxyServerOptions)
	return nil
}

func(cr *ConsoleReporter) Report(Status int) error {
	if cr.CanReport(Status) {
		fmt.Printf("Data:[%s] Status[%d]\n",cr.GetJsonReporterData(),Status)
	}
	return nil
}

func(cr *ConsoleReporter) Destroy() error {
	return nil
}

func(cr *ConsoleReporter) GetType() string {
	return cr._type
}
