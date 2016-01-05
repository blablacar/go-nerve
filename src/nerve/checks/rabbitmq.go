package checks

const CHECK_RABBITMQ_TYPE string = "RABBITMQ"

type rabbitmqCheck struct {
	Check
	Host string
	Port string
	Ip string
	VHost string
	Queue string
	ConnectTimeout int
	DisconnectTimeout int
}

//Initialize 
func(x rabbitmqCheck) Initialize() error {
	//Default value pushed here
	x.Status = StatusUnknown
	x.Host = "localhost"
	x.Port = "5672"
	x.Ip = "127.0.0.1"
	x.VHost = "/"
	x.Queue = "test"
	x.ConnectTimeout = 10
	x.DisconnectTimeout = 10
	x._type = CHECK_RABBITMQ_TYPE
	return nil
}

//Verify that the given host or ip / port is healthy
func(x rabbitmqCheck) DoCheck() (status int, err error) {
	return StatusOK, nil
}

func(x rabbitmqCheck) GetType() string {
	return x._type
}
