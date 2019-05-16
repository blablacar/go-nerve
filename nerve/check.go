package nerve

import (
	"encoding/json"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

type Check struct {
	Checker Checker
	Status  error
}

type Checker interface {
	Init(service *Service) error
	Check() error
	GetFields() data.Fields
	Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup)
}

type CheckCommon struct {
	Type                 string `yaml:"type,omitempty"`
	Host                 string `yaml:"host,omitempty"`
	Port                 *int   `yaml:"port,omitempty"`
	TimeoutInMilli       *int   `yaml:"timeoutInMilli,omitempty"`
	Rise                 *int   `yaml:"rise,omitempty"`
	Fall                 *int   `yaml:"fall,omitempty"`
	CheckIntervalInMilli *int   `yaml:"checkIntervalInMilli,omitempty"`

	fields         data.Fields
	stableStatus   *error
	latestStatuses []error
	service        *Service
}

func (c *CheckCommon) GetFields() data.Fields {
	return c.fields
}

func (c *CheckCommon) CommonInit(s *Service) error {
	c.service = s
	if c.TimeoutInMilli == nil || *c.TimeoutInMilli == 0 {
		i := 1000
		c.TimeoutInMilli = &i
	}
	if c.Rise == nil || *c.Rise == 0 {
		i := 3
		c.Rise = &i
	}
	if c.Fall == nil || *c.Fall == 0 {
		i := 2
		c.Fall = &i
	}
	if c.CheckIntervalInMilli == nil || *c.CheckIntervalInMilli == 0 {
		i := 1000
		c.CheckIntervalInMilli = &i
	}
	if c.Port == nil || *c.Port == 0 {
		c.Port = &s.Port
	}
	if c.Host == "" {
		c.Host = s.Host
	}

	if c.Host == "" {
		c.Host = "127.0.0.1"
	} else if net.ParseIP(s.Host) == nil {
		c.Host = IpLookupNoError(c.Host, s.PreferIpv4).String()
	}
	c.fields = data.WithField("type", c.Type).WithFields(s.fields)

	return nil
}

func (c *CheckCommon) saveStatus(status error) {
	var tmp []error
	tmp = append(tmp, status)
	tmp = append(tmp, c.latestStatuses...)
	if len(tmp) > max(*c.Rise, *c.Fall) {
		c.latestStatuses = tmp[:len(tmp)-1]
	} else {
		c.latestStatuses = tmp
	}

	if len(c.latestStatuses) == 1 ||
		c.latestStatuses[0] == nil && c.latestStatuses[1] != nil ||
		c.latestStatuses[0] != nil && c.latestStatuses[1] == nil {
		logs.WithEF(status, c.fields).Debug("Check changed")
	}

}

func (c *CheckCommon) CommonRun(checker Checker, statusChange chan<- Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
	logs.WithF(c.fields).Info("Starting check")
	doneWait.Add(1)
	defer doneWait.Done()

	for {
		status := checker.Check()
		if logs.IsTraceEnabled() {
			logs.WithEF(status, c.fields).Trace("Check done")
		}
		if status != nil {
			logs.WithEF(status, c.fields).Debug("Failed check")
		}
		if status != nil && !c.service.NoMetrics {
			c.service.nerve.checkerFailureCount.WithLabelValues(c.service.Name, c.Host, strconv.Itoa(*c.Port), c.Type).Inc()
		}
		c.saveStatus(status)

		current := c.stableStatus
		latest := c.latestStatuses
		if (latest[0] == nil && sameLastStatusCount(latest) >= *c.Rise && (current == nil || *current != nil)) ||
			(latest[0] != nil && sameLastStatusCount(latest) >= *c.Fall && (current == nil || *current == nil)) {
			c.stableStatus = &status
			statusChange <- Check{checker, *c.stableStatus}
		}

		select {
		case <-stop:
			logs.WithFields(c.fields).Debug("Stopping check")
			return
		case <-time.After(time.Duration(*c.CheckIntervalInMilli) * time.Millisecond):
		}
	}
}

func CheckerFromJson(data []byte, s *Service) (Checker, error) {
	t := &CheckCommon{}
	if err := json.Unmarshal([]byte(data), t); err != nil {
		return nil, errs.WithE(err, "Failed to unmarshall check type")
	}

	fields := s.fields.WithField("type", t.Type)
	var typedCheck Checker
	switch t.Type {
	case "http":
		typedCheck = NewCheckHttp()
	case "https":
		typedCheck = NewCheckHttps()
	case "proxyhttp":
		typedCheck = NewCheckProxyHttp()
	case "tcp":
		typedCheck = NewCheckTcp()
	case "sql":
		typedCheck = NewCheckSql()
	case "amqp":
		typedCheck = NewCheckAmqp()
	case "exec":
		typedCheck = NewCheckExec()
	case "kubernetes":
		typedCheck = NewCheckKubernetes()
	default:
		return nil, errs.WithF(fields, "Unsupported check type")
	}

	if err := json.Unmarshal([]byte(data), &typedCheck); err != nil {
		return nil, errs.WithEF(err, fields, "Failed to unmarshall check")
	}

	if err := typedCheck.Init(s); err != nil {
		return nil, errs.WithEF(err, fields, "Failed to init check")
	}
	return typedCheck, nil
}

func sameLastStatusCount(statuses []error) int {
	current := statuses[0]
	i := 1
	for ; i < len(statuses); i++ {
		if current == nil && statuses[i] != nil || current != nil && statuses[i] == nil {
			return i
		}
	}
	return i
}
