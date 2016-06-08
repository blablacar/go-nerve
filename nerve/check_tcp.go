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
	x.url = IpLookupNoError(x.Host, conf.PreferIpv4).String() + ":" + strconv.Itoa(x.Port)
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
