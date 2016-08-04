package nerve

import (
	"encoding/json"
	"bytes"
	"fmt"
	"strconv"
)

type Report struct {
	Available            *bool             `json:"available"`
	UnavailableReason    string            `json:"unavailable_reason,omitempty"`
	Host                 string            `json:"host,omitempty"`
	Port                 int               `json:"port,omitempty"`
	Name                 string            `json:"name,omitempty"`
	HaProxyServerOptions string            `json:"haproxy_server_options,omitempty"`
	Weight               *uint8            `json:"weight"`
	Labels               map[string]string `json:"labels,omitempty"`
}

func NewReport(content []byte) (*Report, error) {
	var r Report
	err := json.Unmarshal(content, &r)
	return &r, err
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
		Port:                 s.Port,
		Name:                 s.Name,
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

	buffer.WriteString(fmt.Sprint(r.Available))
	buffer.WriteString(" ")
	buffer.WriteString(r.Name)
	buffer.WriteString(" ")
	buffer.WriteString(r.Host)
	buffer.WriteString(":")
	buffer.WriteString(strconv.Itoa(r.Port))
	if r.Weight != nil {
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Itoa(int(*r.Weight)))
	}
	return buffer.String()
}