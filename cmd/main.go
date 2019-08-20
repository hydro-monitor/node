package main

import (
	"flag"
	"sync"
	"time"

	"github.com/hydro-monitor/node/pkg/analyzer"
	"github.com/hydro-monitor/node/pkg/measurer"
	"github.com/hydro-monitor/node/pkg/trigger"
)

const (
	INTERVAL = 500
)

func init() {
	flag.Set("logtostderr", "true")
}

type Node struct {
	t *trigger.Trigger
	m *measurer.Measurer
	a *analyzer.Analyzer
}

func NewNode(trigger_measurer, trigger_analyzer, measurer_analyzer chan int, wg *sync.WaitGroup) *Node {
	return &Node{
		t: trigger.NewTrigger(INTERVAL, trigger_measurer, trigger_analyzer, wg),
		m: measurer.NewMeasurer(trigger_measurer, measurer_analyzer, wg),
		a: analyzer.NewAnalyzer(measurer_analyzer, trigger_analyzer, wg),
	}
}

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	wg.Add(2)
	trigger_measurer := make(chan int)
	trigger_analyzer := make(chan int)
	measurer_analyzer := make(chan int)
	n := NewNode(trigger_measurer, trigger_analyzer, measurer_analyzer, &wg)

	go n.a.Start()
	go n.m.Start()
	n.t.Start()

	time.Sleep(2000 * time.Millisecond)

	n.t.Stop()
	wg.Wait()
}
