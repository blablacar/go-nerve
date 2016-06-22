package nerve

import (
	"github.com/blablacar/go-nerve/nerve/test"
	"github.com/n0rad/go-erlog/logs"
	. "github.com/onsi/gomega"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"strconv"
	"testing"
	"time"
)

var ts *tests.TestCluster

func TestMain(m *testing.M) {
	os.Exit(wrap(m))
}

func wrap(m *testing.M) int {
	logs.SetLevel(logs.TRACE)
	var err error
	ts, err = tests.StartTestCluster(1, nil, os.Stderr)
	if err != nil {
		logs.WithE(err).Fatal("Failed to start test")
	}
	defer ts.Stop()
	return m.Run()
}

func TestZkReport(t *testing.T) {
	RegisterTestingT(t)

	r := NewReporterZookeeper()
	r.Hosts = []string{"127.0.0.1:" + strconv.Itoa(ts.Servers[0].Port)}
	r.Path = "/test"
	r.Init(&Service{})

	Expect(r.Report(Report{Available: true})).To(BeNil())
	checkThere(r)
}

func TestZkRefresh(t *testing.T) {
	RegisterTestingT(t)

	r := NewReporterZookeeper()
	r.Hosts = []string{"127.0.0.1:" + strconv.Itoa(ts.Servers[0].Port)}
	r.RefreshIntervalInMilli = 2000
	r.Path = "/test"
	r.Init(&Service{})

	Expect(r.Report(Report{Available: true})).To(BeNil())
	checkThere(r)
	remove(r)
	checkNotThere(r)
	time.Sleep(3 * time.Second)
	checkThere(r)
}

func TestZkRefreshClosed(t *testing.T) {
	RegisterTestingT(t)

	r := NewReporterZookeeper()
	r.Hosts = []string{"127.0.0.1:" + strconv.Itoa(ts.Servers[0].Port)}
	r.RefreshIntervalInMilli = 2000
	r.Path = "/test"
	r.Init(&Service{})

	r.Report(Report{Available: true})
	r.Destroy()
	time.Sleep(3 * time.Second)

	select {
	case <-r.stopChecker:
	default:
		t.Fail()
	}
}

func TestZkReportFailure(t *testing.T) {
	RegisterTestingT(t)

	r := NewReporterZookeeper()
	r.Hosts = []string{"127.0.0.1:" + strconv.Itoa(ts.Servers[0].Port)}
	r.Path = "/test"
	r.Init(&Service{})

	Expect(r.Report(Report{Available: true})).To(BeNil())
	Expect(r.Report(Report{Available: false})).To(BeNil())
	checkNotThere(r)
}

func TestZkReportAfterClose(t *testing.T) {
	RegisterTestingT(t)

	r := NewReporterZookeeper()
	r.Hosts = []string{"127.0.0.1:" + strconv.Itoa(ts.Servers[0].Port)}
	r.Path = "/test"
	r.Init(&Service{})

	Expect(r.Report(Report{Available: true})).To(BeNil())
	checkThere(r)
	r.connection.Close()
	checkNotThere(r)
	Expect(r.Report(Report{Available: true})).To(BeNil())
	checkThere(r)
}

func TestZkReportConnectionLost(t *testing.T) {
	RegisterTestingT(t)

	r := NewReporterZookeeper()
	r.Hosts = []string{"127.0.0.1:" + strconv.Itoa(ts.Servers[0].Port)}
	r.Path = "/test"
	r.Init(&Service{})

	Expect(r.Report(Report{Available: true})).To(BeNil())
	checkThere(r)
	r.connection = nil
	checkThere(r)
	Expect(r.Report(Report{Available: true})).To(BeNil())
	checkThere(r)
}

///////////////

func remove(r *ReporterZookeeper) {
	conn, _, _ := zk.Connect(r.Hosts, 10*time.Second)
	defer conn.Close()

	Expect(conn.Delete(r.currentNode, -1)).To(BeNil())
}

func checkThere(r *ReporterZookeeper) {
	conn, _, err := zk.Connect(r.Hosts, 10*time.Second)
	defer conn.Close()

	b, stats, err := conn.Get(r.currentNode)
	Expect(err).To(BeNil())
	Expect(b).ToNot(BeEmpty())
	Expect(stats).ToNot(BeNil())
}

func checkNotThere(r *ReporterZookeeper) {
	conn, _, err := zk.Connect(r.Hosts, 10*time.Second)
	defer conn.Close()

	_, _, err = conn.Get(r.currentNode)
	Expect(err).ToNot(BeNil())
}
