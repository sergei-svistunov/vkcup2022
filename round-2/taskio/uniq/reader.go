package uniq

import (
	"round-2/taskio"
)

var _ taskio.Reader = &Reader{}

type Reader struct {
	r       taskio.Reader
	lastErr error
}

func NewReader(r taskio.Reader) *Reader {
	return &Reader{
		r: r,
	}
}

func (u *Reader) DataCh() <-chan int64 {
	ch := make(chan int64, taskio.DataChSize)

	go u.gen(ch)

	return ch
}

func (u *Reader) Err() error { return u.lastErr }

func (u *Reader) gen(ch chan<- int64) {
	defer close(ch)

	var prevValue int64
	first := true

	for n := range u.r.DataCh() {
		if first {
			prevValue = n
			first = false
			ch <- n
		}

		if prevValue != n {
			prevValue = n
			ch <- n
		}
	}

	u.lastErr = u.r.Err()
}
