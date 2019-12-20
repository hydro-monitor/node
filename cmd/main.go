package main

import (
	"flag"
	"sync"
	"time"

	"github.com/hydro-monitor/node/pkg/analyzer"
	"github.com/hydro-monitor/node/pkg/config"
	"github.com/hydro-monitor/node/pkg/measurer"
	"github.com/hydro-monitor/node/pkg/trigger"
)

const (
	INTERVAL = 500
)

func init() {
	flag.Set("logtostderr", "true")
}

type node struct {
	t  *trigger.Trigger
	m  *measurer.Measurer
	a  *analyzer.Analyzer
	cw *config.ConfigWatcher
}

// NewNode creates a new node with all it's correspondant processes
func newNode(triggerMeasurer, triggerAnalyzer, triggerConfig chan int, measurerAnalyzer chan float64, configAnalyzer chan *config.Configutation, wg *sync.WaitGroup) *node {
	return &node{
		t:  trigger.NewTrigger(INTERVAL, triggerMeasurer, triggerAnalyzer, wg),
		m:  measurer.NewMeasurer(triggerMeasurer, measurerAnalyzer, wg),
		a:  analyzer.NewAnalyzer(measurerAnalyzer, triggerAnalyzer, configAnalyzer, wg),
		cw: config.NewConfigWatcher(triggerConfig, configAnalyzer, INTERVAL, wg),
	}
}

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	wg.Add(4)
	triggerMeasurer := make(chan int)
	triggerAnalyzer := make(chan int)
	measurerAnalyzer := make(chan float64)
	triggerConfig := make(chan int)
	configAnalyzer := make(chan *config.Configutation)
	n := newNode(triggerMeasurer, triggerAnalyzer, triggerConfig, measurerAnalyzer, configAnalyzer, &wg)

	go n.a.Start()
	go n.m.Start()
	go n.t.Start()
	go n.cw.Start()

	time.Sleep(2000 * time.Millisecond)

	n.t.Stop(triggerConfig)
	wg.Wait()
}
