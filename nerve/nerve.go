package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"sync"
)

type Nerve struct {
	LogLevel string
	IPv6     bool
	Services []*Service

	stopChecker chan struct{}
	doneWaiter  sync.WaitGroup
}

func (n *Nerve) Init() error {
	n.stopChecker = make(chan struct{})

	for _, service := range n.Services {
		if err := service.Init(); err != nil {
			return errs.WithE(err, "Failed to init service")
		}
	}
	return nil
}

func (n *Nerve) Start() error {
	logs.Info("Starting nerve")
	for _, service := range n.Services {
		n.doneWaiter.Add(1)
		go service.Run(n.stopChecker, &n.doneWaiter)
	}
	return nil
}

func (n *Nerve) Stop() {
	logs.Info("Stopping nerve")
	close(n.stopChecker)
	n.doneWaiter.Wait()
	logs.Debug("All services stopped")
}
