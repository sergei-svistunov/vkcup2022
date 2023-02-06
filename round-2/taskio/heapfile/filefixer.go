package heapfile

import (
	"container/heap"
	"os"

	"round-2/taskio"
)

var _ taskio.Writer = &Fixer{}

type Fixer struct {
	filename string
	file     *os.File
	size     int
	numBuf   [8]byte
	stat     map[string]int
	memLimit int
	ioCache  *ioCache
}

func NewFixer(filename string, memLimit int) *Fixer {
	return &Fixer{
		filename: filename,
		stat:     map[string]int{},
		memLimit: memLimit,
	}
}

func (f *Fixer) WriteData(taskio.Reader) error {
	ff, err := os.OpenFile(f.filename, os.O_RDWR, 0)
	if err != nil {
		return nil
	}
	defer ff.Close()

	stat, err := ff.Stat()
	if err != nil {
		return err
	}

	f.size = int(stat.Size() / 8)
	f.file = ff

	ioCacheSz := f.memLimit / pageSize
	if ioCacheSz == 0 {
		ioCacheSz = 1
	}

	f.ioCache = &ioCache{
		f:     ff,
		size:  ioCacheSz,
		list:  &list{},
		stat:  f.stat,
		cache: make(map[int]*ioCachePage, ioCacheSz),
	}

	heap.Init(f)

	return f.ioCache.flush()

	return nil
}

func (f *Fixer) Len() int           { return f.size }
func (f *Fixer) Less(i, j int) bool { f.stat["less"]++; return f.read(i) < f.read(j) }

func (f *Fixer) Swap(i, j int) {
	f.stat["swap"]++
	nI := f.read(i)
	nJ := f.read(j)
	f.write(i, nJ)
	f.write(j, nI)
}

func (f *Fixer) Push(x any) {
	f.stat["push"]++
	f.write(f.size, x.(int64))
	f.size++
}

func (f *Fixer) Pop() any { panic("implement me") }

func (f *Fixer) read(pos int) int64 {
	f.stat["read"]++

	n, err := f.ioCache.read(pos)
	if err != nil {
		panic(err)
	}

	return n
}

func (f *Fixer) write(pos int, n int64) {
	f.stat["write"]++

	if err := f.ioCache.write(pos, n); err != nil {
		panic(err)
	}
}
