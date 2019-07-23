package main

import (
	"time"
)

const (
	INTERVAL = 500
)

type Node struct {
	trigger *Trigger
}

func NewNode() *Node {
	return &Node{
		trigger: NewTrigger(INTERVAL),
	}
}

func main() {
	n := NewNode()
	n.trigger.start()
	time.Sleep(2000 * time.Millisecond)
	n.trigger.stop()
}
