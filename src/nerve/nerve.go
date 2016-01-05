package nerve

import (
	log "github.com/Sirupsen/logrus"
	"time"
	"sync"
)

const (
	StatusOK int = 0
	StatusKO int = 1
	StatusUnknown int = 2
)

var closeNerveChan chan bool
var waitGroup sync.WaitGroup

func Initialize() (watcher WatcherI, reporter ReporterI) {
	var err error
	var configChecks []NerveCheckConfiguration
	var configWatcher NerveWatcherConfiguration
        configWatcher.CheckInterval = 2
	configWatcher.Checks = configChecks
	watcher, err = CreateWatcher(configWatcher)
	if err != nil {
		log.Warn("Nerve: CreateWatcher Error")
	}
	reporter, err = CreateReporter("console",nil)
	return watcher, reporter
}

func Run(work <-chan bool,finished chan<-bool, nerveConfig NerveConfiguration) {
	log.Debug("Nerve: Run function started")
	toBreak := true
	for i := 0; toBreak; i=i+1 {
		log.Debug("Nerve: Watcher Run [",i,"]")
		watcher, reporter := Initialize()
		status, err := watcher.Check()
		if err != nil {
			log.Warn("Nerve: Watcher Error")
		}else {
			reporter.Report("127.0.0.1","8080","la",status)
		}
		select {
		case workNonBlocking := <-work:
			log.Debug("Nerve: Run function Close Signal Received")
			toBreak = workNonBlocking
		default:
			toBreak = true
		}
		time.Sleep(time.Second * 1)
	}
	finished <- true
	log.Debug("Nerve: Run function termination")
}
