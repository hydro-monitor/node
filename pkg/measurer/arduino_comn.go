package measurer

import (
	"time"

	"github.com/golang/glog"
	"github.com/tarm/serial"
)

const (
	SERIAL = "/dev/cu.usbmodem1421"
	BAUD   = 9600
)

type ArduinoCommunicator struct {
	port *serial.Port
}

func NewArduinoCommunicator() *ArduinoCommunicator {
	c := &serial.Config{
		Name: SERIAL,
		Baud: BAUD,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		glog.Fatalf("Error opening serial port %v", err)
	}

	// We need to sleep the program for 2 seconds because every time a new
	// serial connection is made with the Arduino it resets similar to when
	// you are uploading your program to it.
	time.Sleep(2 * time.Second)

	return &ArduinoCommunicator{
		port: s,
	}
}

func (ac *ArduinoCommunicator) RequestMeasurement() error {
	req := []byte{1}
	// Write will block until at least one byte is written
	_, err := ac.port.Write(req)
	if err != nil {
		glog.Errorf("Error writing to serial port %v", err)
		return err
	}
	return nil
}

func (ac *ArduinoCommunicator) read(buffer []byte) (int, error) {
	n, err := ac.port.Read(buffer)
	if err != nil {
		glog.Errorf("Error reding from serial port %v", err)
		return n, err
	}
	return n, nil
}

func (ac *ArduinoCommunicator) ReadMeasurement(buffer []byte) (int, error) {
	// Read will block until at least one byte is returned
	n, err := ac.read(buffer)
	if err != nil {
		return n, err
	}
	glog.Infof("Data received is: %q", buffer[:n])

	for buffer[n-1] != '\n' {
		n_tmp, err := ac.read(buffer[n:])
		if err != nil {
			return n + n_tmp, err
		}
		n = n + n_tmp
	}

	glog.Infof("Measurement received is: %q", buffer[:n])
	return n, nil
}

func (ac *ArduinoCommunicator) Close() error {
	if err := ac.port.Close(); err != nil {
		glog.Errorf("Error closing serial port %v", err)
		return err
	}
	return nil
}
