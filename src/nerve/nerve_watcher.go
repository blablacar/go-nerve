package nerve

import (
	log "github.com/Sirupsen/logrus"
	"nerve/checks"
	"errors"
)

type Watcher struct {
	Host string
	Port int
	Status int
	Error error
	CheckInterval int
	Checks []checks.CheckI
}

type WatcherI interface {
	Initialize(config NerveWatcherConfiguration, IP string, Host string, Port int, ipv6 bool) error
	Check() (int, error)
	GetCheckInterval() int
}

func(w *Watcher) GetCheckInterval() int {
	return w.CheckInterval
}

func createChecks(config NerveWatcherConfiguration, IP string, Host string, Port int, ipv6 bool) ([]checks.CheckI, error) {
        var _checks []checks.CheckI
        if len(config.Checks) > 0 {
                for i:=0; i < len(config.Checks); i++ {
			var param1 string
			var param2 string
			var param3 string
			var param4 string
			var param5 []string
			switch config.Checks[i].Type {
			case "http"    :
				param1 = config.Checks[i].Uri
				param2 = ""
				param3 = ""
				param4 = ""
				param5 = []string{""}
			case "mysql"   :
				param1 = config.Checks[i].User
				param2 = config.Checks[i].Password
				param3 = config.Checks[i].SQLRequest
				param4 = ""
				param5 = []string{""}
			case "rabbitmq" :
				param1 = config.Checks[i].User
				param2 = config.Checks[i].Password
				param3 = config.Checks[i].Queue
				param4 = config.Checks[i].VHost
				param5 = []string{""}
			case "zkflag" :
				param1 = config.Checks[i].Path
				param2 = ""
				param3 = ""
				param4 = ""
				param5 = config.Checks[i].Hosts
			case "httpproxy" :
				param1 = config.Checks[i].User
				param2 = config.Checks[i].Password
				param3 = config.Checks[i].Host
				param4 = config.Checks[i].Port
				param5 = config.Checks[i].URLs
			default:
				param1 = ""
				param2 = ""
				param3 = ""
				param4 = ""
				param5 = []string{""}
			}
                        check, err := checks.CreateCheck(
				config.Checks[i].Type,
				IP,
				Host,
				Port,
				config.Checks[i].ConnectTimeout,
				ipv6,
				param1,
				param2,
				param3,
				param4,
				param5,
			)
                        if err != nil {
                                log.Warn("Error when creating a check (",err,")")
                                return _checks, err
                        }
                        _checks = append(_checks,check)
                }
        }else {
                err := errors.New("no check found in configuration")
                return _checks, err
        }
        return _checks, nil
}

func(w *Watcher) Initialize(config NerveWatcherConfiguration, IP string, Host string, Port int, ipv6 bool) (error) {
	w.CheckInterval = config.CheckInterval
	var err error
	w.Checks, err = createChecks(config,IP,Host,Port,ipv6)
	return err
}

func(w *Watcher) Check() (int, error) {
	var status int
	var err error
	if w.Checks == nil {
		return StatusOK, nil
	}
	for i:=0; i < len(w.Checks); i++ {
		status, err = w.Checks[i].DoCheck()
		if err != nil || status != StatusOK {
			return status, err
		}
	}
	return StatusOK , nil
}

func CreateWatcher(config NerveWatcherConfiguration, IP string, Host string, Port int, ipv6 bool) (WatcherI, error) {
	var watcher Watcher
	err := watcher.Initialize(config,IP,Host,Port,ipv6)
	if (err != nil) {
		log.Warn("Error when initialising Watcher [",Host,":",Port,"] (",err,")")
	}
	return &watcher, err
}
