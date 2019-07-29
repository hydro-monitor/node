package main

import (
	"flag"
	"sync"
	"time"

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
}

func NewNode(c chan int, wg *sync.WaitGroup) *Node {
	return &Node{
		t: trigger.NewTrigger(INTERVAL, c, wg),
		m: measurer.NewMeasurer(c, wg),
	}
}

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	wg.Add(2)
	c := make(chan int)
	n := NewNode(c, &wg)

	go n.m.Start()
	n.t.Start()

	time.Sleep(2000 * time.Millisecond)

	n.t.Stop()
	wg.Wait()
}
