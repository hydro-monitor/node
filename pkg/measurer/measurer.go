package measurer

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/golang/glog"
)

type Measurer struct {
	trigger_chan  chan int
	analyzer_chan chan float64
	wg            *sync.WaitGroup
	comm          *ArduinoCommunicator
}

func NewMeasurer(trigger_chan chan int, analyzer_chan chan float64, wg *sync.WaitGroup) *Measurer {
	return &Measurer{
		trigger_chan:  trigger_chan,
		analyzer_chan: analyzer_chan,
		wg:            wg,
		//comm:          NewArduinoCommunicator(), FIXME
	}
}

func (m *Measurer) takeMeasurement() {
	/*
		if err := m.comm.RequestMeasurement(); err != nil {
			glog.Errorf("Error requesting measurement to Arduino %v", err)
		}

		buffer := make([]byte, 128)
		n, err := m.comm.ReadMeasurement(buffer)
		if err != nil {
			glog.Errorf("Error reading measurement from Arduino %v", err)
		}
	*/
	buffer := make([]byte, 128) // FIXME Delete me
	buffer[0] = '6'
	buffer[1] = '5'
	n := 2

	glog.Infof("Measurement received: %q", buffer[:n])
	s := string(buffer[:n])
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		glog.Errorf("Failed to convert string '%s' to int: %v", s, err)
	}
	glog.Infof("Sending measurement %f to analyzer", f)
	m.analyzer_chan <- f
}

func (m *Measurer) Start() error {
	defer m.wg.Done()
	for {
		select {
		case data := <-m.trigger_chan:
			if data == 1 {
				glog.Info("Received alert from Trigger. Requesting measurement")
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
