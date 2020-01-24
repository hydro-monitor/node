package water

import (
	"strconv"
	"strings"

	"github.com/golang/glog"
)

type WaterLevel struct {
	comm *ArduinoCommunicator
}

func NewWaterLevel() *WaterLevel {
	return &WaterLevel{
		comm: NewArduinoCommunicator(),
	}
}

func (w *WaterLevel) TakeWaterLevel() (float64, error) {
	if err := w.comm.RequestMeasurement(); err != nil {
		glog.Errorf("Error requesting measurement to Arduino %v", err)
		return -1, err
	}

	buffer := make([]byte, 128)
	n, err := w.comm.ReadMeasurement(buffer)
	if err != nil {
		glog.Errorf("Error reading measurement from Arduino %v", err)
		return -1, err
	}
	/*
		buffer := make([]byte, 128) FIXME Remove mock measurement
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
		return -1, err
	}

	return f, nil
}
