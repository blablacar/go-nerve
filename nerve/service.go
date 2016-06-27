package nerve

import (
	"encoding/json"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"strconv"
	"sync"
	"time"
)

type Service struct {
	Name                       string
	Port                       int
	Host                       string
	PreferIpv4                 bool
	PreEnableCommand           []string
	PreEnableFailureIgnore     bool
	PostDisableCommand         []string
	Checks                     []json.RawMessage
	Reporters                  []json.RawMessage
	ReportFailureReplayInMilli int
	HaproxyServerOptions       string
	Labels                     map[string]string

	nerve                      *Nerve
	disabled                   error
	currentStatus              *error
	typedCheckersWithStatus    map[Checker]*error
	typedReportersWithReported map[Reporter]bool
	fields                     data.Fields
}

func (s *Service) Init(n *Nerve) error {
	logs.WithField("data", s).Info("service loaded") // todo rewrite with conf only
	s.nerve = n

	if s.ReportFailureReplayInMilli == 0 {
		s.ReportFailureReplayInMilli = 1000
	}

	s.typedReportersWithReported = make(map[Reporter]bool)
	s.typedCheckersWithStatus = make(map[Checker]*error)

	s.fields = data.WithField("service", s.Host+":"+strconv.Itoa(s.Port))
	for _, data := range s.Checks {
		checker, err := CheckerFromJson(data, s)
		if err != nil {
			return errs.WithEF(err, s.fields, "Failed to load check")
		}
		logs.WithF(s.fields).WithFields(checker.GetFields()).Info("check loaded")
		s.typedCheckersWithStatus[checker] = nil
	}
	if len(s.typedCheckersWithStatus) == 0 {
		logs.WithF(s.fields).Warn("No check specified, adding tcp")
		checker := NewCheckTcp()
		checker.Type = "tcp"
		checker.Init(s)
		s.typedCheckersWithStatus[checker] = nil
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

func (s *Service) Start(stopper <-chan struct{}, stopWait *sync.WaitGroup) {
	logs.WithFields(s.fields).Info("Starting service check")
	stopWait.Add(1)
	defer stopWait.Done()
	checkStopWait := &sync.WaitGroup{}

	statusChange := make(chan Check)
	for checker := range s.typedCheckersWithStatus {
		go checker.Run(statusChange, stopper, checkStopWait)
	}

	for {
		select {
		case status := <-statusChange:
			s.processStatus(status)
		case <-stopper:
			logs.WithFields(s.fields).Debug("Stop requested")
			checkStopWait.Wait()
			close(statusChange)
			for reporter := range s.typedReportersWithReported {
				reporter.Destroy()
			}
			return
		case <-time.After(time.Duration(s.ReportFailureReplayInMilli) * time.Millisecond):
			s.report(false)
		}
	}
}

func (s *Service) processStatus(check Check) {
	s.typedCheckersWithStatus[check.Checker] = &check.Status

	var combinedStatus error = nil
	for _, status := range s.typedCheckersWithStatus {
		if status != nil {
			combinedStatus = *status
		}
	}

	reportRequired := false
	if s.currentStatus == nil ||
		(*s.currentStatus == nil && combinedStatus == nil) ||
		(s.currentStatus != nil && combinedStatus != nil) {
		s.currentStatus = &combinedStatus
		reportRequired = true
	}

	if reportRequired {
		if *s.currentStatus == nil {
			logs.WithF(s.fields).Info("Service is available")
		} else {
			logs.WithEF(*s.currentStatus, s.fields).Warn("Service is not available")
		}
		s.report(true)
	}
}

func (s *Service) report(required bool) {
	if s.currentStatus == nil {
		return // no status yet
	}
	status := *s.currentStatus
	if s.disabled != nil {
		status = s.disabled
	}
	report := toReport(status, s)
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

func (s *Service) Disable() {
	s.disabled = errs.With("Service is disabled")
	s.report(true)
}

func (s *Service) Enable() {
	s.disabled = nil
	for reporter := range s.typedReportersWithReported {
		s.typedReportersWithReported[reporter] = false
	}
}
