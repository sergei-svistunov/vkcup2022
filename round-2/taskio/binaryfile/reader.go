package binaryfile

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"round-2/pools"
	"round-2/taskio"
)

var _ taskio.Reader = &Reader{}

type Reader struct {
	filename string
	lastErr  error
}

func NewReader(filename string) *Reader {
	return &Reader{
		filename: filename,
	}
}

func (f *Reader) DataCh() <-chan int64 {
	ch := make(chan int64, taskio.DataChSize)

	go f.gen(ch)

	return ch
}

func (f *Reader) Err() error { return f.lastErr }

func (f *Reader) gen(ch chan<- int64) {
	ff, err := os.Open(f.filename)
	if err != nil {
		f.lastErr = err
		close(ch)
		return
	}
	defer ff.Close()

	bufR := pools.PoolBufReaders.Get().(*bufio.Reader)
	defer pools.PoolBufReaders.Put(bufR)
	bufR.Reset(ff)

	buf := [8]byte{}
	for {
		cnt, err := bufR.Read(buf[:])
		if err != nil {
			if !errors.Is(err, io.EOF) {
				f.lastErr = err
			}
			close(ch)
			return
		}
		if cnt != 8 {
			f.lastErr = fmt.Errorf("cannot read full number")
			close(ch)
			return
		}

		ch <- int64(binary.LittleEndian.Uint64(buf[:]))
	}
}
