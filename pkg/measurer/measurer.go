package measurer

import (
	"strconv"
	"sync"
	"time"

	"github.com/golang/glog"

	"github.com/hydro-monitor/node/pkg/server"
)

type Measurer struct {
	trigger_chan  chan int
	analyzer_chan chan float64
	stop_chan     chan int
	wg            *sync.WaitGroup
	comm          *ArduinoCommunicator
}

func NewMeasurer(trigger_chan chan int, analyzer_chan chan float64, wg *sync.WaitGroup) *Measurer {
	return &Measurer{
		trigger_chan:  trigger_chan,
		analyzer_chan: analyzer_chan,
		stop_chan:     make(chan int),
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
	glog.Infof("Sending measurement %f to server", f)
	err = server.PostNodeMeasurement(server.APIMeasurement{
		Time:       time.Now(),
		WaterLevel: f,
	})
	if err != nil {
		glog.Errorf("Error sending measurement %f to server: %v", f, err)
	}
}

func (m *Measurer) Start() error {
	defer m.wg.Done()
	for {
		select {
		case <-m.trigger_chan:
			glog.Info("Received alert from Trigger. Requesting measurement")
			m.takeMeasurement()
		case <-m.stop_chan:
			glog.Info("Received stop sign")
			return nil
		}
	}
}

func (m *Measurer) Stop() error {
	glog.Info("Sending stop sign")
	m.stop_chan <- 1
	return nil
}
