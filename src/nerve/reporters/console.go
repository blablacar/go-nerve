package reporters

import (
	"fmt"
)

const REPORTER_CONSOLE_TYPE = "CONSOLE"

type ConsoleReporter struct {
	Reporter
}

func(x *ConsoleReporter) Initialize() error {
	x._type = REPORTER_CONSOLE_TYPE
	return nil
}

func(x *ConsoleReporter) Report(IP string, Port string, Host string, Status int) error {
	fmt.Printf("{\"report\":{\"IP\":\"%s\",\"Port\":%s,\"Host\":\"%s\",\"Status\":%d}}\n",IP,Port,Host,Status)
	return nil
}

func(x *ConsoleReporter) GetType() string {
	return x._type
}
