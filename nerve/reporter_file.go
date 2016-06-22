package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"os"
	"path"
)

type FileMode bool

const Replace FileMode = false
const Append FileMode = true

type ReporterFile struct {
	ReporterCommon
	Path string
	Mode FileMode
}

func NewReporterFile() *ReporterFile {
	return &ReporterFile{
		Path: "/tmp/nerve.report",
		Mode: Replace,
	}
}

func (r *ReporterFile) Init(s *Service) error {
	if r.Path == "" {
		return errs.WithF(s.fields, "Reporter file need a path")
	}

	r.fields = r.fields.WithField("path", r.Path)
	if err := os.MkdirAll(path.Dir(r.Path), 0755); err != nil {
		return errs.WithEF(err, r.fields.WithField("path", path.Dir(r.Path)), "Failed to create directories")
	}
	file, err := r.openReport()
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func (r *ReporterFile) openReport() (*os.File, error) {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if r.Mode == Append {
		flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	}
	file, err := os.OpenFile(r.Path, flags, 0644)
	if err != nil {
		return nil, errs.WithEF(err, r.fields, "Failed to open report file")
	}
	return file, nil
}

func (r *ReporterFile) Report(report Report) error {
	file, err := r.openReport()
	if err != nil {
		return err
	}
	defer file.Close()

	var res string
	if r.Mode == Replace && !report.Available {
		res = ""
	} else {
		content, err := report.toJson()
		if err != nil {
			return errs.WithEF(err, r.fields, "Failed to prepare report")
		}
		res = string(content) + "\n"
	}

	if n, err := file.WriteString(res); err != nil || n != len(res) {
		return errs.WithEF(err, r.fields, "Failed write report to file")
	}
	return nil
}
