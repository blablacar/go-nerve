package checks

import (
	log "github.com/Sirupsen/logrus"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

const CHECK_MYSQL_TYPE string = "MYSQL"

type mysqlCheck struct {
	Check
	Username string
	Password string
	SQLRequest string
	Protocol string
}

//Initialize 
func (mc *mysqlCheck) Initialize() error {
	//Default value pushed here
	mc.Status = StatusUnknown
	mc.Protocol = "tcp"
	mc.Username = "nerve"
	mc.Password = "nerve"
	mc.SQLRequest = "select 1 where 1"
	mc.Host = "localhost"
	mc.Port = 3307
	mc.IP = "127.0.0.1"
	mc.ConnectTimeout = 100
	mc.DisconnectTimeout = 100
	mc._type = CHECK_MYSQL_TYPE
	return nil
}

//Set Specific Configuration
func(mc *mysqlCheck) SetMysqlConfiguration(Username string, Password string, SQLRequest string) {
	if Username != "" {
		mc.Username = Username
	}
	if Password != "" {
		mc.Password = Password
	}
	if SQLRequest != "" {
		mc.SQLRequest = SQLRequest
	}
}

//Verify that the given host or ip / port is healthy
func (mc *mysqlCheck) DoCheck() (int, error) {
	log.Debug("MySQL Check of [",mc.IP,":",mc.Port,"] starting")
	conn, err := sql.Open("mysql",mc.Username+":"+mc.Password+"@"+mc.Protocol+"(["+mc.IP+"]:"+strconv.Itoa(mc.Port)+")/?timeout="+strconv.Itoa(mc.ConnectTimeout)+"ms")
	if err != nil {
		log.WithError(err).Warn("Connection Check to [",mc.IP,":",mc.Port,"] failed")
		return StatusKO, err
	}
	_, err = conn.Query(mc.SQLRequest)
	if err != nil {
		log.WithError(err).Warn("Query Check to [",mc.IP,":",mc.Port,"] failed")
		return StatusKO, err
	}
	conn.Close()
	log.Debug("MySQL Check of [",mc.IP,":",mc.Port,"] finished")
	return StatusOK, nil
}

func (mc *mysqlCheck) GetType() string {
	return mc._type
}

func (mc *mysqlCheck) SetBaseConfiguration(IP string, Host string, Port int, ConnectTimeout int, ipv6 bool) {
	mc.IP = IP
	mc.Host = Host
	mc.Port = Port
	mc.ConnectTimeout = ConnectTimeout
	mc.IPv6 = ipv6
}
