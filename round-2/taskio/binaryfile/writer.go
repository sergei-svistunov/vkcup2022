package binaryfile

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"

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

	buf := [8]byte{}

	for n := range r.DataCh() {
		binary.LittleEndian.PutUint64(buf[:], uint64(n))
		if _, err := bufW.Write(buf[:]); err != nil {
			return fmt.Errorf("cannot write sorted file: %w", err)
		}
	}

	return r.Err()
}
