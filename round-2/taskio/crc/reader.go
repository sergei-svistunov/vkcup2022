package crc

import (
	"round-2/taskio"
)

var _ taskio.Reader = &Reader{}

type Reader struct {
	r       taskio.Reader
	sum     int64
	lastErr error
}

func NewReader(r taskio.Reader) *Reader {
	return &Reader{
		r: r,
	}
}

func (c *Reader) DataCh() <-chan int64 {
	ch := make(chan int64, taskio.DataChSize)

	go c.gen(ch)

	return ch
}

func (c *Reader) Err() error { return c.lastErr }
func (c *Reader) Sum() int64 { return c.sum }

func (c *Reader) gen(ch chan<- int64) {
	defer close(ch)

	for n := range c.r.DataCh() {
		c.sum += n
		ch <- n
	}

	c.lastErr = c.r.Err()
}
