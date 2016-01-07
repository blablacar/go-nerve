package checks

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"time"
	"strconv"
)

const CHECK_HTTP_TYPE = "HTTP"

type httpCheck struct {
	Check
	URL string
	client http.Client
}

//Initialize 
func(x *httpCheck) Initialize() error {
	//Default value pushed here
	x.Status = StatusUnknown
	x.Host = "localhost"
	x.Port = 80
	x.IP = "127.0.0.1"
	x.URL = "http://localhost/"
	x.ConnectTimeout = 100
	x.DisconnectTimeout = 100
	x._type = CHECK_HTTP_TYPE
	timeout := time.Duration(x.ConnectTimeout) * time.Millisecond
	x.client = http.Client{
		Timeout: timeout,
	}
	return nil
}

//Verify that the given host or ip / port is healthy
func(x *httpCheck) DoCheck() (status int, err error) {
	resp, err := x.client.Get(x.URL)
	if err != nil {
		log.Warn("HTTP Check of [",x.URL,"] fail (",err,")")
		return StatusKO, err
	}
	resp.Body.Close()
	return StatusOK, nil
}

func(x *httpCheck) GetType() string {
	return x._type
}

func (x *httpCheck) SetBaseConfiguration(IP string, Host string, Port int, ConnectTimeout int, ipv6 bool) {
	x.IP = IP
	x.Host = Host
	x.Port = Port
	x.ConnectTimeout = ConnectTimeout
	x.IPv6 = ipv6
}

func (x *httpCheck) SetURI(uri string) {
	x.URL = "http://" + x.IP + ":" + strconv.Itoa(x.Port) + uri
}
