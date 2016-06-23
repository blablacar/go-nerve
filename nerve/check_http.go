package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"net/http"
	"strconv"
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

func (x *CheckHttp) Init(conf *Service) error {
	x.client = http.Client{
		Timeout: time.Duration(x.TimeoutInMilli) * time.Millisecond,
	}
	if len(x.Path) == 0 || x.Path[0] != '/' {
		x.Path = "/" + x.Path
	}

	x.url = "http://" + x.Host + ":" + strconv.Itoa(x.Port) + x.Path
	x.fields = x.fields.WithField("url", x.url).WithField("timeout", x.TimeoutInMilli).WithField("type", x.Type)
	return nil
}

func (x *CheckHttp) Check() error {
	resp, err := x.client.Get(x.url)
	if err != nil {
		return errs.WithEF(err, x.fields, "Check failed")
	}
	resp.Body.Close()
	return nil
}

func (x *CheckHttp) String() string {
	return x.url
}
