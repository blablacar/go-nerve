package main

import (
	"testing"
)

func TestCreateService(t *testing.T) {
	//First Mock the configuration of a service
	var config NerveServiceConfiguration
	config.Name = "ApelyAgamakou"
	config.Host = "apely"
	config.Port = 1234
	config.CheckInterval = 2000
	var check NerveCheckConfiguration
	check.Type = "tcp"
	check.Port = 2345
	check.Host = "1.2.3.4"
	check.Timeout = 17.23
	config.Watcher.Checks = append(config.Watcher.Checks, check)

	//Now Test the service creation, based on MOCK config object
	service, err := CreateService(config, "Apely", false)
	if err != nil {
		t.Error("Creating service from MOCK Config, failed with ", err)
	}
	if service.Name != "ApelyAgamakou" {
		t.Error("Expecting Service Name to be 'ApelyAgamakou', got ", service.Name)
	}
	if service.Host != "apely" {
		t.Error("Expecting Service Host to be 'apely', got ", service.Host)
	}
	if service.Port != 1234 {
		t.Error("Expecting Service Port to be '1234', got ", service.Port)
	}
	if service.CheckInterval != 2000 {
		t.Error("Expecting Service Check Interval Name to be '2000', got ", service.Name)
	}
}
