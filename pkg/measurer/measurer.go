package measurer

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dhowden/raspicam"
	"github.com/golang/glog"

	"github.com/hydro-monitor/node/pkg/server"
)

const (
	picturesDir = "/home/pi/Documents/pictures"
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
		comm:          NewArduinoCommunicator(),
	}
}

func (m *Measurer) takePicture(time time.Time) (string, error) {
	fileName := time.String()
	file, err := os.Create(fmt.Sprintf("%s/%s", picturesDir, fileName))
	if err != nil {
		glog.Errorf("Error creating file for picture: %v", err)
		return "", err
	}
	defer file.Close()

	stillConfig := raspicam.NewStill()

	errCh := make(chan error)
	go func() {
		for x := range errCh {
			glog.Info(os.Stderr, "%v\n", x)
		}
	}()

	glog.Info("Capturing still with picamera")
	raspicam.Capture(stillConfig, file, errCh)

	return fileName, nil
}

func (m *Measurer) takeWaterLevelMeasurement() float64 {
	if err := m.comm.RequestMeasurement(); err != nil {
		glog.Errorf("Error requesting measurement to Arduino %v", err)
	}

	buffer := make([]byte, 128)
	n, err := m.comm.ReadMeasurement(buffer)
	if err != nil {
		glog.Errorf("Error reading measurement from Arduino %v", err)
	}
	/*
		buffer := make([]byte, 128)
		buffer[0] = '6'
		buffer[1] = '5'
		n := 2
	*/

	glog.Infof("Measurement received: %q", buffer[:n])
	str := string(buffer[:n])
	nStr := strings.TrimRight(str, "\r\n")
	f, err := strconv.ParseFloat(nStr, 64)
	if err != nil {
		glog.Errorf("Failed to convert string '%s' to int: %v", nStr, err)
	}
	glog.Infof("Sending measurement %f to analyzer", f)
	m.analyzer_chan <- f

	return f
}

func (m *Measurer) takeMeasurement() {
	time := time.Now()

	glog.Info("Taking water level")
	waterLevel := m.takeWaterLevelMeasurement()

	glog.Info("Taking picture")
	pictureFile, err := m.takePicture(time)
	if err != nil {
		glog.Errorf("Error taking picture: %v. Skipping measurement", err)
		return
	}

	glog.Infof("Sending measurement %f to server", waterLevel)
	err = server.PostNodeMeasurement(server.APIMeasurement{
		Time:       time,
		WaterLevel: waterLevel,
		Picture:    pictureFile,
	})
	if err != nil {
		glog.Errorf("Error sending measurement %f to server: %v", waterLevel, err)
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
