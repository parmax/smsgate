package main

import (
	"gopkg.in/gcfg.v1"
)

type GateConfig struct {
	Dev_AMI map[string]*MonitorAMIConfig
}

func NewGateConfig(filename string) (*GateConfig, error) {
	var cfg GateConfig
	err := gcfg.ReadFileInto(&cfg, filename)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
