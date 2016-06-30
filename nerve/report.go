package nerve

import "encoding/json"

type Report struct {
	Available            bool              `json:"available"`
	Host                 string            `json:"host,omitempty"`
	Port                 int               `json:"port,omitempty"`
	Name                 string            `json:"name,omitempty"`
	HaProxyServerOptions string            `json:"haproxy_server_options,omitempty"`
	Weight               uint8             `json:"weight,omitempty"`
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
	return Report{
		Available:            status == nil,
		Host:                 s.Host,
		Port:                 s.Port,
		Name:                 s.Name,
		Weight:               s.CurrentWeight(),
		HaProxyServerOptions: s.HaproxyServerOptions,
		Labels:               s.Labels,
	}
}
