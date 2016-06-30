package nerve

import (
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/prometheus/client_golang/prometheus"
	"net"
	"sync"
)

type Nerve struct {
	ApiHost            string
	ApiPort            int
	Services           []*Service
	DisableWaitInMilli *int

	nerveVersion     string
	nerveBuildTime   string
	failureCount     *prometheus.CounterVec
	apiListener      net.Listener
	fields           data.Fields
	serviceStopper   chan struct{}
	servicesStopWait sync.WaitGroup
}

func (n *Nerve) Init(version string, buildTime string) error {
	n.nerveVersion = version
	n.nerveBuildTime = buildTime
	if n.ApiPort == 0 {
		n.ApiPort = 3454
	}
	if n.DisableWaitInMilli == nil {
		val := 3000
		n.DisableWaitInMilli = &val
	}

	n.failureCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nerve",
			Name:      "check_failure_total",
			Help:      "Counter of failed check",
		}, []string{"check_type"})

	if err := prometheus.Register(n.failureCount); err != nil {
		return errs.WithEF(err, n.fields, "Failed to register failure count metrics")
	}

	n.serviceStopper = make(chan struct{})
	for _, service := range n.Services {
		if err := service.Init(n); err != nil {
			return errs.WithE(err, "Failed to init service")
		}
	}

	return nil
}

func (n *Nerve) Start(startStatus chan error) {
	logs.Info("Starting nerve")
	if len(n.Services) == 0 {
		if startStatus != nil {
			startStatus <- errs.WithF(n.fields, "No service specified")
		}
		return
	}

	for _, service := range n.Services {
		go service.Start(n.serviceStopper, &n.servicesStopWait)
	}

	res := n.startApi()
	if startStatus != nil {
		startStatus <- res
	}
}

func (n *Nerve) Stop() {
	logs.Info("Stopping nerve")
	close(n.serviceStopper)
	n.stopApi()
	n.servicesStopWait.Wait()
	logs.Debug("All services stopped")
}
