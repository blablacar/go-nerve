package nerve

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/n0rad/go-erlog/errs"
	"github.com/tevino/tcp-shaker"
)

type CheckTcp struct {
	CheckCommon

	url string `json:"url,omitempty"`
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

	x.url = x.Host + ":" + strconv.Itoa(*x.Port)
	x.fields = x.fields.WithField("url", x.url)
	return nil
}

func (x *CheckTcp) Check() error {
	checker := tcp.NewChecker()
	ctx, stopChecker := context.WithCancel(context.Background())
	defer stopChecker()
	go func() {
		if err := checker.CheckingLoop(ctx); err != nil {
			fmt.Println("checking loop stopped due to fatal error: ", err)
		}
	}()
	<-checker.WaitReady()

	err := checker.CheckAddr(x.url, time.Duration(*x.TimeoutInMilli)*time.Millisecond)
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
