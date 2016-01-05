package reporters

const REPORTER_FILE_TYPE = "FILE"

type fileReporter struct {
	Reporter
	Path string
	Filename string
}

func(x fileReporter) Initialize() error {
	x.Path = "/tmp/"
	x.Filename = "nerve_reporter.txt"
	x._type = REPORTER_FILE_TYPE
	return nil
}

func(x fileReporter) Report(IP string, Port string, Host string, Status int) error {
	return nil
}

func(x fileReporter) GetType() string {
	return x._type
}
