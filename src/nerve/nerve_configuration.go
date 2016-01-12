package nerve

import (
	"encoding/json"
	"io/ioutil"
)

type NerveReporterConfiguration struct {
	Type string `json:"type"`
	Hosts []string `json:"hosts"`
	Path string `json:"path"`
	Rise int `json:"rise"`
	Fall int `json:"fall"`
	Weight int `json:"weight"`
	HAProxyServerOptions string `json:"haproxy_server_options"`
}

type NerveCheckConfiguration struct {
	Type string `json:"type"`
        Timeout float32 `json:"timeout"`
	Uri string `json:"uri"`
	User string `json:"user"`
	Password string `json:"password"`
	ConnectTimeout int `json:"timeout"`
	VHost string `json:"vhost"`
	Exchange string `json:"exchange"`
	BindName string `json:"bind_name"`
	Queue string `json:"queue"`
}

type NerveWatcherConfiguration struct {
	CheckInterval int `json:"check_interval"`
	Checks []NerveCheckConfiguration `json:"checks"`
}

type NerveServiceConfiguration struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int `json:"port"`
	Watcher NerveWatcherConfiguration `json:"watcher"`
	Reporter NerveReporterConfiguration `json:"reporter"`
}

type NerveConfiguration struct {
	InstanceID string `json:"instance_id"`
	LogLevel string `json:"log-level"`
	IPv6 bool `json:"ipv6"`
	Services []NerveServiceConfiguration `json:"services"`
}

// Open Nerve configuration file, and parse it's JSON content
// return a full configuration object and an error
// if the error is different of nil, then the configuration object is empty
// if error is equal to nil, all the JSON content of the configuration file is loaded into the object
func OpenConfiguration(fileName string) (NerveConfiguration, error) {
	var nerveConfiguration NerveConfiguration

	// Open and read the configuration file
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		// If there is an error with opening or reading the configuration file, return the error, and an empty configuration object
		return nerveConfiguration, err
	}

	// Trying to convert the content of the configuration file (theoriticaly in JSON) into a configuration object
	err = json.Unmarshal(file, &nerveConfiguration)
	if err != nil {
		// If there is an error in decoding the JSON entry into configuration object, return a partialy unmarshalled object, and the error
		return nerveConfiguration, err
	}

	return nerveConfiguration, nil
}
