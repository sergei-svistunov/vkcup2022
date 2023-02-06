package main

import (
	"container/heap"
	"flag"
	"fmt"
	"log"

	"round-2/taskio/textfile"
)

var (
	flagFilename = flag.String("file", "", "")
)

type heapTester []int64

func (h heapTester) Len() int           { return len(h) }
func (h heapTester) Less(i, j int) bool { return h[i] < h[j] }
func (h heapTester) Swap(i, j int)      { panic("swap") }
func (h *heapTester) Push(x any)        { panic("unimplemented") }
func (h *heapTester) Pop() any          { panic("unimplemented") }

func main() {
	flag.Parse()

	if *flagFilename == "" {
		log.Fatal("Use -file to select file")
	}

	txtR := textfile.NewReader(*flagFilename)
	var h heapTester

	for n := range txtR.DataCh() {
		h = append(h, n)
	}

	defer func() {
		if r := recover(); r != nil {
			if fmt.Sprintf("%v", r) == "swap" {
				log.Fatal("The heap file is disordered")
			}
			log.Fatalf("Panic: %v", r)
		}
	}()

	heap.Init(&h)
}
