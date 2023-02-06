package mypng

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestMyPngPaletted(t *testing.T) {
	imgF, err := os.Open("test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer imgF.Close()

	img, err := png.Decode(imgF)
	if err != nil {
		t.Fatal(err)
	}

	palette := make(color.Palette, 255)
	for i := range palette {
		palette[i] = color.NRGBA{0, uint8(i), 0, 255}
	}
	palettedImg := image.NewPaletted(img.Bounds(), palette)
	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			_, g, _, _ := img.At(x, y).RGBA()
			palettedImg.Set(x, y, color.NRGBA{0, uint8(g), 0, 255})
		}
	}

	f, err := os.Create("/tmp/test.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	enc := NewEncoder(f)
	if err := enc.WriteIHDR(uint32(img.Bounds().Max.X), uint32(img.Bounds().Max.Y)); err != nil {
		t.Fatal(err)
	}

	if err := enc.WritePLTE(palettedImg.Palette); err != nil {
		t.Fatal(err)
	}

	if err := enc.WriteIDAT(palettedImg); err != nil {
		t.Fatal(err)
	}

	if err := enc.FinishIDAT(); err != nil {
		t.Fatal(err)
	}

	if err := enc.WriteIEND(); err != nil {
		t.Fatal(err)
	}

	f2, err := os.Create("/tmp/test2.png")
	if err != nil {
		panic(err)
	}
	defer f2.Close()
}
