package nerve

import (
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"net"
	"sync"
)

type Nerve struct {
	LogLevel           string
	IPv6               bool
	ApiUrl             string
	Services           []*Service
	DisableWaitInMilli *int

	apiListener net.Listener
	//apiServer   macaron.Macaron
	fields      data.Fields
	stopChecker chan struct{}
	doneWaiter  sync.WaitGroup
}

func (n *Nerve) Init() error {
	if n.ApiUrl == "" {
		n.ApiUrl = ":3454"
	}
	if n.DisableWaitInMilli == nil {
		val := 3000
		n.DisableWaitInMilli = &val
	}

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
	n.startApi()
	return nil
}

func (n *Nerve) Stop() {
	logs.Info("Stopping nerve")
	close(n.stopChecker)
	n.stopApi()
	n.doneWaiter.Wait()
	logs.Debug("All services stopped")
}
