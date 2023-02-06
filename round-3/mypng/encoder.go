package mypng

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"io"
	"strconv"

	"github.com/klauspost/compress/zlib"
)

const pngHeader = "\x89PNG\r\n\x1a\n"

type Encoder struct {
	w      io.Writer
	header [8]byte
	footer [4]byte
	tmp    [4 * 256]byte

	buf *bytes.Buffer
	zw  *zlib.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	buf := bytes.NewBuffer(make([]byte, 0, 4*1024*1024))

	zw, err := zlib.NewWriterLevel(buf, zlib.BestSpeed)
	if err != nil {
		panic(err)
	}

	return &Encoder{
		w:   w,
		buf: buf,
		zw:  zw,
	}
}

func (e *Encoder) WriteIHDR(w, h uint32) error {
	if _, err := io.WriteString(e.w, pngHeader); err != nil {
		return err
	}

	binary.BigEndian.PutUint32(e.tmp[0:4], w)
	binary.BigEndian.PutUint32(e.tmp[4:8], h)
	// Set bit depth and color type.
	e.tmp[8] = 8
	e.tmp[9] = 3
	e.tmp[10] = 0 // default compression method
	e.tmp[11] = 0 // default filter method
	e.tmp[12] = 0 // non-interlaced

	return e.writeChunk(e.tmp[:13], "IHDR")
}

func (e *Encoder) WritePLTE(p color.Palette) error {
	if len(p) < 1 || len(p) > 256 {
		return errors.New("bad palette length: " + strconv.Itoa(len(p)))
	}
	for i, c := range p {
		c1 := color.NRGBAModel.Convert(c).(color.NRGBA)
		e.tmp[3*i+0] = c1.R
		e.tmp[3*i+1] = c1.G
		e.tmp[3*i+2] = c1.B
	}

	return e.writeChunk(e.tmp[:3*len(p)], "PLTE")
}

func (e *Encoder) WriteIDAT(i *image.Paletted) error {
	b := i.Bounds()

	for y := 0; y < b.Max.Y; y++ {
		if _, err := e.zw.Write([]byte{0}); err != nil {
			return err
		}

		offset := y * i.Stride

		if _, err := e.zw.Write(i.Pix[offset : offset+b.Max.X]); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) FinishIDAT() error {
	if err := e.zw.Close(); err != nil {
		return err
	}

	return e.writeChunk(e.buf.Bytes(), "IDAT")
}

func (e *Encoder) WriteIEND() error { return e.writeChunk(nil, "IEND") }

func (e *Encoder) writeChunk(b []byte, name string) error {
	n := uint32(len(b))
	if int(n) != len(b) {
		return errors.New(name + " chunk is too large: " + strconv.Itoa(len(b)))

	}
	binary.BigEndian.PutUint32(e.header[:4], n)
	e.header[4] = name[0]
	e.header[5] = name[1]
	e.header[6] = name[2]
	e.header[7] = name[3]
	crc := crc32.NewIEEE()
	_, _ = crc.Write(e.header[4:8])
	_, _ = crc.Write(b)
	binary.BigEndian.PutUint32(e.footer[:4], crc.Sum32())

	if _, err := e.w.Write(e.header[:8]); err != nil {
		return err
	}
	if _, err := e.w.Write(b); err != nil {
		return err
	}
	if _, err := e.w.Write(e.footer[:4]); err != nil {
		return err
	}

	return nil
}
