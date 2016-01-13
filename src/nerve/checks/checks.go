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
// Create a Check object
// where:
// if _type == tcp
//	param1,param2,param3 are ignored
// if _type == http
//	param1 is the uri to check
//	param2,param3 are ignored
// if _type == mysql
//	param1 is the username
//	param2 is the password
//	param3 is the sql request
func CreateCheck(_type string, IP string, Host string, Port int, ConnectTimeout int, ipv6 bool, param1 string, param2 string, param3 string) (CheckI, error) {
	var check CheckI
        switch (strings.ToUpper(_type)) {
                case CHECK_TCP_TYPE:
			check = new(tcpCheck)
			check.Initialize()
                case CHECK_HTTP_TYPE:
			http_check := new(httpCheck)
			http_check.Initialize()
			http_check.SetURI(param1)
                        check = http_check
                case CHECK_MYSQL_TYPE:
                        mysql_check := new(mysqlCheck)
			mysql_check.Initialize()
			mysql_check.SetMysqlConfiguration(param1,param2,param3)
                        check = mysql_check
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
