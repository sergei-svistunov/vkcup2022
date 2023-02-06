package rand

import (
	"math/rand"

	"round-2/taskio"
)

var _ taskio.Reader = Reader(0)

type Reader int

func (g Reader) DataCh() <-chan int64 {
	ch := make(chan int64, taskio.DataChSize)

	go g.gen(ch)

	return ch
}

func (g Reader) Err() error { return nil }

func (g Reader) gen(ch chan<- int64) {
	defer close(ch)

	for i := 0; i < int(g); i++ {
		ch <- rand.Int63() - rand.Int63()
	}
}
