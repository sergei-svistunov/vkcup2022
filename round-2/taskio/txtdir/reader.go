package txtdir

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"

	"round-2/taskio"
	"round-2/taskio/textfile"
)

var _ taskio.Reader = &Reader{}

type Reader struct {
	path    string
	workers int
	lastErr error
}

func NewReader(path string, workers int) *Reader {
	return &Reader{
		path:    path,
		workers: workers,
	}
}

func (d *Reader) DataCh() <-chan int64 {
	ch := make(chan int64, taskio.DataChSize)

	go d.gen(ch)

	return ch
}

func (d *Reader) Err() error { return d.lastErr }

func (d *Reader) gen(ch chan<- int64) {
	filesCh := make(chan string)
	defer close(filesCh)

	files, err := getTxtFiles(d.path)
	if err != nil {
		d.lastErr = err
		close(ch)
		return
	}

	activeWorkers := int32(d.workers)
	for i := 0; i < d.workers; i++ {
		go func() {
			for filename := range filesCh {
				ns := textfile.NewReader(filename)
				ns.ReadTo(ch, func() {})
				if err := ns.Err(); err != nil {
					d.lastErr = err
					break
				}
			}

			if atomic.AddInt32(&activeWorkers, -1) == 0 { // Last worker closes numbers channel
				close(ch)
			}
		}()
	}

	for _, f := range files {
		if d.lastErr != nil {
			return
		}

		filesCh <- filepath.Join(d.path, f)
	}
}

func getTxtFiles(path string) ([]string, error) {
	d, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	filesInfos, err := d.Readdir(-1)
	if err != nil {
		return nil, err
	}

	sort.Slice(filesInfos, func(i, j int) bool {
		return filesInfos[j].Size() < filesInfos[i].Size()
	})

	res := make([]string, 0, len(filesInfos))
	l := 0
	for _, fi := range filesInfos {
		if !fi.IsDir() && fi.Name() != "res.txt" && strings.HasSuffix(fi.Name(), ".txt") {
			res = append(res, fi.Name())
			l += len(fi.Name())
		}
	}

	// Clean fileinfos memory
	filesInfos = nil
	runtime.GC()

	return res, nil
}
