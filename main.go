package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	cfgfile := flag.String("config", "config.ini", "config file path")
	flag.Parse()

	cfg, err := NewGateConfig(*cfgfile)
	if err != nil {
		log.Fatal("Error parsing config: ", err)
	}

	bus := &MonitorConn{
		State: make(chan State),
		Sms:   make(chan SMS),
	}
	monitors := make(map[string]Monitor)

	for n, c := range cfg.Dev_AMI {
		log.Println("Monitor: creating", n)
		c.Name = n
		mon, err := NewMonitorAMI(c)
		if err != nil {
			log.Fatal(err)
		}
		monitors[n] = mon
	}

	for n, mon := range monitors {
		log.Println("Monitor: starting", n)
		err := mon.Start(bus)
		if err != nil {
			log.Fatal(err)
		}
	}

	for exit := false; exit == false; {
		select {
		case st := <-bus.State:
			log.Println("Event:", st.Name, st)
			dev := monitors[st.Name]
			switch st.Num {
			case MonitorStopped:
				if st.Status == "" {
					exit = true
					break
				}
				time.Sleep(4 * time.Second)

				err := dev.Start(bus)
				if err != nil {
					log.Fatal(err)
				}
				/*case MonitorRunning:
					log.Println("Event:", st.name, st)
				case MonitorStarting:
					log.Println("Event:", st.name, st)*/
			}
		case sms := <-bus.Sms:
			log.Println("SMS:", sms)
		case <-time.After(15 * time.Second):
			dev := monitors["stas0"]
			if dev.GetState().Num == MonitorRunning {
				dev.Status()
				//dev.Stop()
			}
		}
	}
	log.Println("Monitor: Exiting")
}
