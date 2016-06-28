package nerve

import (
	"bytes"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"sync"
	"text/template"
)

type CheckSql struct {
	CheckCommon
	Username   string
	Password   string
	Request    string
	Datasource string
	Driver     string

	templatedDatasource string
}

func NewCheckSql() *CheckSql {
	return &CheckSql{
		CheckCommon: CheckCommon{Port: 3306},
		Username:    "root",
		Password:    "",
		Request:     "SELECT 1",
		Datasource:  "{{.Username}}:{{.Password}}@tcp([{{.Host}}]:{{.Port}})/?timeout={{.TimeoutInMilli}}ms",
		Driver:      "mysql",
	}
}

func (x *CheckSql) Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
	x.CommonRun(x, statusChange, stop, doneWait)
}

func (x *CheckSql) Init(s *Service) error {
	if err := x.CheckCommon.CommonInit(s); err != nil {
		return err
	}

	switch x.Driver {
	case "mysql", "postgres":
	default:
		return errs.WithF(x.fields.WithField("driver", x.Driver), "Unsupported driver")
	}

	template, err := template.New("datasource").Parse(x.Datasource)
	if err != nil {
		return errs.WithEF(err, x.fields, "Failed to parse datasource template")
	}

	var buff bytes.Buffer
	if err := template.Execute(&buff, x); err != nil {
		return errs.WithEF(err, x.fields, "Datasource templating failed")
	}
	x.templatedDatasource = buff.String()
	logs.WithF(x.fields.WithField("datasource", x.templatedDatasource)).Debug("datasource templated")
	return nil
}

func (x *CheckSql) Check() error {
	conn, err := sql.Open(x.Driver, x.templatedDatasource)
	if err != nil {
		return errs.WithEF(err, x.fields, "Cannot open connection")
	}
	defer conn.Close()
	rows, err := conn.Query(x.Request)
	if err != nil {
		return errs.WithEF(err, x.fields, "Check query failed")
	}
	defer rows.Close()
	return nil
}
