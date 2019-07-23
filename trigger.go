package main

import (
	"fmt"
	"time"
)

type Trigger struct {
	ticker   *time.Ticker
	interval time.Duration // In milliseconds
}

func NewTrigger(interval int) *Trigger {
	return &Trigger{
		interval: time.Duration(interval),
	}
}

func (t *Trigger) start() error {
	t.ticker = time.NewTicker(t.interval * time.Millisecond)
	go func() {
		for time := range t.ticker.C {
			fmt.Println("Tick at", time)
		}
	}()
	return nil
}

func (t *Trigger) stop() error {
	t.ticker.Stop()
	fmt.Println("Ticker stopped")
	return nil
}
