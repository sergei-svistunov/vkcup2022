package heapfile

import (
	"bufio"
	"container/heap"
	"encoding/binary"
	"fmt"
	"os"

	"round-2/pools"
	"round-2/taskio"
)

var _ taskio.Writer = &HeapFile{}

type HeapFile struct {
	filename string
	memLimit int
}

func NewWriter(filename string, memLimit int) *HeapFile {
	return &HeapFile{
		filename: filename,
		memLimit: memLimit,
	}
}

func (f *HeapFile) WriteData(r taskio.Reader) error {
	ff, err := os.Create(f.filename)
	if err != nil {
		return err
	}
	defer ff.Close()

	bufW := pools.PoolBufWriters.Get().(*bufio.Writer)
	defer pools.PoolBufWriters.Put(bufW)
	defer bufW.Flush()
	bufW.Reset(ff)

	chunkSize := f.memLimit / 8
	data := make(Int64Heap, 0, chunkSize)

	for n := range r.DataCh() {
		data = append(data, n)

		if len(data) == chunkSize {
			if err := writeChunk(bufW, data); err != nil {
				return err
			}
			data = data[:0]
		}
	}

	if len(data) > 0 {
		if err := writeChunk(bufW, data); err != nil {
			return err
		}
	}

	return r.Err()
}

func writeChunk(w *bufio.Writer, data Int64Heap) error {
	heap.Init(&data)

	buf := [8]byte{}

	for _, n := range data {
		binary.LittleEndian.PutUint64(buf[:], uint64(n))
		if _, err := w.Write(buf[:]); err != nil {
			return fmt.Errorf("cannot write sorted file: %w", err)
		}
	}

	return nil
}
