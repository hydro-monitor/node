package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"

	"github.com/hydro-monitor/node/pkg/server"
)

// Map with all posible states in the node configuration.
type Configutation struct {
	stateNames []string
	states     map[string]server.State
}

func NewConfiguration(states map[string]server.State) (c *Configutation) {
	stateNames := []string{}
	for k := range states {
		stateNames = append(stateNames, k)
	}

	return &Configutation{
		stateNames: stateNames,
		states:     states,
	}
}

func (c *Configutation) GetStates() []string {
	return c.stateNames
}

func (c *Configutation) GetState(stateName string) server.State {
	return c.states[stateName]
}

// Continuously polls the servers for the right configuration of the node
type ConfigWatcher struct {
	wg            *sync.WaitGroup
	trigger_chan  chan int
	analyzer_chan chan *Configutation
	stop_chan     chan int
	interval      time.Duration // In milliseconds
}

func NewConfigWatcher(trigger_chan chan int, analyzer_chan chan *Configutation, interval int, wg *sync.WaitGroup) *ConfigWatcher {
	c := &ConfigWatcher{
		wg:            wg,
		trigger_chan:  trigger_chan,
		analyzer_chan: analyzer_chan,
		stop_chan:     make(chan int),
		interval:      time.Duration(interval),
	}
	return c
}

func (c *ConfigWatcher) updateConfiguration() error {
	serverConfig, err := server.GetNodeConfiguration()
	if err != nil {
		glog.Errorf("Could not get configuration from server: %v", err)
		return err
	}
	config := NewConfiguration(serverConfig.States)
	glog.Infof("Sending new node configuration: %v", config)
	select {
	case c.analyzer_chan <- config:
		glog.Info("Configuration update sent")
		return nil
	case <-time.After(10 * time.Second):
		glog.Info("Configuration update timed out")
		return fmt.Errorf("Configuration update timed out")
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
		case <-c.stop_chan:
			glog.Info("Received stop from sign")
			timer.Stop()
			return nil
		}
	}
}

func (c *ConfigWatcher) Stop() error {
	glog.Info("Sending stop sign")
	c.stop_chan <- 1
	return nil
}
