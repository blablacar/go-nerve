package nerve

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestReportConsole(t *testing.T) {
	RegisterTestingT(t)
	var b bytes.Buffer
	reporter := NewReporterConsole()
	write := bufio.NewWriter(&b)
	reporter.writer = write

	reporter.Report(reportAvailable)
	reporter.Report(reportNotAvailable)

	write.Flush()

	lines := strings.Split(b.String(), "\n")
	Expect(lines).To(HaveLen(3))
	r1, _ := NewReport([]byte(lines[0]))
	r2, _ := NewReport([]byte(lines[1]))

	Expect(*r1.Available).Should(BeTrue())
	Expect(*r2.Available).Should(BeFalse())
	Expect(lines[2]).Should(Equal(""))
}
