package main

import (
	"testing"
)

func TestOpenConfiguration(t *testing.T) {
	config, err := OpenConfiguration("examples/nerve.conf.json")
	if err != nil {
		t.Fatal("Unable to open Configuration file", err)
	}
	if config.InstanceID != "mymachine" {
		t.Error("Expected instance_id to be 'mymachine', got ",config.InstanceID)
	}
	if config.LogLevel != "DEBUG" {
		t.Error("Expected log-level to be 'DEBUG', got ",config.LogLevel)
	}
	if config.IPv6 {
		t.Error("Expected ipv6 to be false, got true")
	}
	if len(config.Services) != 3 {
		t.Error("Expected 3 services to be loaded, got ",len(config.Services))
	}
}
