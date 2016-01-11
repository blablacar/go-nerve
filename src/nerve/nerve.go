package nerve

import (
	log "github.com/Sirupsen/logrus"
	"time"
	"sync"
	"errors"
)

const (
	StatusOK int = 0
	StatusKO int = 1
	StatusUnknown int = 2
)

var closeNerveChan chan bool
var servicesWaitGroup sync.WaitGroup

func createServices(config NerveConfiguration) ([]NerveService, error) {
	var services []NerveService
	if len(config.Services) > 0 {
		for i:=0; i < len(config.Services); i++ {
			service, err := CreateService(config.Services[i],config.InstanceID,config.IPv6)
			if err != nil {
				log.Warn("Error when creating a service (",err,")")
				return services, err
			}
			services = append(services,service)
		}
	}else {
		err := errors.New("no service found in configuration")
		return services, err
	}
	return services, nil
}

func Run(stop <-chan bool,finished chan<-bool, nerveConfig NerveConfiguration) {
	log.Debug("Nerve: Run function started")
	services , err := createServices(nerveConfig)
	if err != nil {
		finished <- false
	}else {
		servicesWaitGroup.Add(len(services))
		stopper := make(chan bool, len(services))
		//Start Services
		for i := 0; i < len(services); i++ {
			go services[i].Run(stopper)
		}

		// Wait for the stop signal
		Loop:
		for {
			select {
			case hasToStop := <-stop:
				if hasToStop {
					log.Debug("Nerve: Run function Close Signal Received")
				}else {
					log.Debug("Nerve: Run function Close Signal Received (but a strange false one)")
				}
				break Loop
			default:
				time.Sleep(time.Second * 1)
			}
		}

		//Inform all services to stop
		for i := 0; i < len(services); i++ {
			stopper <- true
		}

		log.Debug("Nerve: Wait for all services to stop")
		//Wait for all services to shutdown
		servicesWaitGroup.Wait()
		finished <- true
	}
	log.Debug("Nerve: Run function termination")
}
