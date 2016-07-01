package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CheckHttpProxy struct {
	CheckCommon
	ProxyHost            string
	ProxyPort            int
	ProxyUsername        string
	ProxyPassword        string
	Urls                 []string
	FailOnAnyUnreachable bool

	client http.Client
}

func (x *CheckHttpProxy) Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
	x.CommonRun(x, statusChange, stop, doneWait)
}

func NewCheckProxyHttp() *CheckHttpProxy {
	return &CheckHttpProxy{}
}

func (x *CheckHttpProxy) Init(s *Service) error {
	if err := x.CheckCommon.CommonInit(s); err != nil {
		return err
	}

	proxyUrl, err := url.Parse("http://" + x.ProxyUsername + ":" + x.ProxyPassword + "@" + x.ProxyHost + ":" + strconv.Itoa(x.ProxyPort))
	if err != nil {
		return errs.WithEF(err, x.fields, "failed to prepare proxy url")
	}

	for i, url := range x.Urls {
		if !strings.HasPrefix(url, "http://") {
			x.Urls[i] = "http://" + url
		}
	}

	x.client = http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
		Timeout: time.Duration(x.TimeoutInMilli) * time.Millisecond,
	}

	return nil
}

func (x *CheckHttpProxy) Check() error {
	result := make(chan error)
	for _, url := range x.Urls {
		go func(url string) {
			var res error
			resp, err := x.client.Get(url)
			if err != nil || (resp.StatusCode >= 500 && resp.StatusCode < 600) {
				res = errs.WithEF(err, x.fields.WithField("url", url), "Url check failed")
				logs.WithEF(err, x.fields).Trace("Url check failed")
			} else {
				resp.Body.Close()
			}
			result <- res
		}(url)
	}

	failCount := 0
	var oneErr error
	for i := 0; i < len(x.Urls); i++ {
		res := <-result
		if res != nil {
			failCount++
			oneErr = res
		}
	}

	if (x.FailOnAnyUnreachable && failCount > 0) ||
		(!x.FailOnAnyUnreachable && failCount == len(x.Urls)) {
		logs.WithEF(oneErr, x.fields.WithField("count", failCount)).Trace("Enough failed received")
		return errs.WithEF(oneErr, x.fields, "All urls are unreachable")
	}
	return nil
}
