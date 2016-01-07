package reporters

import (
	log "github.com/Sirupsen/logrus"
)

type Reporter struct {
	IP string
	Port int
	Rise int
	Fall int
	lastStatuses []int
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

func max(val1 int, val2 int) int {
	if val1 > val2 {
		return val1
	}
	return val2
}
