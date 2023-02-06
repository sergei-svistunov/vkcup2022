package slice

import (
	"round-2/taskio"
)

var _ taskio.Reader = Reader{}

type Reader []int64

func (f Reader) DataCh() <-chan int64 {
	ch := make(chan int64, taskio.DataChSize)

	go f.gen(ch)

	return ch
}

func (f Reader) Err() error { return nil }

func (f Reader) gen(ch chan<- int64) {
	defer close(ch)

	for _, n := range f {
		ch <- n
	}
}
