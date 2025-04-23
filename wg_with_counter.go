package main

import (
	"sync"
	"time"
)

type WGWithCounter struct {
	wg    sync.WaitGroup
	count int
	max   int
}

func newWGWC() *WGWithCounter {
	return &WGWithCounter{
		wg:    sync.WaitGroup{},
		count: 0,
		max:   wgMax,
	}
}

func (wg *WGWithCounter) Add(function func()) {
	wg.wg.Add(1)
	wg.count += 1

	if wg.count <= wg.max {
		go function()
	} else {
		for {
			if wg.count <= wg.max {
				go function()
				return
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func (wg *WGWithCounter) Done() {
	wg.wg.Done()
	wg.count--
}
