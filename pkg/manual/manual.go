package manual

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/hydro-monitor/node/pkg/server"
)

type ManualMeasurementTrigger struct {
	measurer_chan chan int
	stop_chan     chan int
	wg            *sync.WaitGroup
	timer         *time.Ticker
	interval      time.Duration // In seconds
}

func NewManualMeasurementTrigger(measurer_chan chan int, interval int, wg *sync.WaitGroup) *ManualMeasurementTrigger {
	return &ManualMeasurementTrigger{
		measurer_chan: measurer_chan,
		stop_chan:     make(chan int),
		wg:            wg,
		interval:      time.Duration(interval),
	}
}

func (m *ManualMeasurementTrigger) sendManualMeasurementRequestIfAny() error {
	pending, err := server.GetManualMeasurementRequest()
	if err != nil {
		glog.Errorf("Could not get manual measurement request from server: %v", err)
		return err
	}
	if pending {
		glog.Info("Sending manual measurement request to Measurer")
		select {
		case m.measurer_chan <- 1:
			glog.Info("Manual measurement request sent")
			return nil
		case <-time.After(10 * time.Second):
			glog.Warning("Manual measurement request send timed out")
			return fmt.Errorf("Manual measurement request send timed out")
		}
	}
	glog.Info("No manual measurement requests pending")
	return nil
}

func (m *ManualMeasurementTrigger) Start() error {
	defer m.wg.Done()
	m.timer = time.NewTicker(m.interval * time.Second)
	for {
		select {
		case time := <-m.timer.C:
			glog.Infof("Tick at %v. Quering server for manual measurement request.", time)
			m.sendManualMeasurementRequestIfAny()
		case <-m.stop_chan:
			glog.Info("Received stop sign")
			return nil
		}
	}
}

func (m *ManualMeasurementTrigger) Stop() error {
	m.timer.Stop()
	glog.Info("Timer stopped")
	glog.Info("Sending stop sign")
	m.stop_chan <- 1
	return nil
}
