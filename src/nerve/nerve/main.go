package main

import (
	log "github.com/Sirupsen/logrus"
	"os/signal"
	"nerve"
	"time"
	"os"
	"flag"
)

// The function which is used to manageSignal
// When termination signal is received, we send a message to a chan
func manageSignal(c <-chan os.Signal, working chan<-bool) {
        toBreak := true
        for i := 0; toBreak; i=i+1 {
		select {
		case _signal := <-c:
			if _signal == os.Kill {
				log.Debug("Nerve: Kill Signal Received")
			}
			if _signal == os.Interrupt {
				log.Debug("Nerve: Interrupt Signal Received")
			}
			log.Info("Nerve: Shutdown Signal Received")
			working <- false
			toBreak = false
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

// Set the Logrus global log level
// Converted from a configuration string
func setLogLevel(logLevel string) {
	// Set the Log Level, extracted from the command line
	switch logLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	// The Log Level, from the Sirupsen/logrus level
	var logLevel = flag.String("log-level", "WARN", "A value to choose between [DEBUG INFO WARN FATAL], can be overriden by config file")
	// The configuration filename
	var configurationFileName = flag.String("config", "./nerve.json.conf", "the complete filename of the configuration file")
	// Parse all command line options
	flag.Parse()

	setLogLevel(*logLevel)

	log.Debug("Nerve: Starting config file parsing")
	nerveConfiguration , err := nerve.OpenConfiguration(*configurationFileName)
	if err != nil {
		// If an error is raised when parsing configuration file
		// the configuration object can be either empty, either incomplete
		log.Fatal("Nerve: Unable to load Configuration (",err,")")
		// So of the configuration is incomplete, exit the program now
		os.Exit(1)
	}
	if nerveConfiguration.LogLevel != "" {
		setLogLevel(nerveConfiguration.LogLevel)
	}
	log.Info("Nerve: Starting Run of Instance [",nerveConfiguration.InstanceID,"]")
	c := make(chan os.Signal,1)
	signal.Notify(c, os.Interrupt, os.Kill)
	log.Debug("Nerve: Signal Channel notification setup done")
	working := make(chan bool)
	go manageSignal(c,working)
	log.Debug("Nerve: Signal Management Started")
	finished:= make(chan bool)
	go nerve.Run(working,finished,nerveConfiguration)
	log.Debug("Nerve: Go routine launched")

	log.Debug("Nerve: Waiting for main process to Stop")
	isFinished := <-finished
	if (isFinished) {
		log.Debug("Nerve: Main routine closed correctly")
	}else {
		log.Warn("Nerve: Main routine closed incorrectly")
	}
	log.Info("Nerve: Shutdown of Instance [",nerveConfiguration.InstanceID,"] Completed")
}
