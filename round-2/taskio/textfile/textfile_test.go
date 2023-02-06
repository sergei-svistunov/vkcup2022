package textfile

import (
	"io/ioutil"
	"os"
	"testing"

	"round-2/taskio/crc"
	"round-2/taskio/rand"
)

const NumbersCount = 10_000

func TestTextFile(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Cannot create tmp file: %v", err)
	}
	filename := f.Name()
	_ = f.Close()
	defer os.Remove(filename)

	crcR := crc.NewReader(rand.Reader(NumbersCount))
	if err := NewWriter(filename).WriteData(crcR); err != nil {
		t.Fatal(err)
	}

	ns := NewReader(filename)

	crcSum := int64(0)
	count := 0
	for n := range ns.DataCh() {
		crcSum += n
		count++
	}
	if err := ns.Err(); err != nil {
		t.Fatalf("Cannot read numbers: %v", err)
	}
	if count != NumbersCount {
		t.Fatalf("Read invalid quantity of numbers: %d instead of %d", count, NumbersCount)
	}
	if crcSum != crcR.Sum() {
		t.Fatal("CRC is not valid")
	}
}
