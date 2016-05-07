package main

import (
	"fmt"
)

const (
	MonitorStopped = iota
	MonitorStarting
	MonitorRunning
)

type Monitor interface {
	Start(bus *MonitorConn) error
	Stop()
	Status()
	GetState() State
}

type State struct {
	Name   string
	Num    int
	Status string
	Info   map[string]string
}

func (s State) String() string {
	str := ""
	switch s.Num {
	case MonitorRunning:
		str = "running"
	case MonitorStarting:
		str = "starting"
	case MonitorStopped:
		str = "stopped"
	}
	if s.Status != "" {
		str = fmt.Sprintf("%s (%s)", str, s.Status)
	}
	if s.Info != nil {
		str = fmt.Sprintf("%s info %v", str, s.Info)
	}

	return str
}

type SMS struct {
	Name string
	From string
	Body string
}

func (s SMS) String() string {
	return fmt.Sprintf("&SMS{Dongle: %s, From: %s, Body: %s}", s.Name, s.From, s.Body)
}

type MonitorConn struct {
	State chan State
	Sms   chan SMS
	Ussd  chan string
}
