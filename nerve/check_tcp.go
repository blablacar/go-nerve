package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"net"
	"strconv"
	"sync"
	"time"
)

type CheckTcp struct {
	CheckCommon

	url string
}

func NewCheckTcp() *CheckTcp {
	return &CheckTcp{}
}

func (x *CheckTcp) Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
	x.CommonRun(x, statusChange, stop, doneWait)
}

func (x *CheckTcp) Init(s *Service) error {
	if err := x.CheckCommon.CommonInit(s); err != nil {
		return err
	}

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
