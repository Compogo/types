package counter

import (
	"sync/atomic"
)

type Counter interface {
	Set(value int64)
	Get() int64
	Inc()
	Dec()
	Reset()
}

type counter struct {
	value atomic.Int64
}

func NewCounter() Counter {
	return &counter{}
}

func (c *counter) Set(value int64) {
	c.value.Store(value)
}

func (c *counter) Get() int64 {
	return c.value.Load()
}

func (c *counter) Inc() {
	c.value.Add(1)
}

func (c *counter) Dec() {
	c.value.Add(-1)
}

func (c *counter) Reset() {
	c.value.Store(0)
}
