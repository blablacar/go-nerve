package nerve

import (
	"bufio"
	"bytes"
	"github.com/n0rad/go-erlog/errs"
	. "github.com/onsi/gomega"
	"testing"
)

func TestReportConsole(t *testing.T) {
	RegisterTestingT(t)
	var b bytes.Buffer
	reporter := NewReporterConsole()
	write := bufio.NewWriter(&b)
	reporter.writer = write

	reporter.Report(nil, &Service{})
	reporter.Report(errs.With("fail!"), &Service{})

	write.Flush()
	Expect(b.String()).Should(Equal("OK\nKO\n"))
}
