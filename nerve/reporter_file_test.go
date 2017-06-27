package nerve

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

var s = &Service{}

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

	reporter.Report(Report{Available: &btrue})

	res, _ := ioutil.ReadFile(reporter.Path)
	r, _ := NewReport(res)
	Expect(*r.Available).Should(BeTrue())
}

func TestReportReplace(t *testing.T) {
	RegisterTestingT(t)
	reporter := NewReporterFile()
	Expect(reporter.Init(s)).ToNot(HaveOccurred())

	reporter.Report(Report{Available: &btrue})
	res, _ := ioutil.ReadFile(reporter.Path)
	r, _ := NewReport(res)
	Expect(*r.Available).Should(BeTrue())

	reporter.Report(Report{Available: &bfalse})
	res, _ = ioutil.ReadFile(reporter.Path)
	Expect(res).Should(HaveLen(0))
}

func TestReportAppend(t *testing.T) {
	RegisterTestingT(t)
	reporter := NewReporterFile()
	os.Remove(reporter.Path)
	reporter.Append = true
	Expect(reporter.Init(s)).ToNot(HaveOccurred())

	reporter.Report(Report{Available: &btrue})
	reporter.Report(Report{Available: &bfalse})

	res, _ := ioutil.ReadFile(reporter.Path)
	lines := strings.Split(string(res), "\n")
	Expect(lines).To(HaveLen(3))
	r1, _ := NewReport([]byte(lines[0]))
	r2, _ := NewReport([]byte(lines[1]))

	Expect(*r1.Available).Should(BeTrue())
	Expect(*r2.Available).Should(BeFalse())
	Expect(lines[2]).Should(Equal(""))
}
