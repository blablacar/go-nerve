package reporters

import (
	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

const REPORTER_ZOOKEEPER_TYPE string = "ZOOKEEPER"

type ZookeeperReporter struct {
	Reporter
	ZKHosts []string
	ZKPath string
	ZKKey string
}


func(x *ZookeeperReporter) Initialize(IP string, Port int, Rise int, Fall int) error {
	x.ZKHosts = append(x.ZKHosts,"localhost:443")
	x.ZKPath = "/nerve"
	x.ZKKey = "reporter"
	x._type = REPORTER_ZOOKEEPER_TYPE
	x.SetBaseConfiguration(IP,Port,Rise,Fall)
	return nil
}

func(x ZookeeperReporter) Report(Status int) error {
	if x.CanReport(Status) {
		//Connect to ZooKeeper
		conn, _, err := zk.Connect(x.ZKHosts, time.Second)
		if err != nil {
			log.Warn("Unable to Connect to ZooKeeper (",err,")")
			return err
		}
		realPath := x.ZKPath + "/" + x.IP
		if Status == 0 {
			flags := int32(0)
			acl := zk.WorldACL(zk.PermAll)
			_, err := conn.Create(realPath, []byte(x.ZKKey), flags, acl)
			if err != nil {
				log.Warn("Unable to Create [",realPath,"] into ZooKeeper")
				conn.Close()
				return err
			}
		}else {
			err := conn.Delete(realPath, -1)
			if err != nil {
				log.Warn("Unable to Delete [",realPath,"] into ZooKeeper")
				conn.Close()
				return err
			}
		}
		conn.Close()
	}
	return nil
}

func(x ZookeeperReporter) GetType() string {
	return x._type
}
