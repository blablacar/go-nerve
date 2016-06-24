package nerve

import (
	"encoding/json"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"sync"
	"time"
)

type TypedCheck struct {
	checkType  string
	typedCheck Checker
}

type Status bool

const OK Status = true
const KO Status = false

type Service struct {
	Name                 string
	Port                 int
	Host                 string
	PreferIpv4           bool
	Rise                 int
	Fall                 int
	CheckIntervalInMilli int
	Checks               []json.RawMessage
	Reporters            []json.RawMessage
	HaproxyServerOptions string
	Labels               map[string]string

	statusGauge                *prometheus.Counter
	failureCount               *prometheus.CounterVec
	disabled                   error
	typedChecks                []*TypedCheck
	typedReportersWithReported map[Reporter]bool
	fields                     data.Fields
	currentNotifStatus         *Status
	latestStatuses             []Status
}

func (s *Service) Init() error {
	logs.WithField("data", s).Info("service loaded")

	s.typedReportersWithReported = make(map[Reporter]bool)

	//TODO NewService()
	if s.CheckIntervalInMilli == 0 {
		s.CheckIntervalInMilli = 2000
	}
	if s.Fall == 0 {
		s.Fall = 2
	}
	if s.Rise == 0 {
		s.Rise = 3
	}

	//s.statusGauge = &prometheus.NewCounterVec(prometheus.CounterOpts{
	//	Namespace: "nerve",
	//	Name:      "service status",
	//	Help:      "service status",
	//})

	//s.statusGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	//	Namespace: "nerve",
	//	Name:      "service Status",
	//	Help:      "service Status",
	//})
	s.failureCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nerve",
			Name:      "check_failure_total",
			Help:      "Counter of failed check",
		}, []string{"check_type"})
	if err := prometheus.Register(s.failureCount); err != nil {
		return errs.WithEF(err, s.fields, "Failed to register failure count metrics")
	}

	s.fields = data.WithField("service", s.Host+":"+strconv.Itoa(s.Port))
	for _, data := range s.Checks {
		checkType, checker, err := CheckerFromJson(data, s)
		if err != nil {
			return errs.WithEF(err, s.fields, "Failed to load check")
		}
		logs.WithF(s.fields).WithFields(checker.GetFields()).Info("check loaded")
		s.typedChecks = append(s.typedChecks, &TypedCheck{typedCheck: checker, checkType: checkType})
	}
	if len(s.typedChecks) == 0 {
		logs.WithF(s.fields).Warn("No check specified, adding tcp")
		check := NewCheckTcp()
		check.CheckCommon.Init(s)
		check.Init(s)
		s.typedChecks = append(s.typedChecks, &TypedCheck{typedCheck: check, checkType: "tcp"})
	}

	for _, data := range s.Reporters {
		reporter, err := ReporterFromJson(data, s)
		if err != nil {
			return errs.WithEF(err, s.fields, "Failed to load reporter")
		}
		logs.WithF(s.fields).WithFields(reporter.GetFields()).Info("Reporter loaded")
		s.typedReportersWithReported[reporter] = true
	}
	if len(s.typedReportersWithReported) == 0 {
		logs.WithF(s.fields).Warn("No reporter specified, adding console")
		s.typedReportersWithReported[NewReporterConsole()] = true
	}

	return nil
}

func (s *Service) Run(stop <-chan struct{}, servicesGroup *sync.WaitGroup) {
	logs.WithFields(s.fields).Info("Starting service check")
	defer servicesGroup.Done()

	for {
		errStatus := s.checkServiceStatus()
		s.saveStatus(errStatus)
		required := s.processStatusAndTellIfReportRequired(errStatus)
		s.report(errStatus, required)
		select {
		case <-stop:
			logs.WithFields(s.fields).Debug("Stop requested")
			for reporter := range s.typedReportersWithReported {
				reporter.Destroy()
			}
			return
		default:
			time.Sleep(time.Duration(s.CheckIntervalInMilli) * time.Millisecond)
		}
	}
}

func (s *Service) report(errStatus error, required bool) {
	if s.disabled != nil {
		errStatus = s.disabled
	}
	report := toReport(errStatus, s)
	for reporter, reported := range s.typedReportersWithReported {
		if required || !reported {
			logs.WithFields(s.fields).WithField("reporter", reporter).WithField("report", report).Debug("Sending report")
			if err := reporter.Report(report); err != nil {
				logs.WithEF(err, s.fields.WithFields(reporter.GetFields())).Error("Failed to report")
				s.typedReportersWithReported[reporter] = false
			} else {
				s.typedReportersWithReported[reporter] = true
			}
		}
	}
}

func (s *Service) checkServiceStatus() error {
	var err error
	for _, check := range s.typedChecks {
		logs.WithFields(check.typedCheck.GetFields()).Debug("Running check")
		if err = check.typedCheck.Check(); err != nil {
			s.failureCount.WithLabelValues(check.checkType).Inc()
			logs.WithFields(check.typedCheck.GetFields()).Debug("KO")
			break
		}
		logs.WithFields(check.typedCheck.GetFields()).Debug("OK")
	}
	return err
}

func (s *Service) processStatusAndTellIfReportRequired(statusErr error) bool {
	current := s.currentNotifStatus
	latest := s.latestStatuses

	if (latest[0] == OK && sameLastStatusCount(latest) >= s.Rise && (current == nil || *current == KO)) ||
		(latest[0] == KO && sameLastStatusCount(latest) >= s.Fall && (current == nil || *current == OK)) {

		s.currentNotifStatus = &s.latestStatuses[0]
		if s.latestStatuses[0] == OK {
			logs.WithFields(s.fields).Info("Service is available")
		} else {
			logs.WithError(statusErr).WithFields(s.fields).Warn("Service is not available")
		}

		return true
	}
	return false
}

func sameLastStatusCount(statuses []Status) int {
	var current = statuses[0]
	i := 1
	for ; i < len(statuses); i++ {
		if statuses[i] != current {
			return i
		}
	}
	return i
}

func (s *Service) Disable() {
	s.disabled = errs.With("Service is disabled")
	s.report(s.disabled, true)
}

func (s *Service) Enable() {
	s.disabled = nil
	for reporter := range s.typedReportersWithReported {
		s.typedReportersWithReported[reporter] = false
	}
}

func (s *Service) saveStatus(err error) {
	status := KO
	if err == nil {
		status = OK
	}

	var tmp []Status
	tmp = append(tmp, status)
	tmp = append(tmp, s.latestStatuses...)
	if len(tmp) > max(s.Rise, s.Fall) {
		s.latestStatuses = tmp[:len(tmp)-1]
	} else {
		s.latestStatuses = tmp
	}
}
