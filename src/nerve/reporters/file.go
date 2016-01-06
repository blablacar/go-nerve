package reporters

const REPORTER_FILE_TYPE = "FILE"

type FileReporter struct {
	Reporter
	Path string
	Filename string
}

func(x *FileReporter) Initialize() error {
	x.Path = "/tmp/"
	x.Filename = "nerve_reporter.txt"
	x._type = REPORTER_FILE_TYPE
	return nil
}

func(x *FileReporter) Report(IP string, Port string, Host string, Status int) error {
	return nil
}

func(x *FileReporter) GetType() string {
	return x._type
}
