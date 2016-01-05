package nerve

import (
	"fmt"
	"nerve/watchers"
	"nerve/reporters"
	"nerve/configuration"
	"time"
	"sync"
)

var closeNerveChan chan bool
var waitGroup sync.WaitGroup

func Initialize() (watcher watchers.WatcherI, reporter reporters.ReporterI) {
	var err error
	var configChecks []configuration.NerveCheckConfiguration
	var configWatcher configuration.NerveWatcherConfiguration
	configWatcher.Host = "127.0.0.1"
        configWatcher.Port = 8080
        configWatcher.CheckInterval = 2
	configWatcher.Checks = configChecks
	watcher, err = watchers.CreateWatcher(configWatcher)
	if err != nil {
		fmt.Println("CreateWatcher Error")
	}
	reporter, err = reporters.CreateReporter("console",nil)
	return watcher, reporter
}

func Run(work <-chan bool,finished chan<-bool) {
	fmt.Println("Nerve started")
	toBreak := true
	for i := 0; toBreak; i=i+1 {
		fmt.Println("Watcher ", i)
		watcher, reporter := Initialize()
		status, err := watcher.Check()
		if err != nil {
			fmt.Println("Watcher Error")
		}else {
			reporter.Report("127.0.0.1","8080","la",status)
		}
		select {
		case workNonBlocking := <-work:
			fmt.Println("Close Signal Received")
			toBreak = workNonBlocking
		default:
			toBreak = true
		}
		time.Sleep(time.Second * 1)
	}
	finished <- true
	fmt.Println("Nerve closed")
}
