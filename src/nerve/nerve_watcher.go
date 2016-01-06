package nerve

import "nerve/checks"

type Watcher struct {
	Host string
	Port int
	Status int
	Error error
	CheckInterval int
	Checks []checks.CheckI
}

type WatcherI interface {
	Initialize(config NerveWatcherConfiguration) error
	Check() (int, error)
	GetCheckInterval() int
}

func(w *Watcher) GetCheckInterval() int {
	return w.CheckInterval
}

func(w *Watcher) Initialize(config NerveWatcherConfiguration) (err error) {
	w.CheckInterval = config.CheckInterval
	return nil
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

func CreateWatcher(config NerveWatcherConfiguration) (WatcherI, error) {
	var watcher Watcher
	watcher.Initialize(config)
	return &watcher, nil
}
