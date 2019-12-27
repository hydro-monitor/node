package config

import (
	"math"
	"sync"
	"time"

	"github.com/golang/glog"
)

// estados(ID nodo (text),
//         nombre (text),
//         cantidad de fotos a tomar por medición (int),
//         cada cuantos ms tiempo toma medición (int),
//         límite de nivel de agua para pasar al estado anterior (float),
//         límite de nivel de agua para pasar al estado siguiente (float),
//         nombre estado anterior (text),
//         nombre estado siguiente (text))
type State struct {
	Name        string
	Interval    int
	UpperLimit  float64
	LowerLimit  float64
	PicturesNum int
	Next        string // State name (key)
	Prev        string // State name (key)
}

// Map with all posible states in the node configuration.
type Configutation struct {
	stateNames []string
	states     map[string]State
}

func (c *Configutation) GetStates() []string {
	return c.stateNames
}

func (c *Configutation) GetState(stateName string) State {
	return c.states[stateName]
}

// Continuously polls the servers for the right configuration of the node
type ConfigWatcher struct {
	wg            *sync.WaitGroup
	trigger_chan  chan int
	analyzer_chan chan *Configutation
	interval      time.Duration // In milliseconds
}

func NewConfigWatcher(trigger_chan chan int, analyzer_chan chan *Configutation, interval int, wg *sync.WaitGroup) *ConfigWatcher {
	c := &ConfigWatcher{
		wg:            wg,
		trigger_chan:  trigger_chan,
		analyzer_chan: analyzer_chan,
		interval:      time.Duration(interval),
	}
	return c
}

func (c *ConfigWatcher) updateConfiguration() {
	// TODO Add query for node config to the server
	states := map[string]State{
		"Normal": State{
			Name:        "Normal",
			Interval:    60,
			PicturesNum: 0,
			UpperLimit:  math.Inf(1),
			LowerLimit:  math.Inf(-1),
		},
	}
	glog.Info("Sending new node configuration")
	config := &Configutation{
		stateNames: []string{"Normal"},
		states:     states,
	}
	select {
	case c.analyzer_chan <- config:
		glog.Info("Configuration update sent")
	case <-time.After(10 * time.Second):
		glog.Info("Configuration update timed out")
	}
}

func (c *ConfigWatcher) Start() error {
	defer c.wg.Done()
	timer := time.NewTicker(c.interval * time.Millisecond)
	for {
		select {
		case time := <-timer.C:
			glog.Infof("Tick at %v. Quering server for node configuration.", time)
			c.updateConfiguration()
		case <-c.trigger_chan:
			glog.Info("Received stop from Trigger")
			timer.Stop()
			return nil
		}
	}
}
