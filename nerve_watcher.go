package main

import (
	log "github.com/Sirupsen/logrus"
	"errors"
	"strconv"
	"github.com/blablacar/go-nerve/checks"
)

type Watcher struct {
	Host string
	Port int
	Status int
	Error error
	Checks []checks.CheckI
	MaintenanceChecks []checks.CheckI
}

type WatcherI interface {
	Initialize(config NerveWatcherConfiguration, IP string, Host string, Port int, ipv6 bool) error
	Check(CheckForMaintenance bool) (int, error)
}

func createChecks(config NerveWatcherConfiguration, IP string, Host string, Port int, ipv6 bool, isMaintenance bool) ([]checks.CheckI, error) {
        var _checks []checks.CheckI
	port := Port
	var ncc []NerveCheckConfiguration
	if isMaintenance {
		ncc = config.MaintenanceChecks
	}else {
		ncc = config.Checks
	}
        if len(ncc) > 0 {
                for i:=0; i < len(ncc); i++ {
			var param1 string
			var param2 string
			var param3 string
			var param4 string
			var param5 []string
			switch ncc[i].Type {
			case "http"    :
				if ncc[i].Port != 0 {
					port = ncc[i].Port
				}
				param1 = ncc[i].Uri
				param2 = ""
				param3 = ""
				param4 = ""
				param5 = []string{""}
			case "mysql"   :
				if ncc[i].Port != 0 {
					port = ncc[i].Port
				}
				param1 = ncc[i].User
				param2 = ncc[i].Password
				param3 = ncc[i].SQLRequest
				param4 = ""
				param5 = []string{""}
			case "rabbitmq" :
				if ncc[i].Port != 0 {
					port = ncc[i].Port
				}
				param1 = ncc[i].User
				param2 = ncc[i].Password
				param3 = ncc[i].Queue
				param4 = ncc[i].VHost
				param5 = []string{""}
			case "zkflag" :
				param1 = ncc[i].Path
				param2 = ""
				param3 = ""
				param4 = ""
				param5 = ncc[i].Hosts
			case "httpproxy" :
				param1 = ncc[i].User
				param2 = ncc[i].Password
				param3 = ncc[i].Host
				param4 = strconv.Itoa(ncc[i].Port)
				param5 = ncc[i].URLs
			default:
				if ncc[i].Port != 0 {
					port = ncc[i].Port
				}
				param1 = ""
				param2 = ""
				param3 = ""
				param4 = ""
				param5 = []string{""}
			}
                        check, err := checks.CreateCheck(
				ncc[i].Type,
				IP,
				Host,
				port,
				ncc[i].ConnectTimeout,
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
	var err error
	w.Checks, err = createChecks(config,IP,Host,Port,ipv6,false)
	if len(config.MaintenanceChecks) > 0 {
		w.MaintenanceChecks, err = createChecks(config,IP,Host,Port,ipv6,true)
	}
	return err
}

func(w *Watcher) Check(CheckForMaintenance bool) (int, error) {
	var status int
	var err error
        var _checks []checks.CheckI
	if CheckForMaintenance {
		_checks = w.MaintenanceChecks
	}else {
		_checks = w.Checks
	}
	if _checks == nil {
		return StatusOK, nil
	}
	for i:=0; i < len(_checks); i++ {
		status, err = _checks[i].DoCheck()
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
