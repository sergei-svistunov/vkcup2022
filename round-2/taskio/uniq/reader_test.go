package uniq

import (
	"testing"

	"round-2/taskio/slice"
)

func TestUniqReader(t *testing.T) {
	r := NewReader(slice.Reader{1, 1, 2, 3, 3, 4, 4, 4, 5, 6, 7, 8, 8})

	var got []int64
	for n := range r.DataCh() {
		got = append(got, n)
	}

	expected := []int64{1, 2, 3, 4, 5, 6, 7, 8}

	if len(expected) != len(got) {
		t.Fatalf("Got slice with size %d, but expected with size %d", len(got), len(expected))
	}
	for i := range expected {
		if expected[i] != got[i] {
			t.Fatalf("Element #%d must be %d, but it's %d", i, expected[i], got[i])
		}
	}
}
