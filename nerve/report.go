package nerve

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
)

type Port int

func (p *Port) UnmarshalJSON(b []byte) error {
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		*p = Port(i)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		si, err := strconv.Atoi(s)
		if err != nil {
			return errs.WithEF(err, data.WithField("content", string(b)), "Invalid port value")
		}
		*p = Port(si)
		return nil
	} else {
		return errs.WithEF(err, data.WithField("content", string(b)), "Failed to parse port")
	}
}

type Report struct {
	Available            *bool             `json:"available"`
	UnavailableReason    string            `json:"unavailable_reason,omitempty"`
	Host                 string            `json:"host,omitempty"`
	Port                 Port              `json:"port,omitempty"`
	Name                 string            `json:"name,omitempty"`
	HaProxyServerOptions string            `json:"haproxy_server_options,omitempty"`
	Weight               *uint8            `json:"weight"`
	Labels               map[string]string `json:"labels,omitempty"`
}

type report Report

func NewReport(content []byte) (*Report, error) {
	var r Report
	err := json.Unmarshal(content, &r)
	return &r, err
}

func (r *Report) UnmarshalJSON(b []byte) error {
	var rr report
	if err := json.Unmarshal(b, &rr); err != nil {
		return err
	}
	if rr.Port == 0 || rr.Host == "" || rr.Name == "" {
		return errs.WithF(data.Fields{"report": rr}, "Missing Port or Host or Name in report.")
	}
	if rr.Available != nil && *rr.Available == false {
		w := uint8(0)
		rr.Weight = &w
	}
	*r = Report(rr)
	return nil
}

func (r *Report) toJson() ([]byte, error) {
	return json.Marshal(r)
}

func toReport(status error, s *Service) Report {
	weight := s.CurrentWeight()
	boolStatus := status == nil
	r := Report{
		Available:            &boolStatus,
		Host:                 s.Host,
		Port:                 Port(s.Port),
		Name:                 s.ReporterServiceName,
		Weight:               &weight,
		HaProxyServerOptions: s.HaproxyServerOptions,
		Labels:               s.Labels,
	}
	if status != nil {
		r.UnavailableReason = status.Error()
	}
	return r
}

func (r *Report) String() string {
	var buffer bytes.Buffer
	if r.Available == nil {
		buffer.WriteString(fmt.Sprint(true))
	} else {
		buffer.WriteString(fmt.Sprint(*r.Available))
	}
	buffer.WriteString(" ")
	buffer.WriteString(r.Name)
	buffer.WriteString(" ")
	buffer.WriteString(r.Host)
	buffer.WriteString(":")
	buffer.WriteString(strconv.Itoa(int(r.Port)))
	if r.Weight != nil {
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(int(*r.Weight)))
	}
	return buffer.String()
}
