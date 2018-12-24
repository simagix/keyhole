// Copyright 2018 Kuei-chun Chen. All rights reserved.

package util

import (
	"sync"
)

// WaitGroup -
type WaitGroup struct {
	channel   chan string
	waitGroup sync.WaitGroup
}

// NewWaitGroup -
func NewWaitGroup(size int) *WaitGroup {
	return &WaitGroup{channel: make(chan string, size), waitGroup: sync.WaitGroup{}}
}

// Add -
func (wg *WaitGroup) Add(num int) {
	wg.waitGroup.Add(num)
	select {
	case wg.channel <- "done":
		break
	}
}

// Done -
func (wg *WaitGroup) Done() {
	<-wg.channel
	wg.waitGroup.Done()
}

// Wait -
func (wg *WaitGroup) Wait() {
	wg.waitGroup.Wait()
}
