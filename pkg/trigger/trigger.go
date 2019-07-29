package trigger

import (
	"sync"
	"time"

	"github.com/golang/glog"
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
			glog.Info("Tick at", time)
			t.channel <- 1
		}
	}()
	return nil
}

func (t *Trigger) Stop() error {
	t.ticker.Stop()
	t.channel <- 0
	defer t.wg.Done()
	glog.Info("Ticker stopped")
	return nil
}
