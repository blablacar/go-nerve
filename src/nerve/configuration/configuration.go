package configuration

type NerveReporterConfiguration struct {
	Type string
	Hosts []string
	Path string
}

type NerveCheckConfiguration struct {
	Type string
        Timeout float32
	Rise int
	Fall int
	Uri string
}

type NerveWatcherConfiguration struct {
	Host string
	Port int
	CheckInterval	int
	Checks []NerveCheckConfiguration
}

type NerveServiceConfiguration struct {
	Watcher NerveWatcherConfiguration
	Reporter NerveReporterConfiguration
}

type NerveConfiguration struct {
	InstanceID string
	Services []NerveServiceConfiguration
}
