package reporters

import "fmt"
import log "github.com/Sirupsen/logrus"

const REPORTER_CONSOLE_TYPE = "CONSOLE"

type consoleReporter struct {
	Reporter
}

func(x consoleReporter) Initialize() error {
	x._type = REPORTER_CONSOLE_TYPE
	return nil
}

func(x consoleReporter) Report(IP string, Port string, Host string, Status int) error {
	log.Info("{\"report\":{\"IP\":\"", IP, "\",\"Port\":", Port, ",\"Host\":", Host, ",\"Status\":", Status, "}}")
	fmt.Println("{\"report\":{\"IP\":\"", IP, "\",\"Port\":", Port, ",\"Host\":", Host, ",\"Status\":", Status, "}}")
	return nil
}

func(x consoleReporter) GetType() string {
	return x._type
}
