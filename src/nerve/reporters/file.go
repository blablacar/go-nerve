package reporters

import (
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

const REPORTER_FILE_TYPE = "FILE"

type FileReporter struct {
	Reporter
	Path string
	Filename string
	Mode string
}

func(fr *FileReporter) Initialize(IP string, Port int, Rise int, Fall int, Weight int, ServiceName string, InstanceID string, HAProxyServerOptions string, Tags []string) error {
	fr.Path = "/tmp/"
	fr.Filename = "nerve.report"
	fr._type = REPORTER_FILE_TYPE
	fr.Mode = "write"
	fr.SetBaseConfiguration(IP,Port,Rise,Fall,Weight,ServiceName,InstanceID,HAProxyServerOptions,Tags)
	return nil
}

func(fr *FileReporter) SetFileConfiguration(Path string, Filename string, Mode string) {
	if Path != "" {
		fr.Path = Path
	}
	if Filename != "" {
		fr.Filename = Filename
	}
	if Mode != "" {
		fr.Mode = Mode
	}
}

func(fr *FileReporter) Report(Status int) error {
	if fr.Mode == "append" {
		file, err := os.OpenFile(fr.Path+"/"+fr.Filename,os.O_APPEND|os.O_CREATE|os.O_WRONLY,0644)
		if err != nil {
			log.WithField("Filename",fr.Path+"/"+fr.Filename).WithError(err).Warn("Unable to open File to Report to")
			return err
		}
		if _, err = file.WriteString(fr.GetJsonReporterData()); err != nil {
			log.WithField("Filename",fr.Path+"/"+fr.Filename).WithError(err).Warn("Unable to Append to File to Report to")
			return err
		}
		file.Close()
	} else {
		err := ioutil.WriteFile(fr.Path+"/"+fr.Filename,[]byte(fr.GetJsonReporterData()),0644)
		if err != nil {
			log.WithField("Filename",fr.Path+"/"+fr.Filename).WithError(err).Warn("Unable to Write File to Report to")
			return err
		}
	}
	return nil
}

func(fr *FileReporter) Destroy() error {
	return nil
}

func(fr *FileReporter) GetType() string {
	return fr._type
}
