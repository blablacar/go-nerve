package nerve

import (
	"net"
	"sync"

	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/prometheus/client_golang/prometheus"
)

type Nerve struct {
	LogLevel            *logs.Level
	ApiHost             string
	ApiPort             int
	Services            []*Service
	TemplatedConfigPath string

	nerveVersion         string
	nerveBuildTime       string
	checkerFailureCount  *prometheus.CounterVec
	reporterFailureCount *prometheus.CounterVec
	execFailureCount     *prometheus.CounterVec
	availableGauge       *prometheus.GaugeVec
	apiListener          net.Listener
	fields               data.Fields
	serviceStopper       chan struct{}
	servicesStopWait     sync.WaitGroup
}

func (n *Nerve) getService(name string) (*Service, error) {
	for _, s := range n.Services {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, errs.WithF(n.fields.WithField("name", name), "No service found with this name")
}

func (n *Nerve) Init(version string, buildTime string, logLevelIsSet bool) error {
	n.nerveVersion = version
	n.nerveBuildTime = buildTime
	if n.ApiPort == 0 {
		n.ApiPort = 3454
	}
	if n.TemplatedConfigPath == "" {
		n.TemplatedConfigPath = "/tmp/nerve-Config-templated.yaml"
	}
	if !logLevelIsSet && n.LogLevel != nil {
		logs.SetLevel(*n.LogLevel)
	}

	n.checkerFailureCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nerve",
			Name:      "checker_failure_total",
			Help:      "Counter of failed check",
		}, []string{"name", "ip", "port", "type"})

	n.execFailureCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nerve",
			Name:      "exec_failure_total",
			Help:      "Counter of failed exec",
		}, []string{"name", "ip", "port", "type"})

	n.reporterFailureCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nerve",
			Name:      "reporter_failure_total",
			Help:      "Counter of report failure",
		}, []string{"name", "ip", "port", "type"})

	n.availableGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "nerve",
			Name:      "service_available",
			Help:      "service available status",
		}, []string{"name", "ip", "port"})

	if err := prometheus.Register(n.execFailureCount); err != nil {
		return errs.WithEF(err, n.fields, "Failed to register prometheus exec_failure_total")
	}
	if err := prometheus.Register(n.checkerFailureCount); err != nil {
		return errs.WithEF(err, n.fields, "Failed to register prometheus check_failure_total")
	}
	if err := prometheus.Register(n.reporterFailureCount); err != nil {
		return errs.WithEF(err, n.fields, "Failed to register prometheus reporterFailureCount")
	}
	if err := prometheus.Register(n.availableGauge); err != nil {
		return errs.WithEF(err, n.fields, "Failed to register prometheus service_available")
	}

	n.serviceStopper = make(chan struct{})
	for _, service := range n.Services {
		if err := service.Init(n); err != nil {
			return errs.WithEF(err, data.WithField("service", service.Name), "Failed to init service")
		}
	}

	m := make(map[string]*Service)
	for i, service := range n.Services {
		m[service.Name] = service
		if len(m) != i+1 {
			return errs.WithF(n.fields.WithField("name", service.Name), "Duplicate service name")
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
