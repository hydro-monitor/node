package main

import (
	"fmt"
	"sync"
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

func (m *Measurer) start() error {
	defer m.wg.Done()
	for {
		select {
		case data := <-m.channel:
			if data == 1 {
				fmt.Println("Received alert from Trigger")
			} else if data == 0 {
				fmt.Println("Received stop from Trigger")
				return nil
			} else {
				return fmt.Errorf("Did not recognize data sent though chan: %v", data)
			}
		}
	}
}
