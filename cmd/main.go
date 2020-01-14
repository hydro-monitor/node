package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang/glog"

	"github.com/hydro-monitor/node/pkg/analyzer"
	"github.com/hydro-monitor/node/pkg/config"
	"github.com/hydro-monitor/node/pkg/measurer"
	"github.com/hydro-monitor/node/pkg/trigger"
)

const (
	interval                    = 500
	configurationUpdateInterval = 60
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
		t:  trigger.NewTrigger(interval, triggerMeasurer, triggerAnalyzer, wg),
		m:  measurer.NewMeasurer(triggerMeasurer, measurerAnalyzer, wg),
		a:  analyzer.NewAnalyzer(measurerAnalyzer, triggerAnalyzer, configAnalyzer, wg),
		cw: config.NewConfigWatcher(triggerConfig, configAnalyzer, configurationUpdateInterval, wg),
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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	glog.Info("Awaiting signal")
	sig := <-sigs
	glog.Infof("Signal received: %v. Stopping workers", sig)

	n.t.Stop()
	n.m.Stop()
	n.a.Stop()
	n.cw.Stop()

	wg.Wait()
}
