package heapfile

import (
	"container/heap"
	"io/ioutil"
	"math"
	"os"
	"testing"
	"time"

	"round-2/taskio/binaryfile"
	"round-2/taskio/crc"
	"round-2/taskio/rand"
)

const (
	//NumbersCount = 7_000_000
	//MemLimit     = 10 * 1024 * 1024
	NumbersCount = 100_000
	MemLimit     = 131072
)

func TestHeapFile(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Cannot create tmp file: %v", err)
	}
	filename := f.Name()
	_ = f.Close()
	defer os.Remove(filename)

	t.Logf("Writing chunks")
	start := time.Now()

	crcR := crc.NewReader(rand.Reader(NumbersCount))
	if err := NewWriter(filename, MemLimit).WriteData(crcR); err != nil {
		t.Fatal(err)
	}

	t.Logf("Done: %s", time.Now().Sub(start))

	t.Logf("Merging chunks")
	start = time.Now()

	fixer := NewFixer(filename, MemLimit)
	if err := fixer.WriteData(nil); err != nil {
		t.Fatal(err)
	}
	t.Logf("Stat: %#v", fixer.stat)

	t.Logf("Done: %s", time.Now().Sub(start))

	t.Logf("Checking result")
	bf := binaryfile.NewReader(filename)

	crcSum := int64(0)
	count := 0
	intHeap := make(Int64Heap, 0, NumbersCount)
	for n := range bf.DataCh() {
		intHeap = append(intHeap, n)
		crcSum += n
		count++
	}
	if err := bf.Err(); err != nil {
		t.Fatalf("Cannot read numbers: %v", err)
	}
	if count != NumbersCount {
		t.Fatalf("Read invalid quantity of numbers: %d instead of %d", count, NumbersCount)
	}
	if crcSum != crcR.Sum() {
		t.Fatal("CRC is not valid")
	}

	prevValue := int64(math.MinInt64)
	for intHeap.Len() > 0 {
		n := heap.Pop(&intHeap).(int64)
		if n < prevValue {
			t.Fatalf("Disorder: %d > %d", prevValue, n)
		}
		prevValue = n
	}
	t.Logf("Done: %s", time.Now().Sub(start))
}
