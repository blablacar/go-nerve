package reporters

const REPORTER_FILE_TYPE = "FILE"

type FileReporter struct {
	Reporter
	Path string
	Filename string
}

func(fr *FileReporter) Initialize(IP string, Port int, Rise int, Fall int, Weight int, ServiceName string, InstanceID string, HAProxyServerOptions string) error {
	fr.Path = "/tmp/"
	fr.Filename = "nerve_reporter.txt"
	fr._type = REPORTER_FILE_TYPE
	fr.SetBaseConfiguration(IP,Port,Rise,Fall,Weight,ServiceName,InstanceID,HAProxyServerOptions)
	return nil
}

func(fr *FileReporter) SetFileConfiguration(Path string, Filename string, Mode string) {
	fr.Path = Path
	fr.Filename = Filename
	fr.Mode = mode
}

func(fr *FileReporter) Report(Status int) error {
	return nil
}

func(fr *FileReporter) Destroy() error {
	return nil
}

func(fr *FileReporter) GetType() string {
	return fr._type
}
