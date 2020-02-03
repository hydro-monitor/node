package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang/glog"
	"github.com/joho/godotenv"

	"github.com/hydro-monitor/node/pkg/analyzer"
	"github.com/hydro-monitor/node/pkg/config"
	"github.com/hydro-monitor/node/pkg/envconfig"
	"github.com/hydro-monitor/node/pkg/manual"
	"github.com/hydro-monitor/node/pkg/measurer"
	"github.com/hydro-monitor/node/pkg/trigger"
)

func init() {
	flag.Set("logtostderr", "true")

	// Loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		glog.Infof("No .env file found")
	}
}

type node struct {
	t  *trigger.Trigger
	m  *measurer.Measurer
	a  *analyzer.Analyzer
	cw *config.ConfigWatcher
	mt *manual.ManualMeasurementTrigger
}

// NewNode creates a new node with all it's correspondant processes
func newNode(triggerMeasurer, triggerAnalyzer, triggerConfig, manualMeasurer chan int, measurerAnalyzer chan float64, configAnalyzer chan *config.Configutation, wg *sync.WaitGroup) *node {
	env := envconfig.New()
	glog.Infof("Env config: %v", env)

	return &node{
		t:  trigger.NewTrigger(env.InitialTriggerInterval, triggerMeasurer, triggerAnalyzer, wg),
		m:  measurer.NewMeasurer(triggerMeasurer, manualMeasurer, measurerAnalyzer, wg),
		a:  analyzer.NewAnalyzer(measurerAnalyzer, triggerAnalyzer, configAnalyzer, wg),
		cw: config.NewConfigWatcher(triggerConfig, configAnalyzer, env.ConfigurationUpdateInterval, wg),
		mt: manual.NewManualMeasurementTrigger(manualMeasurer, env.ManualMeasurementPollInterval, wg),
	}
}

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	wg.Add(5)
	triggerMeasurer := make(chan int)
	triggerAnalyzer := make(chan int)
	measurerAnalyzer := make(chan float64)
	triggerConfig := make(chan int)
	manualMeasurer := make(chan int)
	configAnalyzer := make(chan *config.Configutation)
	n := newNode(triggerMeasurer, triggerAnalyzer, triggerConfig, manualMeasurer, measurerAnalyzer, configAnalyzer, &wg)

	go n.a.Start()
	go n.m.Start()
	go n.t.Start()
	go n.cw.Start()
	go n.mt.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	glog.Info("Awaiting signal")
	sig := <-sigs
	glog.Infof("Signal received: %v. Stopping workers", sig)

	n.t.Stop()
	n.m.Stop()
	n.a.Stop()
	n.cw.Stop()
	n.mt.Stop()

	wg.Wait()
}
