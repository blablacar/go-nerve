package watchers

import "nerve/configuration"
import "nerve/watchers/checks"

const (
	StatusOK int = 0
	StatusKO int = 1
	StatusUnknown int = 2
)

type Watcher struct {
	Host string
	Port int
	Status int
	Error error
	Checks []checks.CheckI
}

type WatcherI interface {
	Initialize(config configuration.NerveWatcherConfiguration) error
	Check() (status int, err error)
}

func(w Watcher) Initialize(config configuration.NerveWatcherConfiguration) (err error) {
	return nil
}

func(w Watcher) Check() (status int, err error) {
	if w.Checks == nil {
		return StatusOK, nil
	}
	for i:=0; i < len(w.Checks); i++ {
		status, err := w.Checks[i].DoCheck()
		if err != nil || status != StatusOK {
			return status, err
		}
	}
	return StatusOK , nil
}

func CreateWatcher(config configuration.NerveWatcherConfiguration) (watcher WatcherI, err error) {
	watcher = new(Watcher)
	watcher.Initialize(config)
	return watcher, nil
}
