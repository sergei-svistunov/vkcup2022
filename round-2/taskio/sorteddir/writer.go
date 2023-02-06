package sorteddir

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"

	"round-2/taskio"
	"round-2/taskio/binaryfile"
	"round-2/taskio/slice"
)

var _ taskio.Writer = &Writer{}

type Writer struct {
	path      string
	workers   int
	memLimit  int
	lastErr   error
	nextChunk int32
	filesMtx  sync.Mutex
}

func NewWriter(path string, workers, memLimit int) *Writer {
	return &Writer{
		path:     path,
		workers:  workers,
		memLimit: memLimit,
	}
}

func (d *Writer) WriteData(r taskio.Reader) error {
	if err := os.RemoveAll(d.path); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(d.path, 0700); err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(d.workers)
	ch := r.DataCh()

	for i := 0; i < d.workers; i++ {
		go d.consumeNums(ch, &wg)
	}
	wg.Wait()

	if err := r.Err(); err != nil {
		return err
	}

	if d.lastErr != nil {
		return d.lastErr
	}

	return nil
}

func (d *Writer) consumeNums(ch <-chan int64, wg *sync.WaitGroup) {
	defer wg.Done()

	chunkSize := d.memLimit / d.workers / 8
	data := make([]int64, 0, chunkSize)

	for n := range ch {
		data = append(data, n)

		if len(data) == chunkSize {
			if err := d.addChunk(data); err != nil {
				d.lastErr = err
			}
			data = data[:0]
		}
	}

	if len(data) > 0 {
		if err := d.addChunk(data); err != nil {
			d.lastErr = err
		}
	}
}

func (d *Writer) addChunk(data []int64) error {
	chunkId := atomic.AddInt32(&d.nextChunk, 1)

	sort.Slice(data, func(i, j int) bool {
		return data[i] < data[j]
	})

	d.filesMtx.Lock()
	defer d.filesMtx.Unlock()

	return binaryfile.NewWriter(filepath.Join(d.path, strconv.Itoa(int(chunkId)))).
		WriteData(slice.Reader(data))
}
