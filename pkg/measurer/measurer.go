package measurer

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
)

type Measurer struct {
	channel chan int
	wg      *sync.WaitGroup
	comm    *ArduinoCommunicator
}

func NewMeasurer(channel chan int, wg *sync.WaitGroup) *Measurer {
	return &Measurer{
		channel: channel,
		wg:      wg,
		comm:    NewArduinoCommunicator(),
	}
}

func (m *Measurer) takeMeasurement() {
	if err := m.comm.RequestMeasurement(); err != nil {
		glog.Errorf("Error requesting measurement to Arduino %v", err)
	}

	buffer := make([]byte, 128)
	n, err := m.comm.ReadMeasurement(buffer)
	if err != nil {
		glog.Errorf("Error reading measurement from Arduino %v", err)
	}
	glog.Infof("Measurement received: %q", buffer[:n])
}

func (m *Measurer) Start() error {
	defer m.wg.Done()
	for {
		select {
		case data := <-m.channel:
			if data == 1 {
				glog.Info("Received alert from Trigger")
				m.takeMeasurement()
			} else if data == 0 {
				glog.Info("Received stop from Trigger")
				return nil
			} else {
				glog.Errorf("Did not recognize data sent though chan: %v", data)
				return fmt.Errorf("Did not recognize data sent though chan: %v", data)
			}
		}
	}
}
