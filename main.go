package main

import (
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
	"github.com/spf13/cobra"
)

var Version = "No Version Defined"
var BuildTime = "1970-01-01_00:00:00_UTC"

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
	var logLevel string
	var version bool

	rootCmd := &cobra.Command{
		Use: "nerve config.yml",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if version {
				fmt.Println("Nerve")
				fmt.Println("Version :", Version)
				fmt.Println("Build Time :", BuildTime)
				os.Exit(0)
			}

			level, err := logs.ParseLevel(logLevel)
			if err != nil {
				logs.WithField("value", logLevel).Fatal("Unknown log level")
			}
			logs.SetLevel(level)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				logs.Fatal("Nerve require a configuration file as argument")
			}
			nerve, err := LoadConfig(args[0])
			if err != nil {
				logs.WithE(err).Fatal("Cannot start, failed to load configuration")
			}

			if err := nerve.Init(); err != nil {
				logs.WithE(err).Fatal("Failed to init nerve")
			}

			go nerve.Start()
			waitForSignal()
			nerve.Stop()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", "info", "Set log level")
	rootCmd.PersistentFlags().BoolVarP(&version, "version", "V", false, "Display version")

	if err := rootCmd.Execute(); err != nil {
		logs.WithE(err).Fatal("Failed to process args")
	}
}
