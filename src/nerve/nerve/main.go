package main

import (
	log "github.com/Sirupsen/logrus"
	"os/signal"
	"nerve"
	"time"
	"os"
)

func manageSignal(c <-chan os.Signal, working chan<-bool) {
        toBreak := true
        for i := 0; toBreak; i=i+1 {
		select {
		case _signal := <-c:
			if _signal == os.Kill {
				log.Info("Kill Signal Received")
			}
			log.Info("Kill Signal Received")
			working <- false
			toBreak = false
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func main() {

	log.SetLevel(log.InfoLevel)
	log.Info("nerve: starting run")
	c := make(chan os.Signal,1)
	working := make(chan bool)
	finished:= make(chan bool)
	signal.Notify(c, os.Interrupt, os.Kill)
	go manageSignal(c,working)
	go nerve.Run(working,finished)
	log.Info("nerve: Go routine launched")

	log.Info("nerve: Waiting for main process to Stop")
	isFinished := <-finished
	if (isFinished) {
		log.Info("nerve: finished correctly")
	}else {
		log.Warn("nerve: finished incorrectly")
	}
	log.Info("nerve: closing run")
}
