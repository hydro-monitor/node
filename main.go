package main

import (
	"sync"
	"time"
)

const (
	INTERVAL = 500
)

type Node struct {
	trigger  *Trigger
	measurer *Measurer
}

func NewNode(c chan int, wg *sync.WaitGroup) *Node {
	return &Node{
		trigger:  NewTrigger(INTERVAL, c, wg),
		measurer: NewMeasurer(c, wg),
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	c := make(chan int)
	n := NewNode(c, &wg)

	go n.measurer.start()
	n.trigger.start()

	time.Sleep(2000 * time.Millisecond)

	n.trigger.stop()
	wg.Wait()
}
