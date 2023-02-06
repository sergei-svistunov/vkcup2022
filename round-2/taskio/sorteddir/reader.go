package sorteddir

import (
	"container/heap"
	"io/fs"
	"path/filepath"

	"round-2/taskio"
	"round-2/taskio/binaryfile"
)

var _ taskio.Reader = &Reader{}

type Reader struct {
	path    string
	lastErr error
}

type filesHeap []*filesHeapItem

func (s filesHeap) Len() int { return len(s) }

func (s filesHeap) Less(i, j int) bool { return s[i].curVal < s[j].curVal }
func (s filesHeap) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s *filesHeap) Push(x any)        { *s = append(*s, x.(*filesHeapItem)) }
func (s *filesHeap) Pop() any {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type filesHeapItem struct {
	r      taskio.Reader
	ch     <-chan int64
	curVal int64
}

func (i *filesHeapItem) next() (ok bool) {
	if i.ch == nil {
		i.ch = i.r.DataCh()
	}

	i.curVal, ok = <-i.ch

	return
}

func NewReader(path string) *Reader {
	return &Reader{
		path: path,
	}
}

func (d *Reader) DataCh() <-chan int64 {
	ch := make(chan int64, taskio.DataChSize)

	go d.gen(ch)

	return ch
}

func (d *Reader) Err() error { return d.lastErr }

func (d *Reader) gen(ch chan<- int64) {
	var fHeap filesHeap

	if err := filepath.WalkDir(d.path, func(path string, e fs.DirEntry, err error) error {
		if e.IsDir() {
			return nil
		}

		f := binaryfile.NewReader(path)
		heapItem := &filesHeapItem{
			r:  f,
			ch: f.DataCh(),
		}
		if !heapItem.next() {
			return f.Err()
		}

		fHeap = append(fHeap, heapItem)
		return nil
	}); err != nil {
		d.lastErr = err
		close(ch)
		return
	}

	heap.Init(&fHeap)

	for {
		if fHeap.Len() == 0 {
			close(ch)
			return
		}

		fItem := heap.Pop(&fHeap).(*filesHeapItem)
		ch <- fItem.curVal

		if fItem.next() {
			heap.Push(&fHeap, fItem)
		} else if err := fItem.r.Err(); err != nil {
			d.lastErr = err
			close(ch)
			return
		}
	}
}
