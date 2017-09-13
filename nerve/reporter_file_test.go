package nerve

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

var s = &Service{}
var btrue = true
var bfalse = false
var reportAvailable = Report{
	Available: &btrue,
	Host:      "127.0.0.1",
	Port:      80,
	Name:      "reportAvailable",
}
var reportNotAvailable = Report{
	Available: &bfalse,
	Host:      "127.0.0.1",
	Port:      80,
	Name:      "reportNotAvailable",
}
var reportMissingOptions = Report{
	Available: &btrue,
}

func TestCannotInitOnWrongPath(t *testing.T) {
	RegisterTestingT(t)
	reporter := NewReporterFile()
	reporter.Path = "/dev/null/nerve.status"

	Expect(reporter.Init(s)).To(HaveOccurred())
}

func TestCanReport(t *testing.T) {
	RegisterTestingT(t)
	reporter := NewReporterFile()
	Expect(reporter.Init(s)).ToNot(HaveOccurred())

	reporter.Report(reportAvailable)

	res, _ := ioutil.ReadFile(reporter.Path)
	r, _ := NewReport(res)
	Expect(*r.Available).Should(BeTrue())
}

func TestShouldErrWhenMissingOptionsReport(t *testing.T) {
	RegisterTestingT(t)
	reporter := NewReporterFile()
	Expect(reporter.Init(s)).ToNot(HaveOccurred())

	reporter.Report(reportMissingOptions)

	res, _ := ioutil.ReadFile(reporter.Path)
	_, err := NewReport(res)
	Expect(err).ShouldNot(BeNil())
}

func TestReportReplace(t *testing.T) {
	RegisterTestingT(t)
	reporter := NewReporterFile()
	Expect(reporter.Init(s)).ToNot(HaveOccurred())

	reporter.Report(reportAvailable)
	res, _ := ioutil.ReadFile(reporter.Path)
	r, _ := NewReport(res)
	Expect(*r.Available).Should(BeTrue())

	reporter.Report(reportNotAvailable)
	res, _ = ioutil.ReadFile(reporter.Path)
	Expect(res).Should(HaveLen(0))
}

func TestReportAppend(t *testing.T) {
	RegisterTestingT(t)
	reporter := NewReporterFile()
	os.Remove(reporter.Path)
	reporter.Append = true
	Expect(reporter.Init(s)).ToNot(HaveOccurred())

	reporter.Report(reportAvailable)
	reporter.Report(reportNotAvailable)

	res, _ := ioutil.ReadFile(reporter.Path)
	lines := strings.Split(string(res), "\n")
	Expect(lines).To(HaveLen(3))
	r1, _ := NewReport([]byte(lines[0]))
	r2, _ := NewReport([]byte(lines[1]))

	Expect(*r1.Available).Should(BeTrue())
	Expect(*r2.Available).Should(BeFalse())
	Expect(lines[2]).Should(Equal(""))
}
