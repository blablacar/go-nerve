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
	Connection *zk.Conn
}

type ZKDebugLogger struct {}

func(ZKDebugLogger) Printf(format string, a ...interface{}) {
	log.Debug(format, a)
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

func(zc *zkflagCheck) Connect() (zk.State, error) {
	if zc.Connection != nil {
		state := zc.Connection.State()
		switch state {
			case zk.StateUnknown,zk.StateConnectedReadOnly,zk.StateExpired,zk.StateAuthFailed,zk.StateConnecting: {
				//Disconnect, and let Reconnection happen
				log.Warn("ZKFlag Connection is in BAD State [",state,"] Reconnect")
				zc.Connection.Close()
			}
			case zk.StateConnected, zk.StateHasSession: {
				log.Debug("ZKFlag Connection established(",state,"), nothing to do.")
				return state, nil
			}
			case zk.StateDisconnected: {
				log.Info("ZKFlag Connection is Disconnected -> Reconnection")
			}
		}
	}
	conn, _, err := zk.Connect(zc.Hosts, 10 * time.Second)
	if err != nil {
		zc.Connection = nil
		log.Warn("Unable to Connect to ZKFlag (",err,")")
		return zk.StateDisconnected, err
	}
	zc.Connection = conn
	var zkLogger ZKDebugLogger
	zc.Connection.SetLogger(zkLogger)
	state := zc.Connection.State()
	return state, nil
}

//Verify that the given host or ip / port is healthy
func (zc *zkflagCheck) DoCheck() (int, error) {
	log.Debug("Check of ZooKeeper Flag starting")
	state, err := zc.Connect()
	if err != nil {
		log.WithError(err).Warn("Unable to Connect to ZKFlag")
		return StatusKO, err
	}
	if state == zk.StateHasSession {
		exists, _, _ := zc.Connection.Exists(zc.Path)
		if exists {
			log.Debug("ZooKeeper Flag detected")
			return StatusKO, nil
		}
	}else {
		log.Warn("ZKFlag Session in BAD State [",state,"]")
	}
	log.Debug("Check of ZooKeeper Flag stopping")
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
