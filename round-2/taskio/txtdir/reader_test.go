package txtdir

import (
	"os"
	"runtime"
	"testing"

	"round-2/datagen"
)

const (
	FilesCount     = 64
	ElementsInFile = 1000
)

func TestTxtDir(t *testing.T) {
	tmpDir, targetCrc, err := datagen.GenerateTmpTestDir(FilesCount, ElementsInFile, t)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	txtDir := NewReader(tmpDir, runtime.NumCPU())

	count := 0
	crc := int64(0)
	for n := range txtDir.DataCh() {
		count++
		crc += n
	}
	if err := txtDir.Err(); err != nil {
		t.Fatalf("Cannot read numbers: %v", err)
	}
	if count != FilesCount*ElementsInFile {
		t.Fatalf("Read invalid quantity of numbers: %d instead of %d", count, FilesCount*ElementsInFile)
	}
	if targetCrc != crc {
		t.Fatal("CRC is not valid")
	}
}
