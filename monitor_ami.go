package main

import (
	"errors"
	"fmt"
	"github.com/bit4bit/gami"
	"github.com/xlab/at/sms"
	"github.com/xlab/at/util"
	"net/textproto"
	"strings"
)

type MonitorAMIConfig struct {
	Name, Dev, Host, User, Password string
}

type MonitorAMI struct {
	state *State
	stop  chan bool
	ami   *gami.AMIClient
	*MonitorAMIConfig
}

func NewMonitorAMI(conf *MonitorAMIConfig) (*MonitorAMI, error) {
	m := &MonitorAMI{
		&State{
			Name:   conf.Name,
			Num:    MonitorStopped,
			Status: "",
		},
		make(chan bool),
		nil,
		conf,
	}

	return m, nil
}

func (m *MonitorAMI) setState(bus *MonitorConn, s *State) {
	s.Name = m.Name
	m.state = s
	bus.State <- *s
}

func (m *MonitorAMI) setStatus(bus *MonitorConn, msg string) {
	m.setState(bus, &State{Num: MonitorRunning, Status: msg})
}

func (m *MonitorAMI) GetState() State {
	return *m.state
}

func (m *MonitorAMI) Start(bus *MonitorConn) error {
	if m.state.Num != MonitorStopped {
		return errors.New("Monitor already running")
	}
	go func() {
		var err error
		m.setState(bus, &State{Num: MonitorStarting})
		ami, err := gami.Dial(m.Host)
		if err != nil {
			goto stop
		}
		m.ami = ami

		ami.Run()

		err = ami.Login(m.User, m.Password)
		if err != nil {
			goto stop
		}
		ami.Action("Events", gami.Params{"EventMask": "call"})
		m.setState(bus, &State{Num: MonitorRunning})
		for err == nil {
			select {
			case <-m.stop:
				goto stop
			case err = <-m.ami.NetError:
			case err = <-m.ami.Error:
				if e, ok := err.(*textproto.ProtocolError); ok {
					m.setState(bus, &State{Num: MonitorRunning, Status: fmt.Sprint("MIME error:", e)})
					err = nil
				}
			case ev := <-ami.Events:
				//log.Println("AMI:", m.name, ev)
				p := ev.Params
				switch ev.ID {
				case "DongleDeviceEntry":
					if p["Device"] == m.Dev {
						rssi := strings.Split(p["Rssi"], ", ")
						gsmreg := strings.Split(p["Gsmregistrationstatus"], ", ")
						i := map[string]string{
							"imei":   p["Imeistate"],
							"imsi":   p["Imsistate"],
							"msisdn": p["Subscribernumber"],
							"prov":   p["Providername"],
							"cell":   p["Cellid"],
							"area":   p["Locationareacode"],
							"rssi":   rssi[0],
							"reg":    gsmreg[0],
							"net":    gsmreg[1],
						}
						m.setState(bus, &State{Num: MonitorRunning, Info: i})
					}
				case "DongleNewCMGR":
					if p["Device"] == m.Dev {
						lines := strings.Split(p["Message"], "\\n")
						if len(lines) < 2 {
							m.setStatus(bus, "unexpected line count when parsing CMGR")
							continue
						}
						octets, err := util.Bytes(lines[1])
						if err != nil {
							m.setStatus(bus, err.Error())
							continue
						}

						var msg sms.Message
						if _, err = msg.ReadFrom(octets); err != nil {
							m.setStatus(bus, err.Error())
							continue
						}
						bus.Sms <- SMS{m.Name, string(msg.Address), msg.Text}
					}
				case "DongleStatus":
					if p["Device"] == m.Dev {
						m.setStatus(bus, fmt.Sprint("Status:", p["Status"]))
					}
				}
			}
		}
	stop:
		m.ami = nil

		s := ""
		if err != nil {
			s = err.Error()
		}
		m.setState(bus, &State{Num: MonitorStopped, Status: s})
	}()
	return nil
}
func (m *MonitorAMI) Status() {
	a, _ := m.ami.AsyncAction("DongleShowDevices", gami.Params{"Device": m.Dev})
	<-a
}

func (m *MonitorAMI) Stop() {
	m.stop <- true
}
