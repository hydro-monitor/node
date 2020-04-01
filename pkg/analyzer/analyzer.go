package analyzer

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/hydro-monitor/node/pkg/config"
	"github.com/hydro-monitor/node/pkg/envconfig"
)

type Analyzer struct {
	wg                    *sync.WaitGroup
	trigger_chan          chan int
	measurer_chan         chan float64
	config_watcher_chan   chan *config.Configutation
	stop_chan             chan int
	config                *config.Configutation
	currentState          string
	intervalUpdateTimeout time.Duration
}

func (a *Analyzer) updateConfiguration(newConfig *config.Configutation) error {
	glog.Info("Saving node configuration")
	a.config = newConfig
	return nil
}

func NewAnalyzer(measurer_chan chan float64, trigger_chan chan int, config_watcher_chan chan *config.Configutation, wg *sync.WaitGroup) *Analyzer {
	env := envconfig.New()
	a := &Analyzer{
		wg:                    wg,
		trigger_chan:          trigger_chan,
		measurer_chan:         measurer_chan,
		config_watcher_chan:   config_watcher_chan,
		stop_chan:             make(chan int),
		intervalUpdateTimeout: time.Duration(env.IntervalUpdateTimeout) * time.Second,
	}
	return a
}

func (a *Analyzer) lookForCurrentState(measurement float64) (string, error) {
	for _, stateName := range a.config.GetStates() {
		// TODO Deal with measurements equal to limits
		if measurement > a.config.GetState(stateName).LowerLimit && measurement < a.config.GetState(stateName).UpperLimit {
			return stateName, nil
		}
	}
	glog.Errorf("Could not found current state for measurement %f", measurement)
	return "", fmt.Errorf("Could not found current state for measurement %f", measurement)
}

func (a *Analyzer) updateCurrentState(newStateName string) {
	glog.Infof("New current state is %s", newStateName)
	a.currentState = newStateName
	newInterval := a.config.GetState(newStateName).Interval
	glog.Infof("Sending new current interval (%d) to Trigger", newInterval)
	select {
	case a.trigger_chan <- newInterval:
		glog.Info("Interval update sent")
	case <-time.After(a.intervalUpdateTimeout):
		glog.Warning("Interval update timed out")
	}
}

func (a *Analyzer) analyze(measurement float64) {
	glog.Info("Analyzing measurement")
	if a.currentState == "" {
		glog.Info("Current state unset. Setting current state")
		if currentStateName, err := a.lookForCurrentState(measurement); err != nil {
			glog.Info("Current state not found, skipping analysis")
			return
		} else {
			a.updateCurrentState(currentStateName)
		}
	}
	if measurement > a.config.GetState(a.currentState).UpperLimit { // TODO Deal with measurements equal to limits
		glog.Info("Upper limit surpassed")
		a.updateCurrentState(a.config.GetState(a.currentState).Next)
	} else if measurement < a.config.GetState(a.currentState).LowerLimit {
		glog.Info("Lower limit surpassed")
		a.updateCurrentState(a.config.GetState(a.currentState).Prev)
	}
}

func (a *Analyzer) Start() error {
	defer a.wg.Done()
	for {
		select {
		case configuration := <-a.config_watcher_chan:
			glog.Infof("Configuration received: %v", configuration)
			a.updateConfiguration(configuration)
		case measurement := <-a.measurer_chan:
			glog.Infof("Measurement received: %f", measurement)
			if a.config == nil {
				glog.Info("Node configuration not loaded, skipping analysis")
				continue
			}
			a.analyze(measurement)
		case <-a.stop_chan:
			glog.Info("Received stop sign")
			return nil
		}
	}
}

func (a *Analyzer) Stop() error {
	glog.Info("Sending stop sign")
	a.stop_chan <- 1
	return nil
}
