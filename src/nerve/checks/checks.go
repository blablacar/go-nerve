package checks

import (
	"strings"
)

const (
	StatusOK int = 0
	StatusKO int = 1
	StatusUnknown int = 2
)

type Check struct {
	Host string
	IP string
	Port int
	Status int
	ConnectTimeout int
	DisconnectTimeout int
	IPv6 bool
	Error error
	_type string
}

type CheckI interface {
	Initialize() error
	DoCheck() (status int, err error)
	GetType() string
	SetBaseConfiguration(IP string, Host string, Port int, ConnectTimeout int, IPv6 bool)
}

func CreateCheck(_type string, IP string, Host string, Port int, ConnectTimeout int, ipv6 bool) (CheckI, error) {
	var check CheckI
        switch (strings.ToUpper(_type)) {
                case CHECK_TCP_TYPE:
                        check = new(tcpCheck)
			check.Initialize()
                case CHECK_HTTP_TYPE:
                        check = new(httpCheck)
			check.Initialize()
                case CHECK_RABBITMQ_TYPE:
                        check = new(rabbitmqCheck)
			check.Initialize()
                default:
                        check = new(tcpCheck)
                        check.Initialize()
        }
	check.SetBaseConfiguration(IP,Host,Port,ConnectTimeout,ipv6)
        return check, nil
}
