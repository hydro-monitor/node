package measurer

import (
	"sync"
	"time"

	"github.com/golang/glog"

	"github.com/hydro-monitor/node/pkg/camera"
	"github.com/hydro-monitor/node/pkg/server"
	"github.com/hydro-monitor/node/pkg/water"
)

type Measurer struct {
	trigger_chan  chan int
	manual_chan   chan int
	analyzer_chan chan float64
	stop_chan     chan int
	wg            *sync.WaitGroup
	waterLevel    *water.WaterLevel
}

func NewMeasurer(trigger_chan, manual_chan chan int, analyzer_chan chan float64, wg *sync.WaitGroup) *Measurer {
	return &Measurer{
		trigger_chan:  trigger_chan,
		analyzer_chan: analyzer_chan,
		manual_chan:   manual_chan,
		stop_chan:     make(chan int),
		wg:            wg,
		waterLevel:    water.NewWaterLevel(),
	}
}

func (m *Measurer) takePicture(time time.Time) (string, error) {
	c := camera.Camera{}
	fileName, err := c.TakeStill(time.String())
	return fileName, err
}

func (m *Measurer) takeWaterLevelMeasurement() float64 {
	f, _ := m.waterLevel.TakeWaterLevel()

	glog.Infof("Sending measurement %f to analyzer", f)
	select {
	case m.analyzer_chan <- f:
		glog.Info("Measurement sent")
	case <-time.After(10 * time.Second):
		glog.Warning("Measurement send timed out")
	}

	return f
}

func (m *Measurer) takeMeasurement(manual bool) {
	time := time.Now()

	glog.Info("Taking water level")
	waterLevel := m.takeWaterLevelMeasurement()

	glog.Infof("Sending measurement (water level: %f and picture) to server", waterLevel)
	measurementID, err := server.PostNodeMeasurement(server.APIMeasurement{
		Time:       time,
		WaterLevel: waterLevel,
		WasManual:  manual,
	})
	if err != nil {
		glog.Errorf("Error sending measurement %f to server: %v. Skipping measurement", waterLevel, err)
		//FIXME What to do here? skip?
		return
	}

	glog.Info("Taking picture")
	go func() {
		pictureFile, err := m.takePicture(time)
		if err != nil {
			glog.Errorf("Error taking picture: %v. Skipping measurement", err)
			return
		}

		if err := server.PostNodePicture(server.APIPicture{
			MeasurementID: *measurementID,
			Picture:       pictureFile,
			PictureNumber: 1, // TODO Pending implementation of multiple pictures per measurement
		}); err != nil {
			glog.Errorf("Error sending picture to server: %v", err)
			return
		}
	}()
}

func (m *Measurer) Start() error {
	defer m.wg.Done()
	for {
		select {
		case <-m.trigger_chan:
			glog.Info("Received alert from Trigger. Requesting measurement")
			m.takeMeasurement(false)
		case <-m.manual_chan:
			glog.Info("Received alert from ManualTrigger. Requesting measurement")
			m.takeMeasurement(true)
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
