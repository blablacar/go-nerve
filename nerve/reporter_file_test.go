package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"testing"
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

	reporter.Report(nil, &Service{})

	res, err := ioutil.ReadFile(reporter.Path)
	Expect(string(res), err).Should(Equal("OK\n"))
}

func TestReportReplace(t *testing.T) {
	RegisterTestingT(t)
	reporter := NewReporterFile()
	Expect(reporter.Init(s)).ToNot(HaveOccurred())

	reporter.Report(nil, &Service{})
	reporter.Report(errs.With("fail!"), &Service{})

	res, err := ioutil.ReadFile(reporter.Path)
	Expect(string(res), err).Should(Equal("KO\n"))
}

func TestReportAppend(t *testing.T) {
	RegisterTestingT(t)
	reporter := NewReporterFile()
	os.Remove(reporter.Path)
	reporter.Mode = Append
	Expect(reporter.Init(s)).ToNot(HaveOccurred())

	reporter.Report(nil, &Service{})
	reporter.Report(errs.With("fail!"), &Service{})

	res, err := ioutil.ReadFile(reporter.Path)
	Expect(string(res), err).Should(Equal("OK\nKO\n"))
}
