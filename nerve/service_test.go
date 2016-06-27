package nerve

//import (
//	"github.com/n0rad/go-erlog/errs"
//	_ "github.com/n0rad/go-erlog/register"
//	"github.com/stretchr/testify/require"
//	"testing"
//	"time"
//	"sync"
//)
//
//type ReporterMemory struct {
//	ReporterCommon
//	fail   error
//	report Report
//	count  int
//}
//
//func (r *ReporterMemory) Init(s *Service) error {
//	return nil
//}
//func (r *ReporterMemory) Report(report Report) error {
//	r.count++
//	r.report = report
//	return r.fail
//}
//
//type CheckMemory struct {
//	CheckCommon
//	status error
//}
//
//func (x *CheckMemory) Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
//	x.CommonRun(x, statusChange, stop, doneWait)
//}
//
//func (x *CheckMemory) Init(s *Service) error {
//	return nil
//}
//func (x *CheckMemory) Check() error {
//	return x.status
//}
//
//func TestReplayReportFailure(t *testing.T) {
//	nerve := &Nerve{
//		DisableWaitInMilli: &disableWait,
//		Services: []*Service{{
//			Port:                 1234,
//		}},
//	}
//	require.NoError(t, nerve.Init("", ""))
//
//	reporter := &ReporterMemory{}
//	nerve.Services[0].typedReportersWithReported[reporter] = true
//
//	check := &CheckMemory{}
//	nerve.Services[0].typedCheckersWithStatus = make(map[Checker]*error)
//	nerve.Services[0].typedCheckersWithStatus[check] = nil
//
//	nerve.Start(nil)
//	defer nerve.Stop()
//
//	// everything is ok, reported once
//	require.Nil(t, reporter.fail)
//	time.Sleep(30 * time.Millisecond)
//	require.Nil(t, reporter.fail)
//	require.Equal(t, reporter.count, 1)
//
//	time.Sleep(30 * time.Millisecond)
//	require.Equal(t, reporter.count, 1)
//	require.True(t, reporter.report.Available)
//
//	// now service fail and cannot report
//	check.status = errs.With("Check failed")
//	reporter.fail = errs.With("Cannot report")
//	reporter.count = 0
//
//	time.Sleep(30 * time.Millisecond) // fall + 1
//	require.True(t, reporter.count >= 2)
//
//	// reporter is back
//	reporter.fail = nil
//	reporter.count = 0
//	time.Sleep(10 * time.Millisecond) // check interval
//
//	require.Equal(t, reporter.count, 1)
//	require.False(t, reporter.report.Available)
//	time.Sleep(10 * time.Millisecond)
//	require.Equal(t, reporter.count, 1)
//
//	// service is back
//	check.status = nil
//	time.Sleep(40 * time.Millisecond) // rise + 2
//	require.True(t, reporter.count == 2)
//	require.True(t, reporter.report.Available)
//}
