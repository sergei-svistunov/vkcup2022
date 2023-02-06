package sorteddir

import (
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"round-2/datagen"
	"round-2/taskio/txtdir"
)

const (
	FilesCount     = 100 //_000
	ElementsInFile = 1000
	MemLimit       = 400 * 1024 * 1024
)

func TestSortedDir(t *testing.T) {
	tmpDir, targetCrc, err := datagen.GenerateTmpTestDir(FilesCount, ElementsInFile, t)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srtDir := filepath.Join(tmpDir, ".srt")

	t.Logf("Start sorting")
	start := time.Now()

	if err := NewWriter(srtDir, runtime.NumCPU(), MemLimit).WriteData(txtdir.NewReader(tmpDir, runtime.NumCPU()/2)); err != nil {
		t.Fatal(err)
	}

	t.Logf("Done: %s", time.Now().Sub(start))

	t.Logf("Start reading")
	start = time.Now()

	sortedDirR := NewReader(srtDir)

	crc := int64(0)
	count := 0
	prevValue := int64(math.MinInt64)
	for n := range sortedDirR.DataCh() {
		if prevValue > n {
			t.Fatalf("Disorder: %d > %d", prevValue, n)
		}
		prevValue = n
		crc += n
		count++
	}
	if err := sortedDirR.Err(); err != nil {
		t.Fatalf("Cannot read sorted numbers: %v", err)
	}
	if count != FilesCount*ElementsInFile {
		t.Fatalf("Read invalid quantity of numbers: %d instead of %d", count, FilesCount*ElementsInFile)
	}
	if targetCrc != crc {
		t.Fatal("CRC is not valid")
	}
	t.Logf("Done: %s", time.Now().Sub(start))
}
