package nerve

import (
	_ "github.com/n0rad/go-erlog/register"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"github.com/n0rad/go-erlog/errs"
)

type ReporterMemory struct {
	ReporterCommon
	reported error
	status error
	count    int
}
func (r *ReporterMemory) Init(s *Service) error {
	return nil
}
func (r *ReporterMemory) Report(status error, s *Service) error {
	r.count++
	r.status = status
	return r.reported
}

type CheckMemory struct {
	CheckCommon
	status error
}
func (x *CheckMemory) Init(s *Service) error {
	return nil
}
func (x *CheckMemory) Check() error {
	return x.status
}

func TestReplayReportFailure(t *testing.T) {
	nerve := &Nerve{
		Services: []*Service{{
			Host:                 "127.0.0.1",
			Port:                 1234,
			CheckIntervalInMilli: 1000,
			Rise:                 2,
			Fall:                 2,
		}},
	}
	require.NoError(t, nerve.Init())

	reporter := &ReporterMemory{}
	nerve.Services[0].typedReportersWithReported[reporter] = true

	check := &CheckMemory{}
	nerve.Services[0].typedChecks = append(nerve.Services[0].typedChecks, &TypedCheck{typedCheck: check, checkType: "mem"})

	require.NoError(t, nerve.Start())
	defer nerve.Stop()

	// everything is ok, reported once
	require.Nil(t, reporter.reported)
	time.Sleep(3 * time.Second)
	require.Nil(t, reporter.reported)
	require.Equal(t, reporter.count, 1)

	time.Sleep(3 * time.Second)
	require.Equal(t, reporter.count, 1)
	require.Nil(t, reporter.status)

	// now service fail and cannot report
	check.status = errs.With("Check failed")
	reporter.reported = errs.With("Cannot report")
	reporter.count = 0

	time.Sleep(3 * time.Second) // fall + 1
	require.True(t, reporter.count >= 2)

	// reporter is back
	reporter.reported = nil
	reporter.count = 0
	time.Sleep(1 * time.Second) // check interval

	require.Equal(t, reporter.count, 1)
	require.NotNil(t, reporter.status)
	time.Sleep(1 * time.Second)
	require.Equal(t, reporter.count, 1)

	// service is back
	check.status = nil
	time.Sleep(4 * time.Second) // rise + 2
	require.True(t, reporter.count == 2)
	require.Nil(t, reporter.status)
}
