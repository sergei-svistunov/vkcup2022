package mypng_test

import (
	"image"
	"image/png"
	"io"
	"os"
	"testing"

	"round-3/mypng"
)

var (
	img   image.Image
	chunk []byte
)

func BenchmarkStdDecoder(b *testing.B) {
	f, err := os.Open("test.png")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	for i := 0; i < b.N; i++ {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			b.Fatal()
		}

		img, err = png.Decode(f)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMyDecoder(b *testing.B) {
	f, err := os.Open("test.png")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	for i := 0; i < b.N; i++ {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			b.Fatal()
		}

		img, err = mypng.Decode(f)
		if err != nil {
			b.Fatal(err)
		}
		rawImg := img.(*mypng.RawData)
		for i := 0; i < 512; i++ {
			chunk = rawImg.Next()
		}
	}
}
