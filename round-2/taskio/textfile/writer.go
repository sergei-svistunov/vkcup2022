package textfile

import (
	"bufio"
	"os"

	"round-2/fastconv"
	"round-2/pools"
	"round-2/taskio"
)

var _ taskio.Writer = &Writer{}

type Writer struct {
	filename string
}

func NewWriter(filename string) *Writer {
	return &Writer{
		filename: filename,
	}
}

func (f *Writer) WriteData(r taskio.Reader) error {
	ff, err := os.Create(f.filename)
	if err != nil {
		return err
	}
	defer ff.Close()

	bufW := pools.PoolBufWriters.Get().(*bufio.Writer)
	defer pools.PoolBufWriters.Put(bufW)
	defer bufW.Flush()
	bufW.Reset(ff)

	buf := [32]byte{}

	for n := range r.DataCh() {
		if _, err := bufW.Write(fastconv.Itoa(n, buf[:])); err != nil {
			return err
		}

		if _, err := bufW.WriteRune('\n'); err != nil {
			return err
		}
	}

	return r.Err()
}
