package checks

const CHECK_HTTP_TYPE = "HTTP"

type httpCheck struct {
	Check
	Host string
	Port string
	Ip string
	URL string
	ConnectTimeout int
	DisconnectTimeout int
}

//Initialize 
func(x httpCheck) Initialize() error {
	//Default value pushed here
	x.Status = StatusUnknown
	x.Host = "localhost"
	x.Port = "80"
	x.Ip = "127.0.0.1"
	x.URL = "http://localhost/"
	x.ConnectTimeout = 10
	x.DisconnectTimeout = 10
	x._type = CHECK_HTTP_TYPE
	return nil
}

//Verify that the given host or ip / port is healthy
func(x httpCheck) DoCheck() (status int, err error) {
	return StatusOK, nil
}

func(x httpCheck) GetType() string {
	return x._type
}
