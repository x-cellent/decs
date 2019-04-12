package decs

import (
	"sync/atomic"
	"time"
)

type commandQueue struct {
	next int32
	curr int32
	q    chan *command
}

func newCommandQueue() *commandQueue {
	return &commandQueue{
		q: make(chan *command),
	}
}

func (cq *commandQueue) queue(cmd *command) {
	idx := atomic.AddInt32(&cq.next, 1) - 1
	go func() {
		for atomic.LoadInt32(&cq.curr) != idx {
			time.Sleep(100 * time.Nanosecond)
		}
		cq.q <- cmd
		atomic.AddInt32(&cq.curr, 1)
	}()
}
