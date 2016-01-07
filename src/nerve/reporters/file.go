package reporters

const REPORTER_FILE_TYPE = "FILE"

type FileReporter struct {
	Reporter
	Path string
	Filename string
}

func(x *FileReporter) Initialize(IP string, Port int, Rise int, Fall int) error {
	x.Path = "/tmp/"
	x.Filename = "nerve_reporter.txt"
	x._type = REPORTER_FILE_TYPE
	x.SetBaseConfiguration(IP,Port,Rise,Fall)
	return nil
}

func(x *FileReporter) Report(Status int) error {
	return nil
}

func(x *FileReporter) GetType() string {
	return x._type
}
