package processor

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"round-2/taskio"
)

type Processor struct {
	stages []stage
}

type stage struct {
	src      taskio.Reader
	dst      taskio.Writer
	caption  string
	doneFile string
}

type stageOption func(*stage)

func WithDoneFile(filename string) stageOption {
	return func(s *stage) {
		s.doneFile = filename
	}
}

func New() *Processor {
	return &Processor{}
}

func (p *Processor) AddStage(src taskio.Reader, dst taskio.Writer, caption string, opts ...stageOption) {
	s := stage{
		src:     src,
		dst:     dst,
		caption: caption,
	}

	for _, opt := range opts {
		opt(&s)
	}

	p.stages = append(p.stages, s)
}

func (p *Processor) Run() error {
	for i, s := range p.stages {
		log.Printf("Stage %d: %s", i+1, s.caption)

		if s.doneFile != "" && isFileExists(s.doneFile) {
			log.Printf("\tAlready done, skipped")
			continue
		}

		start := time.Now()

		if err := s.dst.WriteData(s.src); err != nil {
			return err
		}

		//ms := runtime.MemStats{}
		//runtime.ReadMemStats(&ms)
		//log.Printf("\tMem used: %d Mb", ms.HeapInuse/(1024*1024))

		runtime.GC()

		if s.doneFile != "" {
			if err := createFile(s.doneFile); err != nil {
				log.Printf("\tWarning: cannot create done file: %v", err)
			}
		}

		log.Printf("\tDone in %s", time.Now().Sub(start))
	}

	return nil
}

func isFileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}

func createFile(filename string) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	return f.Close()
}
