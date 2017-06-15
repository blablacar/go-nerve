package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"io/ioutil"
	"net/http"
	"strconv"
	"crypto/tls"
	"sync"
	"time"
)

type CheckHttps struct {
	CheckCommon
	Path string
	url string

	client http.Client
}

func NewCheckHttps() *CheckHttps {
	return &CheckHttps{
		Path: "/",
	}
}

func (x *CheckHttps) Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
	x.CommonRun(x, statusChange, stop, doneWait)
}

func (x *CheckHttps) Init(s *Service) error {
	if err := x.CheckCommon.CommonInit(s); err != nil {
		return err
	}

	tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }

	x.client = http.Client{
		Timeout: time.Duration(x.TimeoutInMilli) * time.Millisecond,
		Transport: tr,
	}
	if len(x.Path) == 0 || x.Path[0] != '/' {
		x.Path = "/" + x.Path
	}

	x.url = "https://" + x.Host + ":" + strconv.Itoa(x.Port) + x.Path
	x.fields = x.fields.WithField("url", x.url).WithField("type", x.Type)
	return nil
}

func (x *CheckHttps) Check() error {
	resp, err := x.client.Get(x.url)
	if err != nil {
		return errs.WithEF(err, x.fields, "Url check failed")
	}

	content, bodyErr := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		ff := x.fields.WithField("status_code", resp.StatusCode)
		if bodyErr == nil {
			ff = ff.WithField("content", string(content))
		}
		resp.Body.Close()
		return errs.WithEF(err, ff, "Url check failed")
	}
	return nil
}

func (x *CheckHttps) String() string {
	return x.url
}
