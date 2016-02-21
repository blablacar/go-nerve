package reporters

import (
	log "github.com/Sirupsen/logrus"
	"strconv"
)

type Reporter struct {
	IP string
	Port int
	Rise int
	Fall int
	Weight int
	HAProxyServerOptions string
	ServiceName string
	InstanceID string
	lastStatuses []int
	Tags []string
	_type string
}

func(r *Reporter) CanReport(status int) bool {
	var _lastStatuses []int
	_lastStatuses = append(_lastStatuses, status)
	_lastStatuses = append(_lastStatuses, r.lastStatuses...)
	if len(_lastStatuses) > max(r.Rise,r.Fall) {
		r.lastStatuses = _lastStatuses[:len(_lastStatuses) -1]
	}else {
		r.lastStatuses = _lastStatuses
	}
	if status == 0 {
		if len(r.lastStatuses) < r.Rise {
			log.Debug("Not Enough Valid Check to report got[",len(r.lastStatuses),"] expect [",r.Rise,"]")
			return false
		}
		for i := 0; i < r.Rise; i++ {
			if r.lastStatuses[i] != 0 {
				log.Debug("Not Enough Valid Check to report got[",i+1,"] expect [",r.Rise,"]")
				return false
			}
		}
	}else {
		if len(r.lastStatuses) < r.Fall {
			log.Debug("Not Enough Unvalid Check to report got[",len(r.lastStatuses),"] expect [",r.Fall,"]")
			return false
		}
		for i := 0; i < r.Fall; i++ {
			if r.lastStatuses[i] == 0 {
				log.Debug("Not Enough Unvalid Check to report got[",i+1,"] expect [",r.Fall,"]")
				return false
			}
		}
	}
	return true
}

func(r *Reporter) SetBaseConfiguration(IP string, Port int, Rise int, Fall int, Weight int, ServiceName string, InstanceID string, HAProxyServerOptions string, Tags []string) {
	r.IP = IP
	r.Port = Port
	if Rise > 0 {
		r.Rise = Rise
	}else {
		r.Rise = 1
	}
	if Fall > 0 {
		r.Fall = Fall
	}else {
		r.Fall = 1
	}
	if len(Tags) > 0 {
		r.Tags = Tags
	}
	r.Weight = Weight
	r.InstanceID = InstanceID
	r.HAProxyServerOptions = HAProxyServerOptions
	r.ServiceName = ServiceName
}

func(r *Reporter) GetJsonReporterData(inMaintenance bool) string {
	var jsonReporterData string
	jsonReporterData = "{\"host\":\"" + r.IP + "\",\"port\":" + strconv.Itoa(r.Port) + ","
	jsonReporterData += "\"name\":\"" + r.InstanceID + "\""
	if r.Weight > 0 {
		jsonReporterData += ",\"weight\":" + strconv.Itoa(r.Weight)
	}
	if r.HAProxyServerOptions != "" {
		jsonReporterData += ",\"haproxy_server_options\":\"" + r.HAProxyServerOptions + "\""
	}
	jsonReporterData += ",\"maintenance\":"
	if inMaintenance {
		jsonReporterData += "true"
	}else {
		jsonReporterData += "false"
	}
	if len(r.Tags) > 0 {
		jsonReporterData += ",\"tags\":["
		for i, tag := range r.Tags {
			if i > 0 {
				jsonReporterData += ","
			}
			jsonReporterData += "\""+tag+"\""
		}
		jsonReporterData += "]"
	}
	jsonReporterData += "}"
	return jsonReporterData
}

func max(val1 int, val2 int) int {
	if val1 > val2 {
		return val1
	}
	return val2
}
