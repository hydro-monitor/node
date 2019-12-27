package trigger

import (
	"sync"
	"time"

	"github.com/golang/glog"
)

type Trigger struct {
	timer         *time.Ticker
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
	t.timer = time.NewTicker(t.interval * time.Millisecond)
	for {
		select {
		case newInterval := <-t.analyzer_chan:
			t.timer.Stop()
			t.interval = time.Duration(newInterval)
			glog.Infof("NEW INTERVAL RECEIVED, CREATING NEW TICKER WITH INTERVAL %d %v", newInterval, t.interval) // FIXME remove me
			t.timer = time.NewTicker(t.interval * time.Millisecond)
			glog.Infof("Old timer stopped. New interval: %d", newInterval)
		case time := <-t.timer.C:
			glog.Infof("Tick at %v. Awaking Measurer", time)
			t.measurer_chan <- 1
			glog.Infof("SENT awake to Measurer") // FIXME i never get to log
		}
	}
}

func (t *Trigger) Stop(config_watcher_chan chan int) error {
	t.timer.Stop()
	glog.Info("Timer stopped")
	t.measurer_chan <- 0
	t.analyzer_chan <- 0
	config_watcher_chan <- 0 // FIXME I don't belong here, delete me whenever this turns into an endless job
	defer t.wg.Done()
	return nil
}
