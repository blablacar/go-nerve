package checks

import (
	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

const CHECK_ZKFLAG_TYPE string = "ZKFLAG"

type zkflagCheck struct {
	Check
	Hosts []string
	Path string
}

//Initialize 
func (zc *zkflagCheck) Initialize() error {
	//Default value pushed here
	zc.Status = StatusUnknown
	zc.ConnectTimeout = 100
	zc.DisconnectTimeout = 100
	zc._type = CHECK_ZKFLAG_TYPE
	return nil
}

func(zc *zkflagCheck) SetZKFlagConfiguration(Hosts []string, Path string) {
	if len(Hosts) > 0 {
		zc.Hosts = Hosts
	}
	if Path != "" {
		zc.Path = Path
	}
}

func(zc *zkflagCheck) Connect() (*zk.Conn, error) {
	conn, _, err := zk.Connect(zc.Hosts, time.Second)
	if err != nil {
		log.Warn("Unable to Connect to ZooKeeper (",err,")")
		return nil, err
	}
	return conn, nil
}

//Verify that the given host or ip / port is healthy
func (zc *zkflagCheck) DoCheck() (int, error) {
	log.Debug("Check of ZooKeeper Maintenance starting")
	conn, err := zc.Connect()
	if err != nil {
		log.WithError(err).Warn("Unable to Connect to ZooKeeper")
		return StatusKO, err
	}
	exists, _, _ := conn.Exists(zc.Path)
	if exists {
		return StatusKO, nil
	}
	conn.Close()
	log.Debug("Check of ZooKeeper Maintenance stopping")
	return StatusOK, nil
}

func (zc *zkflagCheck) GetType() string {
	return zc._type
}

func (zc *zkflagCheck) SetBaseConfiguration(IP string, Host string, Port int, ConnectTimeout int, ipv6 bool) {
	zc.IP = IP
	zc.Host = Host
	zc.Port = Port
	zc.ConnectTimeout = ConnectTimeout
	zc.IPv6 = ipv6
}
