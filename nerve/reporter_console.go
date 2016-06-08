package nerve

import (
	"fmt"
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

func (r *ReporterConsole) Report(status error, s *Service) error {
	fmt.Fprintf(r.writer, "%s\n", r.toJsonReport(status, s))
	return nil
}
