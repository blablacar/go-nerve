package checks

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

const CHECK_HTTPPROXY_TYPE = "HTTPPROXY"

type httpproxyCheck struct {
	Check
	ProxyUsername string
	ProxyPassword string
	ProxyHost string
	ProxyPort string
	URLs []string
	client http.Client
}

//Initialize 
func(x *httpproxyCheck) Initialize() error {
	//Default value pushed here
	x.Status = StatusUnknown
	x.Host = "localhost"
	x.Port = 80
	x.IP = "127.0.0.1"
	x.ConnectTimeout = 100
	x.DisconnectTimeout = 100
	x._type = CHECK_HTTPPROXY_TYPE
	timeout := time.Duration(x.ConnectTimeout) * time.Millisecond
	x.client = http.Client{
		Timeout: timeout,
	}
	return nil
}

func(hc *httpproxyCheck) SetHTTPProxyConfiguration(User string, Password string, Host string, Port string, URLs []string) {
	has_proxy := false
	if User != "" {
		hc.ProxyUsername = User
	}
	if Password != "" {
		hc.ProxyPassword = Password
	}
	if Host != "" {
		has_proxy = true
		hc.ProxyHost = Host
	}
	if Port != "" {
		has_proxy = true
		hc.ProxyPort = Port
	}
	if len(URLs) > 0 {
		hc.URLs = URLs
	}
	if has_proxy {
		proxyUrl, err := url.Parse("http://"+hc.ProxyHost+":"+hc.ProxyPort)
		if err == nil {
			timeout := time.Duration(hc.ConnectTimeout) * time.Millisecond
			hc.client = http.Client{
				Transport: &http.Transport {
					Proxy: http.ProxyURL(proxyUrl),
				},
			        Timeout: timeout,
			}
		}
	}
}

//Verify that the given proxy is healthy
func(hc *httpproxyCheck) DoCheck() (status int, err error) {
	for _, url := range hc.URLs {
		resp, err := hc.client.Get(url)
		if err != nil {
			log.WithError(err).Warn("HTTP Check of [",url,"] fail")
			return StatusKO, err
		}
		resp.Body.Close()
	}
	return StatusOK, nil
}

func(hc *httpproxyCheck) GetType() string {
	return hc._type
}

func (hc *httpproxyCheck) SetBaseConfiguration(IP string, Host string, Port int, ConnectTimeout int, ipv6 bool) {
	hc.IP = IP
	hc.Host = Host
	hc.Port = Port
	hc.ConnectTimeout = ConnectTimeout
	hc.IPv6 = ipv6
}
