package nerve

import (
	"fmt"
	"github.com/n0rad/go-erlog/errs"
	"io"
	"os"
)

type ReporterConsole struct {
	ReporterCommon
	writer io.Writer
}

func NewReporterConsole() *ReporterConsole {
	return &ReporterConsole{
		writer: os.Stdout,
	}
}

func (r *ReporterConsole) Init(s *Service) error {
	return nil
}

func (r *ReporterConsole) Report(report Report) error {
	content, err := report.toJson()
	if err != nil {
		return errs.WithEF(err, r.fields, "Failed to prepare report")
	}
	fmt.Fprintf(r.writer, "%s\n", content)
	return nil
}
