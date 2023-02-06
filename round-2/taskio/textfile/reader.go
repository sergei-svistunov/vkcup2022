package textfile

import (
	"bufio"
	"os"

	"round-2/fastconv"
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

	go f.gen(ch, func() {
		close(ch)
	})

	return ch
}

func (f *Reader) ReadTo(ch chan<- int64, done func()) {
	f.gen(ch, done)
}

func (f *Reader) Err() error { return f.lastErr }

func (f *Reader) gen(ch chan<- int64, done func()) {
	ff, err := os.Open(f.filename)
	if err != nil {
		f.lastErr = err
		close(ch)
		return
	}
	defer ff.Close()

	buf := pools.PoolBuf4k.Get().([]byte)
	defer pools.PoolBuf4k.Put(buf)

	scanner := bufio.NewScanner(ff)
	scanner.Buffer(buf, cap(buf))

	for {
		if !scanner.Scan() {
			f.lastErr = scanner.Err()
			done()
			return
		}

		n, err := fastconv.Atoi(scanner.Bytes())
		if err != nil {
			f.lastErr = err
			done()
			return
		}

		ch <- n
	}
}
