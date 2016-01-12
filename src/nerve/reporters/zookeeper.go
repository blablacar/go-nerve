package reporters

import (
	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"time"
	"strings"
)

const REPORTER_ZOOKEEPER_TYPE string = "ZOOKEEPER"

type ZookeeperReporter struct {
	Reporter
	ZKHosts []string
	ZKPath string
	ZKConnection *zk.Conn
	CurrentNode string
}


func(zr *ZookeeperReporter) Initialize(IP string, Port int, Rise int, Fall int, Weight int, ServiceName string, InstanceID string, HAProxyServerOptions string) error {
	zr._type = REPORTER_ZOOKEEPER_TYPE
	zr.SetBaseConfiguration(IP,Port,Rise,Fall,Weight,ServiceName,InstanceID,HAProxyServerOptions)
	zr.ZKConnection = nil
	zr.CurrentNode = zr.ZKPath + "/" + zr.IP + "_" + zr.InstanceID
	return nil
}

func(zr *ZookeeperReporter) SetZKConfiguration(ZKHosts []string, ZKPath string) {
	zr.ZKHosts = ZKHosts
	zr.ZKPath = ZKPath
}

func(zr *ZookeeperReporter) Connect() (zk.State, error) {
	if zr.ZKConnection != nil {
		state := zr.ZKConnection.State()
		switch state {
			case zk.StateUnknown,zk.StateConnectedReadOnly,zk.StateExpired,zk.StateAuthFailed,zk.StateConnecting: {
				//Disconnect, and let Reconnection happen
				log.Warn("Zookeeper Connection is in BAD State [",state,"] Reconnect")
				zr.ZKConnection.Close()
			}
			case zk.StateConnected, zk.StateHasSession: {
				log.Debug("Zookeeper Connection of [",zr.ServiceName,"][",zr.InstanceID,"] connected(",state,"), nothing to do.")
				return state, nil
			}
			case zk.StateDisconnected: {
				log.Info("Reporter Connection is Disconnected -> Reconnection")
			}
		}
	}
	conn, _, err := zk.Connect(zr.ZKHosts, time.Second)
	if err != nil {
		zr.ZKConnection = nil
		log.Warn("Unable to Connect to ZooKeeper (",err,")")
		return zk.StateDisconnected, err
	}
	zr.ZKConnection = conn
	state := zr.ZKConnection.State()
	return state, nil
}

func(zr *ZookeeperReporter) mkdirStaticPath(acl []zk.ACL) error {
	paths := strings.Split(zr.ZKPath, "/")
	full := ""
	for i, path := range paths {
		if i > 0 {
			full +=  "/"
		}
		full += path
		if exists, _, _ := zr.ZKConnection.Exists(full); exists {
			continue
		}

		log.WithField("full", full).Debug("Creating zk path")

		_, err := zr.ZKConnection.Create(full, []byte(""), int32(0), acl)
		if err != nil {
			log.WithError(err).WithField("path", full).Warn("Cannot create path")
			return err
		}
	}
	return nil
}

func(zr *ZookeeperReporter) Report(Status int) error {
	//Test Connection to ZooKeeper
	state, err := zr.Connect() //internally the connection is maintained
	if err != nil {
		log.Warn("Unable to Report... Connection to Zookeeper Fail")
		return err
	}
	if state == zk.StateHasSession && zr.CanReport(Status) {
		realPath := zr.ZKPath + "/" + zr.IP + "_" + zr.InstanceID
		exists, _, _ := zr.ZKConnection.Exists(zr.CurrentNode)
		if Status == 0 {
			if !exists {
				acl := zk.WorldACL(zk.PermAll)
				//Create Full Static Path if not exists
				err = zr.mkdirStaticPath(acl)
				if err != nil {
					log.WithError(err).Warn("Unable to Create Static Path into ZooKeeper")
					return err
				}
				//Don't use Create, as it's an ephemeral zk node, and there's some race condition to avoid
				zr.CurrentNode, err = zr.ZKConnection.CreateProtectedEphemeralSequential(realPath, []byte(zr.GetJsonReporterData()), acl)
				if err != nil {
					log.WithError(err).Warn("Unable to Create [",realPath,"] into ZooKeeper")
					return err
				}
			} else {
				zr.ZKConnection.Set(zr.CurrentNode, []byte(zr.GetJsonReporterData()), int32(0)	)
			}
		}else {
			if exists {
				err = zr.ZKConnection.Delete(zr.CurrentNode, -1)
				if err != nil {
					log.Warn("Unable to Delete [",realPath,"] into ZooKeeper")
					return err
				}
			}
		}
	}
	return nil
}

func(zr * ZookeeperReporter) Destroy() error {
	if zr.ZKConnection != nil {
		zr.ZKConnection.Close()
	}
	return nil
}

func(zr *ZookeeperReporter) GetType() string {
	return zr._type
}
