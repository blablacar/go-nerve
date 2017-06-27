package nerve

import (
	"strconv"
	"sync"
	"time"

	"github.com/n0rad/go-erlog/errs"
	"github.com/tevino/tcp-shaker"
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
	c := tcp.NewChecker(true)
	if err := c.InitChecker(); err != nil {
		return errs.WithEF(err, x.fields, "Checker init failed:")
	}
	err := c.CheckAddr(x.url, time.Duration(x.TimeoutInMilli)*time.Millisecond)
	switch err {
	case tcp.ErrTimeout:
		return errs.WithEF(err, x.fields, "Check failed (timeout)")
	case nil:
		return nil
	default:
		if e, ok := err.(*tcp.ErrConnect); ok {
			return errs.WithEF(e, x.fields, "Check failed (tcp connect failed)")
		} else {
			return errs.WithEF(err, x.fields, "Check failed (error while connecting)")
		}
	}
}
