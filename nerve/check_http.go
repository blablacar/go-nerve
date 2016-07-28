package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type CheckHttp struct {
	CheckCommon
	Path string

	url    string
	client http.Client
}

func NewCheckHttp() *CheckHttp {
	return &CheckHttp{
		Path: "/",
	}
}

func (x *CheckHttp) Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
	x.CommonRun(x, statusChange, stop, doneWait)
}

func (x *CheckHttp) Init(s *Service) error {
	if err := x.CheckCommon.CommonInit(s); err != nil {
		return err
	}

	x.client = http.Client{
		Timeout: time.Duration(x.TimeoutInMilli) * time.Millisecond,
	}
	if len(x.Path) == 0 || x.Path[0] != '/' {
		x.Path = "/" + x.Path
	}

	x.url = "http://" + x.Host + ":" + strconv.Itoa(x.Port) + x.Path
	x.fields = x.fields.WithField("url", x.url).WithField("type", x.Type)
	return nil
}

func (x *CheckHttp) Check() error {
	resp, err := x.client.Get(x.url)
	if err != nil || (resp.StatusCode >= 500 && resp.StatusCode < 600) {
		ff := x.fields
		if err == nil {
			ff = ff.WithField("status_code", resp.StatusCode)
			if content, err := ioutil.ReadAll(resp.Body); err == nil {
				ff = ff.WithField("content", content)
			}
			resp.Body.Close()
		}
		return errs.WithEF(err, ff, "Url check failed")
	}
	if err == nil {
		resp.Body.Close()
	}
	return nil
}

func (x *CheckHttp) String() string {
	return x.url
}
