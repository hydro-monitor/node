package measurer

import (
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang/glog"

	"github.com/hydro-monitor/node/pkg/camera"
	"github.com/hydro-monitor/node/pkg/envconfig"
	"github.com/hydro-monitor/node/pkg/server"
	"github.com/hydro-monitor/node/pkg/water"
)

// Measurer represents a measurer
type Measurer struct {
	trigger_chan                     chan int
	manual_chan                      chan int
	analyzer_chan                    chan float64
	stop_chan                        chan int
	wg                               *sync.WaitGroup
	waterLevel                       *water.WaterLevel
	camera                           *camera.Camera
	server                           *server.Server
	measurementToAnalyzerSendTimeout time.Duration
}

// NewMeasurer creates and returns a new measurer
func NewMeasurer(trigger_chan, manual_chan chan int, analyzer_chan chan float64, wg *sync.WaitGroup) *Measurer {
	env := envconfig.New()
	return &Measurer{
		trigger_chan:                     trigger_chan,
		analyzer_chan:                    analyzer_chan,
		manual_chan:                      manual_chan,
		stop_chan:                        make(chan int),
		wg:                               wg,
		waterLevel:                       water.NewWaterLevel(),
		camera:                           camera.NewCamera(),
		server:                           server.NewServer(),
		measurementToAnalyzerSendTimeout: time.Duration(env.MeasurementToAnalyzerSendTimeout) * time.Second,
	}
}

// takePicture takes a new picture with camera. Uses time as picture name
func (m *Measurer) takePicture(time time.Time) (string, error) {
	fileName, err := m.camera.TakeStill(time.String())
	return fileName, err
}

// takeWaterLevelMeasurement takes water level with water level module
func (m *Measurer) takeWaterLevelMeasurement() float64 {
	f, _ := m.waterLevel.TakeWaterLevel()

	glog.Infof("Sending measurement %f to analyzer", f)
	select {
	case m.analyzer_chan <- f:
		glog.Info("Measurement sent")
	case <-time.After(m.measurementToAnalyzerSendTimeout):
		glog.Warning("Measurement send timed out")
	}

	return f
}

// takeMeasurement takes water level, sends water measurement to server. 
// Takes picture and uploads picture to new server measurement.
func (m *Measurer) takeMeasurement(manual bool) {
	measurementIDChan := make(chan *gocql.UUID)
	timeOfMeasurement := time.Now()

	go func(measurementIDChan chan *gocql.UUID, timeOfMeasurement time.Time, manual bool) {
		glog.Info("Taking water level")
		waterLevel := m.takeWaterLevelMeasurement()

		glog.Infof("Sending measurement (water level: %f and picture) to server", waterLevel)
		measurementID, err := m.server.PostNodeMeasurement(server.APIMeasurement{
			Time:       timeOfMeasurement,
			WaterLevel: waterLevel,
			ManualReading:  manual,
		})
		if err != nil {
			glog.Errorf("Error sending measurement %f to server: %v. Skipping measurement", waterLevel, err)
			measurementIDChan <- nil
			return
		}
		glog.Infof("Sending measurement ID %v to picture routine", measurementID)
		measurementIDChan <- measurementID
		glog.Infof("Measurement ID %v sent", measurementID)
	}(measurementIDChan, timeOfMeasurement, manual)

	go func(measurementIDChan chan *gocql.UUID, timeOfMeasurement time.Time) {
		glog.Info("Taking picture")
		pictureFile, err := m.takePicture(timeOfMeasurement)
		if err != nil {
			glog.Errorf("Error taking picture: %v. Skipping measurement", err)
			return // FIXME Esto va a provocar que se bloquee la rutina anterior porque nunca se saca el readingID del channel?
		}

		glog.Infof("Picture taken. Waiting for measurement ID from water level routine")
		measurementID := <-measurementIDChan
		glog.Infof("Measurement ID %v received", measurementID)
		if measurementID == nil {
			glog.Errorf("Measurement ID is nil, skipping picture upload")
			return
		}

		if err := m.server.PostNodePicture(server.APIPicture{
			MeasurementID: *measurementID,
			Picture:       pictureFile,
			PictureNumber: 1, // TODO Pending implementation of multiple pictures per measurement
		}); err != nil {
			glog.Errorf("Error sending picture to server: %v", err)
			return
		}
	}(measurementIDChan, timeOfMeasurement)
}

// Start starts measurer process. Exits when stop is received
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

// Stop stops measurer process
func (m *Measurer) Stop() error {
	glog.Info("Sending stop sign")
	m.stop_chan <- 1
	return nil
}
