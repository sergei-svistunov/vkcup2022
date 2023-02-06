package mypng

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"sync"

	"github.com/klauspost/compress/zlib"
)

var RawDataPool = &sync.Pool{
	New: func() any {
		return &RawData{
			Buffer: bytes.NewBuffer(make([]byte, 0, 8*1024)),
		}
	},
}

var discardBuf [8 * 1024]byte

type RawData struct {
	*bytes.Buffer
	cr, pr [512*3 + 1]byte
	zr     io.ReadCloser
}

func (r *RawData) ColorModel() color.Model { panic("implement me") }
func (r *RawData) Bounds() image.Rectangle { panic("implement me") }
func (r *RawData) At(x, y int) color.Color { panic("implement me") }

func (r *RawData) Reset() {
	r.Buffer.Reset()
	if r.zr != nil {
		r.zr.Close()
		r.zr = nil
	}
}

func (r *RawData) Next() []byte {
	const bytesPerPixel = 3

	if r.zr == nil {
		r.zr, _ = zlib.NewReader(r.Buffer)
	}

	// Read the decompressed bytes.
	_, err := io.ReadFull(r.zr, r.cr[:])
	if err != nil {
		panic(err)
	}

	// Apply the filter.
	cdat := r.cr[1:]
	pdat := r.pr[1:]
	switch r.cr[0] {
	case 0:
		// No-op.
	case 1:
		for i := bytesPerPixel; i < len(cdat); i++ {
			cdat[i] += cdat[i-bytesPerPixel]
		}
	case 2:
		for i, p := range pdat {
			cdat[i] += p
		}
	case 3:
		// The first column has no column to the left of it, so it is a
		// special case. We know that the first column exists because we
		// check above that width != 0, and so len(cdat) != 0.
		for i := 0; i < bytesPerPixel; i++ {
			cdat[i] += pdat[i] / 2
		}
		for i := bytesPerPixel; i < len(cdat); i++ {
			cdat[i] += uint8((int(cdat[i-bytesPerPixel]) + int(pdat[i])) / 2)
		}
	case 4:
		filterPaeth(cdat, pdat, bytesPerPixel)
	default:
		panic(fmt.Errorf("bad filter type %d", r.cr[0]))
	}

	r.pr, r.cr = r.cr, r.pr

	return cdat
}

type pngIHDR struct {
	Width             uint32
	Height            uint32
	BitDepth          uint8
	ColorType         uint8
	CompressionMethod uint8
	FilterMethod      uint8
	InterlaceMethod   uint8
}

func Decode(r io.Reader) (image.Image, error) {
	sr := NewSniffer(r)
	hdr, err := readIHDR(sr)
	if err != nil {
		return nil, err
	}

	if hdr.Width != 512 || hdr.Height != 512 {
		return nil, errors.New("invalid PNG size")
	}

	if hdr.BitDepth != 8 || hdr.ColorType != 2 || hdr.CompressionMethod != 0 || hdr.FilterMethod != 0 || hdr.InterlaceMethod != 0 {
		sr.Restart()
		return png.Decode(sr)
	}

	chunkId, length, err := readChunkHeader(sr)
	if chunkId != "IDAT" {
		return nil, errors.New("invalid PNG")
	}

	rawData := RawDataPool.Get().(*RawData)
	rawData.Reset()
	if _, err := io.CopyN(rawData, sr, int64(length)); err != nil {
		return nil, err
	}

	// Skip all rest data
	for {
		if _, err := sr.Read(discardBuf[:]); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	return rawData, nil
}

func readIHDR(r io.Reader) (pngIHDR, error) {
	var tmp [17]byte

	// Read signature
	_, err := io.ReadFull(r, tmp[:8])
	if err != nil {
		return pngIHDR{}, err
	}

	id, length, err := readChunkHeader(r)
	if id != "IHDR" || length != 13 {
		return pngIHDR{}, errors.New("invalid PNG")
	}

	// Read header data
	if _, err := io.ReadFull(r, tmp[:17]); err != nil {
		return pngIHDR{}, err
	}

	return pngIHDR{
		Width:             binary.BigEndian.Uint32(tmp[0:4]),
		Height:            binary.BigEndian.Uint32(tmp[4:8]),
		BitDepth:          tmp[8],
		ColorType:         tmp[9],
		CompressionMethod: tmp[10],
		FilterMethod:      tmp[11],
		InterlaceMethod:   tmp[12],
	}, nil
}

func readChunkHeader(r io.Reader) (string, uint32, error) {
	var tmp [8]byte
	if _, err := io.ReadFull(r, tmp[:]); err != nil {
		return "", 0, err
	}

	return string(tmp[4:8]), binary.BigEndian.Uint32(tmp[:4]), nil

}

func filterPaeth(cdat, pdat []byte, bytesPerPixel int) {
	var a, b, c, pa, pb, pc int
	for i := 0; i < bytesPerPixel; i++ {
		a, c = 0, 0
		for j := i; j < len(cdat); j += bytesPerPixel {
			b = int(pdat[j])
			pa = b - c
			pb = a - c
			pc = abs(pa + pb)
			pa = abs(pa)
			pb = abs(pb)
			if pa <= pb && pa <= pc {
				// No-op.
			} else if pb <= pc {
				a = b
			} else {
				a = c
			}
			a += int(cdat[j])
			a &= 0xff
			cdat[j] = uint8(a)
			c = b
		}
	}
}

// intSize is either 32 or 64.
const intSize = 32 << (^uint(0) >> 63)

func abs(x int) int {
	// m := -1 if x < 0. m := 0 otherwise.
	m := x >> (intSize - 1)

	// In two's complement representation, the negative number
	// of any number (except the smallest one) can be computed
	// by flipping all the bits and add 1. This is faster than
	// code with a branch.
	// See Hacker's Delight, section 2-4.
	return (x ^ m) - m
}
