package checks

const CHECK_HTTP_TYPE = "HTTP"

type httpCheck struct {
	Check
	URL string
}

//Initialize 
func(x *httpCheck) Initialize() error {
	//Default value pushed here
	x.Status = StatusUnknown
	x.Host = "localhost"
	x.Port = 80
	x.IP = "127.0.0.1"
	x.URL = "http://localhost/"
	x.ConnectTimeout = 100
	x.DisconnectTimeout = 100
	x._type = CHECK_HTTP_TYPE
	return nil
}

//Verify that the given host or ip / port is healthy
func(x *httpCheck) DoCheck() (status int, err error) {
	return StatusOK, nil
}

func(x *httpCheck) GetType() string {
	return x._type
}

func (x *httpCheck) SetBaseConfiguration(IP string, Host string, Port int, ConnectTimeout int, ipv6 bool) {
	x.IP = IP
	x.Host = Host
	x.Port = Port
	x.ConnectTimeout = ConnectTimeout
	x.IPv6 = ipv6
}
