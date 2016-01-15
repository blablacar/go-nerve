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
//	param2,param3,param4,param5 are ignored
// if _type == mysql
//	param1 is the username
//	param2 is the password
//	param3 is the sql request
//	param4,param5 are ignored
// if _type == rabbitmq
//	param1 is the username
//	param2 is the password
//	param3 is the queue name
//	param4 is the vhost name
//	param5 is ignored
// if _type == zkflag
//	param5 is the list of hosts
//	param1 is the path to check
//	param2,param3,param3 are ignored
// if _type == httpproxy
//	param1 is the username
//	param2 is the password
//	param3 is the host
//	param4 is the port
//	param5 is the urls to checks
func CreateCheck(_type string, IP string, Host string, Port int, ConnectTimeout int, ipv6 bool, param1 string, param2 string, param3 string, param4 string, param5 []string) (CheckI, error) {
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
                        rabbitmq_check := new(rabbitmqCheck)
			rabbitmq_check.Initialize()
			rabbitmq_check.SetRabbitMQConfiguration(param1,param2,param3,param4)
			check = rabbitmq_check
                case CHECK_ZKFLAG_TYPE:
                        zkflag_check := new(zkflagCheck)
			zkflag_check.Initialize()
			zkflag_check.SetZKFlagConfiguration(param5,param2)
			check = zkflag_check
                default:
                        check = new(tcpCheck)
                        check.Initialize()
        }
	check.SetBaseConfiguration(IP,Host,Port,ConnectTimeout,ipv6)
        return check, nil
}
