package trigger

import (
	"sync"
	"time"

	"github.com/golang/glog"
)

type Trigger struct {
	timer         *time.Timer
	interval      time.Duration // In milliseconds
	measurer_chan chan int
	analyzer_chan chan int
	wg            *sync.WaitGroup
}

func NewTrigger(interval int, measurer_chan chan int, analyzer_chan chan int, wg *sync.WaitGroup) *Trigger {
	return &Trigger{
		interval:      time.Duration(interval),
		measurer_chan: measurer_chan,
		analyzer_chan: analyzer_chan,
		wg:            wg,
	}
}

func (t *Trigger) Start() error {
	for {
		t.timer = time.NewTimer(t.interval * time.Millisecond)
		select {
		case newInterval := <-t.analyzer_chan:
			t.timer.Stop()
			t.interval = time.Duration(newInterval)
			glog.Info("Old timer stopped. New interval: %s", newInterval)
		case time := <-t.timer.C:
			glog.Info("Tick at ", time)
			t.measurer_chan <- 1
		}
	}
}

func (t *Trigger) Stop() error {
	t.timer.Stop()
	t.measurer_chan <- 0
	t.analyzer_chan <- 0
	defer t.wg.Done()
	glog.Info("Timer stopped")
	return nil
}
