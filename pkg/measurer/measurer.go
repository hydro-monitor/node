package measurer

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
)

type Measurer struct {
	channel chan int
	wg      *sync.WaitGroup
}

func NewMeasurer(channel chan int, wg *sync.WaitGroup) *Measurer {
	return &Measurer{
		channel: channel,
		wg:      wg,
	}
}

func (m *Measurer) Start() error {
	defer m.wg.Done()
	for {
		select {
		case data := <-m.channel:
			if data == 1 {
				glog.Info("Received alert from Trigger")
			} else if data == 0 {
				glog.Info("Received stop from Trigger")
				return nil
			} else {
				glog.Errorf("Did not recognize data sent though chan: %v", data)
				return fmt.Errorf("Did not recognize data sent though chan: %v", data)
			}
		}
	}
}
