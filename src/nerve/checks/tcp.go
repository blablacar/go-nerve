package checks

import (
	log "github.com/Sirupsen/logrus"
	"time"
	"net"
	"strconv"
)

const CHECK_TCP_TYPE string = "TCP"

type tcpCheck struct {
	Check
}

//Initialize 
func (x *tcpCheck) Initialize() error {
	//Default value pushed here
	x.Status = StatusUnknown
	x.Host = "localhost"
	x.Port = 80
	x.IP = "127.0.0.1"
	x.ConnectTimeout = 100
	x.DisconnectTimeout = 100
	x._type = CHECK_TCP_TYPE
	return nil
}

//Verify that the given host or ip / port is healthy
func (x *tcpCheck) DoCheck() (int, error) {
	log.Debug("Check of [",x.IP,":",x.Port,"] starting")
	conn, err := net.DialTimeout("tcp",x.IP + ":" + strconv.Itoa(x.Port),time.Duration(x.ConnectTimeout)*time.Millisecond)
	if err != nil {
		log.Warn("Check to [",x.IP,":",x.Port,"] failed(",err,")")
		return StatusKO, err
	}
	conn.Close()
	log.Debug("Check of [",x.IP,":",x.Port,"] finished")
	return StatusOK, nil
}

func (x *tcpCheck) GetType() string {
	return x._type
}

func (x *tcpCheck) SetBaseConfiguration(IP string, Host string, Port int, ConnectTimeout int, ipv6 bool) {
	x.IP = IP
	x.Host = Host
	x.Port = Port
	x.ConnectTimeout = ConnectTimeout
	x.IPv6 = ipv6
}
