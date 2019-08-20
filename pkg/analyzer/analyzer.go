package analyzer

import (
	"sync"

	"github.com/golang/glog"
)

const (
	PREVIOUS_STATE = -1
	NEXT_STATE     = 1
)

type Analyzer struct {
	wg            *sync.WaitGroup
	trigger_chan  chan int
	measurer_chan chan int
}

func NewAnalyzer(measurer_chan chan int, trigger_chan chan int, wg *sync.WaitGroup) *Analyzer {
	return &Analyzer{
		wg:            wg,
		trigger_chan:  trigger_chan,
		measurer_chan: measurer_chan,
	}
}

func (a *Analyzer) analyze(measurement int) {
	glog.Info("Analyzing measurement")
	// TODO Get limits from shmem
	upper_limit := 200
	lower_limit := 50

	if measurement > upper_limit {
		glog.Info("Upper limit surpassed")
		// Change state to next one
		a.trigger_chan <- NEXT_STATE
	} else if measurement < lower_limit {
		glog.Info("Lower limit surpassed")
		// Change state to previous one
		a.trigger_chan <- PREVIOUS_STATE
	}
}

func (a *Analyzer) Start() error {
	defer a.wg.Done()
	for {
		select {
		case measurement := <-a.measurer_chan:
			glog.Infof("Measurement received: %d", measurement)
			a.analyze(measurement)
		case <-a.trigger_chan:
			glog.Info("Received stop from Trigger")
			return nil
		}
	}
}
