package main

import (
	"flag"
	"fmt"
	"github.com/blablacar/go-nerve/nerve"
	"github.com/ghodss/yaml"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

var Version = "No Version Defined"
var BuildTime = "1970-01-01_00:00:00_UTC"

func printVersion() {
	fmt.Println("Nerve")
	fmt.Println("Version :", Version)
	fmt.Println("Build Time :", BuildTime)
}

func processArgsGetConfigPath() (string, error) {
	var logLevel = flag.String("log-level", "info", "log level")
	var configPath = flag.String("config", "./nerve.yml", "the complete filename of the configuration file")
	var version = flag.Bool("version", false, "Display version and exit")
	flag.Parse()

	// version
	if *version {
		printVersion()
		os.Exit(0)
	}

	// log-level
	lvl, err := logs.ParseLevel(*logLevel)
	if err != nil {
		return "", errs.WithEF(err, data.WithField("input", logLevel), "Invalid log level")
	}
	logs.SetLevel(lvl)

	return *configPath, nil
}

func LoadConfig(configPath string) (*nerve.Nerve, error) {
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("file", configPath), "Failed to read configuration file")
	}

	conf := &nerve.Nerve{}
	err = yaml.Unmarshal(file, conf)
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("file", configPath), "Invalid configuration format")
	}

	return conf, nil
}

func waitForSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-sigs
	logs.Debug("Stop signal received")
}

func main() {
	configPath, err := processArgsGetConfigPath()
	if err != nil {
		logs.WithE(err).Fatal("Argument process failed")
	}

	nerve, err := LoadConfig(configPath)
	if err != nil {
		logs.WithE(err).Fatal("Cannot start, failed to load configuration")
	}

	if err := nerve.Init(); err != nil {
		logs.WithE(err).Fatal("Failed to init nerve")
	}

	go nerve.Start()
	waitForSignal()
	nerve.Stop()
}
