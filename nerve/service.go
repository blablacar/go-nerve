package nerve

import (
	"encoding/json"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Service struct {
	Name                       string
	Port                       int
	Host                       string
	PreferIpv4                 bool
	Weight                     uint8
	Checks                     []json.RawMessage
	Reporters                  []json.RawMessage
	ReporterServiceName        string
	ReportReplayInMilli        int
	HaproxyServerOptions       string
	SetServiceAsDownOnShutdown *bool
	Labels                     map[string]string
	ExcludeFromGlobalDisable   bool

	PreAvailableCommand            []string
	PreAvailableMaxDurationInMilli int

	EnableCheckStableCommand            []string
	EnableCheckStableMaxDurationInMilli int
	EnableCheckStableIntervalInMilli    int

	EnableWarmupIntervalInMilli    int
	EnableWarmupMaxDurationInMilli int

	DisableShutdownCommand               []string
	DisableShutdownMaxDurationInMilli    int
	DisableGracefullyDoneCommand         []string
	DisableGracefullyDoneIntervalInMilli int
	DisableMaxDurationInMilli            int
	DisableMinDurationInMilli            int
	NoMetrics                            bool

	nerve                      *Nerve
	forceEnable                bool
	disabled                   error
	runNotifyMutex             sync.Mutex
	warmupGiveUp               chan struct{}
	warmupMutex                sync.Mutex
	warmupGiveUpMutex          sync.Mutex
	currentWeightIndex         int32
	currentStatus              *error
	typedCheckersWithStatus    map[Checker]*error
	typedReportersWithReported map[Reporter]bool
	fields                     data.Fields
}

var weights = []float64{0, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233}

const postFullWeightMax = 10

func (s *Service) Init(n *Nerve) error {
	logs.WithField("data", s).Info("service loaded") // todo rewrite with conf only
	s.nerve = n

	if s.Host == "" {
		s.Host = "127.0.0.1"
	}
	if s.Name == "" {
		s.Name = s.Host + ":" + strconv.Itoa(s.Port)
	}
	if s.ReporterServiceName == "" {
		s.ReporterServiceName = s.Name
	}
	if s.SetServiceAsDownOnShutdown == nil {
		val := true
		s.SetServiceAsDownOnShutdown = &val
	}

	if s.Weight == 0 {
		s.Weight = 255
	}
	if s.ReportReplayInMilli == 0 {
		s.ReportReplayInMilli = 1000
	}
	if s.EnableWarmupIntervalInMilli == 0 {
		s.EnableWarmupIntervalInMilli = 2000
	}
	if s.EnableWarmupMaxDurationInMilli == 0 {
		s.EnableWarmupMaxDurationInMilli = s.EnableWarmupIntervalInMilli * (postFullWeightMax + len(weights) + 7)
	}

	if s.EnableCheckStableMaxDurationInMilli == 0 {
		s.EnableCheckStableMaxDurationInMilli = 2 * 1000
	}

	if s.EnableCheckStableIntervalInMilli == 0 {
		s.EnableCheckStableIntervalInMilli = 1000
	}

	if s.PreAvailableMaxDurationInMilli == 0 {
		s.PreAvailableMaxDurationInMilli = 1000
	}

	if s.DisableShutdownMaxDurationInMilli == 0 {
		s.DisableShutdownMaxDurationInMilli = 30000
	}

	if s.DisableGracefullyDoneIntervalInMilli == 0 {
		s.DisableGracefullyDoneIntervalInMilli = 1000
	}
	if s.DisableMinDurationInMilli == 0 {
		s.DisableMinDurationInMilli = 3000
	}
	if s.DisableMaxDurationInMilli == 0 {
		s.DisableMaxDurationInMilli = 60 * 1000
	}

	s.fields = data.WithField("service", s.Host+":"+strconv.Itoa(s.Port))

	minDuration := s.EnableWarmupIntervalInMilli * (postFullWeightMax + len(weights))
	if s.EnableWarmupMaxDurationInMilli < minDuration {
		return errs.WithF(s.fields.WithField("minValue", minDuration),
			"EnableWarmupMaxDurationInMilli must be at least "+strconv.Itoa(postFullWeightMax+len(weights))+" times EnableWarmupIntervalInMilli")
	}

	s.typedReportersWithReported = make(map[Reporter]bool)
	s.typedCheckersWithStatus = make(map[Checker]*error)

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

	statusChange := make(chan Check, 2)
	for checker := range s.typedCheckersWithStatus {
		go checker.Run(statusChange, stopper, checkStopWait)
	}

	for {
		select {
		case status := <-statusChange:
			logs.WithF(s.fields.WithField("status", status)).Debug("New status received")
			s.processCheckResult(status)
		case <-stopper:
			//TODO since stop is the same everywhere, statusChange chan may stay stuck on shutdown
			logs.WithFields(s.fields).Debug("Stop requested")
			checkStopWait.Wait()
			close(statusChange)
			if *s.SetServiceAsDownOnShutdown {
				wait := &sync.WaitGroup{}
				wait.Add(1)
				s.Disable(wait, true)
				wait.Wait()
			}
			for reporter := range s.typedReportersWithReported {
				reporter.Destroy()
			}
			return
		case <-time.After(time.Duration(s.ReportReplayInMilli) * time.Millisecond):
			s.reportAndTellIfAtLeastOneReported(false)
		}
	}
}

func (s *Service) processCheckResult(check Check) {
	s.typedCheckersWithStatus[check.Checker] = &check.Status
	var combinedStatus error
	for _, status := range s.typedCheckersWithStatus {
		if status == nil {
			logs.WithF(s.fields).Debug("One check have no value, cannot report yet")
			return
		}
		if combinedStatus == nil {
			combinedStatus = *status
		}
	}

	if logs.IsDebugEnabled() {
		logs.WithF(s.fields.WithField("status", check).WithField("combined", combinedStatus)).Debug("combined status process")
	}

	if s.currentStatus == nil ||
		(*s.currentStatus == nil && combinedStatus != nil) ||
		(*s.currentStatus != nil && combinedStatus == nil) {
		s.currentStatus = &combinedStatus
		s.runNotify()
	} else {
		logs.WithF(s.fields).Debug("Combined status is same as previous, no report required")
	}
}

func (s *Service) runNotify() {
	s.runNotifyMutex.Lock()
	defer s.runNotifyMutex.Unlock()

	s.giveUpWarmup()
	if s.currentStatus == nil {
		logs.WithF(s.fields).Info("No status to notify")
		return
	}

	if (*s.currentStatus == nil && s.disabled == nil) || s.forceEnable {
		logs.WithF(s.fields).Info("Service is available")

		if len(s.PreAvailableCommand) > 0 {
			if err := ExecCommand(s.PreAvailableCommand, s.PreAvailableMaxDurationInMilli); err != nil {
				s.nerve.execFailureCount.WithLabelValues(s.Name, "pre-available", s.Host, strconv.Itoa(s.Port)).Inc()
				logs.WithEF(err, s.fields).Warn("Pre available command failed")
			}
		}

		s.warmup()
	} else {
		if !s.NoMetrics {
			s.nerve.availableGauge.WithLabelValues(s.Name, s.Host, strconv.Itoa(s.Port)).Set(0)
		}
		atomic.StoreInt32(&s.currentWeightIndex, 0)
		logs.WithEF(*s.currentStatus, s.fields).Warn("Service is not available")
		s.reportAndTellIfAtLeastOneReported(true)
	}
}

func (s *Service) giveUpWarmup() {
	s.warmupGiveUpMutex.Lock()
	defer s.warmupGiveUpMutex.Unlock()

	if s.warmupGiveUp != nil {
		close(s.warmupGiveUp)
		s.warmupGiveUp = nil
	}
}

func (s *Service) warmup() {
	s.warmupMutex.Lock()
	defer s.warmupMutex.Unlock()

	s.giveUpWarmup()
	s.warmupGiveUp = make(chan struct{})
	go s.Warmup(s.warmupGiveUp)
}

func (s *Service) Warmup(giveUp <-chan struct{}) {
	start := time.Now()
	atomic.StoreInt32(&s.currentWeightIndex, 0)

	done := make(chan struct{})
	if len(s.EnableCheckStableCommand) > 0 {
		go func() {
			for {
				if err := ExecCommand(s.EnableCheckStableCommand, s.EnableCheckStableMaxDurationInMilli); err != nil {
					s.nerve.execFailureCount.WithLabelValues(s.Name, s.Host, strconv.Itoa(s.Port), "check-stable").Inc()
					logs.WithEF(err, s.fields).Warn("Check stable command failed. Reset weight")
					atomic.StoreInt32(&s.currentWeightIndex, 0)
					s.reportAndTellIfAtLeastOneReported(true)
				}

				select {
				case <-giveUp:
					return
				case <-done:
					return
				case <-time.After(time.Duration(s.EnableCheckStableIntervalInMilli) * time.Millisecond):
				}
			}
		}()
	}

	for {
		if atomic.LoadInt32(&s.currentWeightIndex) < int32(len(weights)) && !s.reportAndTellIfAtLeastOneReported(true) {
			logs.WithF(s.fields).Debug("No report succeed. Reset weight")
			atomic.StoreInt32(&s.currentWeightIndex, 0)
		}

		atomic.AddInt32(&s.currentWeightIndex, 1)

		if atomic.LoadInt32(&s.currentWeightIndex) > int32(postFullWeightMax+len(weights)) {
			logs.WithF(s.fields).Debug("Service is fully stable")
			done <- struct{}{}
			s.warmupMutex.Lock()
			s.warmupGiveUp = nil
			s.warmupMutex.Unlock()
			return
		}

		if time.Now().After(start.Add(time.Duration(s.EnableWarmupMaxDurationInMilli) * time.Millisecond)) {
			logs.WithF(s.fields).Warn("Warmup reach max duration. set Full Weight")
			done <- struct{}{}
			s.warmupMutex.Lock()
			s.warmupGiveUp = nil
			s.warmupMutex.Unlock()
			atomic.StoreInt32(&s.currentWeightIndex, int32(len(weights)-1))
			s.reportAndTellIfAtLeastOneReported(true)
			return
		}

		select {
		case <-giveUp:
			logs.WithF(s.fields).Debug("Warmup giveup requested")
			return
		case <-time.After(time.Duration(s.EnableWarmupIntervalInMilli) * time.Millisecond):
		}
	}

}

func (s *Service) reportAndTellIfAtLeastOneReported(required bool) bool {
	if !s.NoMetrics {
		s.nerve.availableGauge.WithLabelValues(s.Name, s.Host, strconv.Itoa(s.Port)).Set(float64(s.CurrentWeight()))
	}
	if s.currentStatus == nil {
		return false // no status yet
	}
	status := *s.currentStatus
	if s.forceEnable {
		var e error
		status = e
	} else if s.disabled != nil {
		status = s.disabled
	}
	report := toReport(status, s)
	globalReported := 0
	for reporter, reported := range s.typedReportersWithReported {
		if required || !reported {
			logs.WithFields(s.fields).WithField("reporter", reporter).WithField("report", report).Debug("Sending report")
			if err := reporter.Report(report); err != nil {
				if reported == true {
					logs.WithEF(err, s.fields.WithFields(reporter.GetFields())).Error("Failed to report")
				}
				if !s.NoMetrics {
					s.nerve.reporterFailureCount.WithLabelValues(s.Name, s.Host, strconv.Itoa(s.Port), reporter.getCommon().Type).Inc()
				}
				s.typedReportersWithReported[reporter] = false
			} else {
				if reported == false {
					logs.WithF(s.fields).Info("Reported with success")
				}
				s.typedReportersWithReported[reporter] = true
				globalReported++
			}
		}
	}
	return globalReported > 0
}

func (s *Service) CurrentWeight() uint8 {
	if (!s.forceEnable && (s.currentStatus == nil || *s.currentStatus != nil)) || s.disabled != nil {
		return 0
	}

	index := atomic.LoadInt32(&s.currentWeightIndex)
	if index > int32(len(weights)-1) {
		index = int32(len(weights) - 1)
	}
	res := uint8(math.Ceil(weights[index] * float64(s.Weight) / weights[len(weights)-1]))
	if res == 0 {
		res++
	}
	return res
}

func (s *Service) Disable(doneWaiter *sync.WaitGroup, shutdown bool) {
	start := time.Now()
	logs.WithF(s.fields).Info("Disabling service")
	defer doneWaiter.Done()

	s.forceEnable = false
	s.disabled = errs.With("Service is disabled")
	s.runNotify()

	if len(s.DisableShutdownCommand) > 0 && shutdown {
		logs.WithF(s.fields).Debug("Run disableShutdown command")
		if err := ExecCommand(s.DisableShutdownCommand, s.DisableShutdownMaxDurationInMilli); err != nil {
			logs.WithEF(err, s.fields).Error("Shutdown result")
			s.nerve.execFailureCount.WithLabelValues(s.Name, s.Host, strconv.Itoa(s.Port), "disable-shutdown").Inc()
		}
	}

	if len(s.DisableGracefullyDoneCommand) > 0 {
		for {
			var err error
			if err = ExecCommand(s.DisableGracefullyDoneCommand, s.DisableGracefullyDoneIntervalInMilli); err == nil {
				logs.WithF(s.fields).Debug("Gracefull check succeed")
				break
			}

			s.nerve.execFailureCount.WithLabelValues(s.Name, s.Host, strconv.Itoa(s.Port), "disable-grace").Inc()
			logs.WithEF(err, s.fields).Debug("Gracefull check command fail")

			select {
			case <-time.After(start.Add(time.Duration(s.DisableMaxDurationInMilli) * time.Millisecond).Sub(time.Now())):
				logs.WithEF(err, s.fields).Warn("Disable max duration reached")
				return
			case <-time.After(time.Duration(s.DisableGracefullyDoneIntervalInMilli) * time.Millisecond):
			}
		}
	}

	time.Sleep(start.Add(time.Duration(s.DisableMinDurationInMilli) * time.Millisecond).Sub(time.Now()))
}

func (s *Service) Enable(force bool) {
	logs.WithF(s.fields.WithField("force", force)).Info("Enabling service")
	s.forceEnable = force
	s.disabled = nil
	s.runNotify()
}
