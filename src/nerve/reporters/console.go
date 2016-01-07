package reporters

import (
	"fmt"
)

const REPORTER_CONSOLE_TYPE = "CONSOLE"

type ConsoleReporter struct {
	Reporter
}

func(x *ConsoleReporter) Initialize(IP string, Port int, Rise int, Fall int) error {
	x._type = REPORTER_CONSOLE_TYPE
	x.SetBaseConfiguration(IP,Port,Rise,Fall)
	return nil
}

func(x *ConsoleReporter) Report(Status int) error {
	if x.CanReport(Status) {
		fmt.Printf("{\"report\":{\"IP\":\"%s\",\"Port\":%d,\"Status\":%d}}\n",x.IP,x.Port,Status)
	}
	return nil
}

func(x *ConsoleReporter) GetType() string {
	return x._type
}
