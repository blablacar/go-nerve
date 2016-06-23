package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"net"
	"strconv"
	"time"
)

type CheckTcp struct {
	CheckCommon

	url string
}

func NewCheckTcp() *CheckTcp {
	return &CheckTcp{}
}

func (x *CheckTcp) Init(conf *Service) error {
	x.url = x.Host + ":" + strconv.Itoa(x.Port)
	x.fields = x.fields.WithField("url", x.url)
	return nil
}

func (x *CheckTcp) Check() error {
	conn, err := net.DialTimeout("tcp", x.url, time.Duration(x.TimeoutInMilli)*time.Millisecond)
	if err != nil {
		return errs.WithEF(err, x.fields, "Check failed")
	}
	conn.Close()
	return nil
}
