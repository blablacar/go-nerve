package nerve

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/n0rad/go-erlog/errs"
	"strconv"
	"sync"
)

type CheckSql struct {
	CheckCommon
	Username string
	Password string
	Request  string
	Protocol string
	Driver   string

	datasource string
}

func NewCheckSql() *CheckSql {
	return &CheckSql{
		CheckCommon: CheckCommon{Port: 3306},
		Username:    "root",
		Password:    "",
		Request:     "SELECT 1",
		Protocol:    "tcp",
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
	ip := IpLookupNoError(x.Host, s.PreferIpv4).String()
	x.datasource = x.Username + ":" + x.Password + "@" + x.Protocol + "([" + ip + "]:" + strconv.Itoa(x.Port) + ")/?timeout=" + strconv.Itoa(x.TimeoutInMilli) + "ms"
	x.fields = x.fields.WithField("datasource", x.datasource)
	return nil
}

func (x *CheckSql) Check() error {
	conn, err := sql.Open(x.Driver, x.datasource)
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
