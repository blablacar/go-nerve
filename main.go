package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"bytes"

	"github.com/blablacar/go-nerve/nerve"
	"github.com/ghodss/yaml"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"github.com/spf13/cobra"
)

var Version = "No Version Defined"
var BuildTime = "1970-01-01_00:00:00_UTC"

func envToMap() (map[string]string, error) {
	envMap := make(map[string]string)
	var err error

	for _, v := range os.Environ() {
		split_v := strings.SplitN(v, "=", 2)
		envMap[split_v[0]] = split_v[1]
	}
	return envMap, err
}

func LoadConfig(configPath string) (*nerve.Nerve, error) {
	envMap, err := envToMap()
	if err != nil {
		return nil, errs.WithE(err, "Failed to create environment var map to template configuration")
	}

	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("file", configPath), "Failed to read configuration file")
	}

	templatingBuffer := new(bytes.Buffer)
	t, err := template.New("tmpl1").Parse(string(file))
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("content", string(file)), "Failed to parse configuration as template")
	}
	if err := t.Execute(templatingBuffer, envMap); err != nil {
		return nil, errs.WithE(err, "Failed to template configuration")
	}

	confTemplatedFirstPass := templatingBuffer.String()
	templatingBuffer.Reset()
	t2, err := template.New("tmpl2").Parse(confTemplatedFirstPass)
	if err != nil {
		return nil, errs.WithEF(err, data.WithField("content", string(file)), "Failed to parse configuration as template")
	}
	if err := t2.Execute(templatingBuffer, envMap); err != nil {
		return nil, errs.WithE(err, "Failed to template configuration")
	}

	if err := os.MkdirAll("/run/nerve", os.ModePerm); err != nil {
		logs.Warn("Failed to prepare /run/nerve directory")
	}
	if err := ioutil.WriteFile("/run/nerve/config.yaml", templatingBuffer.Bytes(), 0644); err != nil {
		logs.Warn("Failed to write template configuration to /run/nerve/config.yaml")
	}

	conf := &nerve.Nerve{}
	err = yaml.Unmarshal(templatingBuffer.Bytes(), conf)
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

func sigQuitThreadDump() {
	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			stacktrace := make([]byte, 10<<10)
			length := runtime.Stack(stacktrace, true)
			fmt.Println(string(stacktrace[:length]))

			ioutil.WriteFile("/tmp/"+strconv.Itoa(os.Getpid())+".dump", stacktrace[:length], 0644)
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	sigQuitThreadDump()

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

			if logLevel != "" {
				level, err := logs.ParseLevel(logLevel)
				if err != nil {
					logs.WithField("value", logLevel).Fatal("Unknown log level")
				}
				logs.SetLevel(level)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				logs.Fatal("Nerve require a configuration file as argument")
			}
			nerve, err := LoadConfig(args[0])
			if err != nil {
				logs.WithE(err).Fatal("Cannot start, failed to load configuration")
			}

			if err := nerve.Init(Version, BuildTime, logLevel != ""); err != nil {
				logs.WithE(err).Fatal("Failed to init nerve")
			}

			startStatus := make(chan error)
			go nerve.Start(startStatus)
			if status := <-startStatus; status != nil {
				logs.WithE(status).Fatal("Failed to start nerve")
			}
			waitForSignal()
			nerve.Stop()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", "", "Set log level")
	rootCmd.PersistentFlags().BoolVarP(&version, "version", "V", false, "Display version")

	if err := rootCmd.Execute(); err != nil {
		logs.WithE(err).Fatal("Failed to process args")
	}
}
