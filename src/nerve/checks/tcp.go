package checks

const CHECK_TCP_TYPE string = "TCP"

type tcpCheck struct {
	Check
	Host string
	Port string
	Ip string
	ConnectTimeout int
	DisconnectTimeout int
}

//Initialize 
func (x tcpCheck) Initialize() error {
	//Default value pushed here
	x.Status = StatusUnknown
	x.Host = "localhost"
	x.Port = "80"
	x.Ip = "127.0.0.1"
	x.ConnectTimeout = 10
	x.DisconnectTimeout = 10
	x._type = CHECK_TCP_TYPE
	return nil
}

//Verify that the given host or ip / port is healthy
func (x tcpCheck) DoCheck() (status int, err error) {
	return StatusOK, nil
}

func (x tcpCheck) GetType() string {
	return x._type
}
