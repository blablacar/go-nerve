package nerve

import (
	"io/ioutil"
	"net/http"
        "crypto/tls"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

type CheckProxyHttp struct {
	CheckCommon
	ProxyHost            string
	ProxyPort            int
	ProxyUsername        string
	ProxyPassword        string
	Urls                 []string
	FailOnAnyUnreachable bool

	client http.Client
}

func (x *CheckProxyHttp) Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
	x.CommonRun(x, statusChange, stop, doneWait)
}

func NewCheckProxyHttp() *CheckProxyHttp {
	return &CheckProxyHttp{}
}

func (x *CheckProxyHttp) Init(s *Service) error {
	if err := x.CheckCommon.CommonInit(s); err != nil {
		return err
	}

	if x.ProxyHost == "" {
		x.ProxyHost = s.Host
	}
	if x.ProxyPort == 0 {
		x.ProxyPort = s.Port
	}

	proxyUrl, err := url.Parse("http://" + x.ProxyUsername + ":" + x.ProxyPassword + "@" + x.ProxyHost + ":" + strconv.Itoa(x.ProxyPort))
	if err != nil {
		return errs.WithEF(err, x.fields, "failed to prepare proxy url")
	}

	for i, url := range x.Urls {
		if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
			x.Urls[i] = "http://" + url
		}
	}

	x.client = http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
                        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Duration(x.TimeoutInMilli) * time.Millisecond,
	}

	return nil
}

func (x *CheckProxyHttp) Check() error {
	result := make(chan error)
	for _, url := range x.Urls {
		go func(url string) {
			var res error
			resp, err := x.client.Get(url)
			if err != nil || (resp.StatusCode >= 500 && resp.StatusCode < 600) {
				ff := x.fields.WithField("url", url)
				if err == nil {
					ff = ff.WithField("status_code", resp.StatusCode)
					if content, err := ioutil.ReadAll(resp.Body); err == nil {
						ff = ff.WithField("content", string(content))
					}
					resp.Body.Close()
				}
				res = errs.WithEF(err, ff, "Url check failed")
				logs.WithEF(err, x.fields).Trace("Url check failed")
			}
			if err == nil {
				resp.Body.Close()
			}
			result <- res
		}(url)
	}

	errors := errs.WithF(x.fields, "Url(s) unreachable")
	for i := 0; i < len(x.Urls); i++ {
		res := <-result
		if res != nil {
			errors.WithErr(res)
		}
	}

	if (x.FailOnAnyUnreachable && len(errors.Errs) > 0) || (!x.FailOnAnyUnreachable && len(errors.Errs) == len(x.Urls)) {
		logs.WithEF(errors, x.fields).Trace("Enough failed received")
		return errs.WithEF(errors, x.fields, "Enough failed received")
	}
	return nil
}
