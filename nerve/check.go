package nerve

import (
	"encoding/json"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
)

type Checker interface {
	Init(service *Service) error
	Check() error
	getCommon() *CheckCommon
	GetFields() data.Fields
}

type CheckCommon struct {
	Type           string
	Host           string
	Port           int
	TimeoutInMilli int
	Rise           int
	Fall           int

	fields data.Fields
	ip     string
}

func (c *CheckCommon) GetFields() data.Fields {
	return c.fields
}

func (c *CheckCommon) Init(s *Service) error {
	c.Host = s.Host
	c.Port = s.Port
	//c.Rise = s.Rise TODO by check
	//c.Fall = s.Fall
	c.TimeoutInMilli = 2000
	c.fields = s.fields
	c.ip = IpLookupNoError(c.Host, s.PreferIpv4).String()
	return nil
}

func (c *CheckCommon) getCommon() *CheckCommon {
	return c
}

func CheckerFromJson(data []byte, s *Service) (string, Checker, error) {
	t := &CheckCommon{}
	if err := json.Unmarshal([]byte(data), t); err != nil {
		return "", nil, errs.WithE(err, "Failed to unmarshall check type")
	}

	fields := s.fields.WithField("type", t.Type)
	var typedCheck Checker
	switch t.Type {
	case "http":
		typedCheck = NewCheckHttp()
	case "tcp":
		typedCheck = NewCheckTcp()
	case "sql":
		typedCheck = NewCheckSql()
	case "amqp":
		typedCheck = NewCheckAmqp()
	case "exec":
		typedCheck = NewCheckExec()
	default:
		return "", nil, errs.WithF(fields, "Unsupported check type")
	}

	if err := typedCheck.getCommon().Init(s); err != nil {
		return "", nil, errs.WithEF(err, fields, "Failed to init common check")
	}

	if err := json.Unmarshal([]byte(data), &typedCheck); err != nil {
		return "", nil, errs.WithEF(err, fields, "Failed to unmarshall check")
	}

	if err := typedCheck.Init(s); err != nil {
		return "", nil, errs.WithEF(err, fields, "Failed to init check")
	}
	return t.Type, typedCheck, nil
}
