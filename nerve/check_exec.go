package nerve

import (
	"sync"

	"github.com/n0rad/go-erlog/errs"
)

type CheckExec struct {
	CheckCommon
	Command []string
}

func NewCheckExec() *CheckExec {
	return &CheckExec{}
}

func (x *CheckExec) Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
	x.CommonRun(x, statusChange, stop, doneWait)
}

func (x *CheckExec) Init(s *Service) error {
	if err := x.CheckCommon.CommonInit(s); err != nil {
		return err
	}

	if len(x.Command) == 0 {
		return errs.With("Exec command type require a command")
	}
	x.fields = x.fields.WithField("command", x.Command)
	return nil
}

func (x *CheckExec) Check() error {
	if err := ExecCommand(x.Command, x.TimeoutInMilli); err != nil {
		return errs.WithEF(err, x.fields, "Check command failed")
	}
	return nil
}
