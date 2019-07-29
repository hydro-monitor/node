package trigger

import (
	"fmt"
	"sync"
	"time"
)

type Trigger struct {
	ticker   *time.Ticker
	interval time.Duration // In milliseconds
	channel  chan int
	wg       *sync.WaitGroup
}

func NewTrigger(interval int, channel chan int, wg *sync.WaitGroup) *Trigger {
	return &Trigger{
		interval: time.Duration(interval),
		channel:  channel,
		wg:       wg,
	}
}

func (t *Trigger) Start() error {
	t.ticker = time.NewTicker(t.interval * time.Millisecond)
	go func() {
		for time := range t.ticker.C {
			fmt.Println("Tick at", time)
			t.channel <- 1
		}
	}()
	return nil
}

func (t *Trigger) Stop() error {
	t.ticker.Stop()
	t.channel <- 0
	defer t.wg.Done()
	fmt.Println("Ticker stopped")
	return nil
}
