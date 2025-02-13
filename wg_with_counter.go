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

func (wg *WGWithCounter) Add(delta int, function func()) {
	wg.wg.Add(delta)
	wg.count += delta

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
