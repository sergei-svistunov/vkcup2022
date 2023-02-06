package mypng

import (
	"io"
)

const bufSize = 33

type Sniffer struct {
	r    io.Reader
	buf  [bufSize]byte
	left int
}

func NewSniffer(r io.Reader) *Sniffer {
	return &Sniffer{
		r:    r,
		left: -1,
	}
}

func (s *Sniffer) Read(p []byte) (n int, err error) {
	if s.left == -1 {
		n, err := io.ReadFull(s.r, s.buf[:])
		if err != nil {
			return 0, err
		}
		s.left = n
	}

	if s.left == 0 {
		return s.r.Read(p)
	}

	if len(p) <= s.left {
		copy(p, s.buf[bufSize-s.left:])
		s.left -= len(p)
		return len(p), err
	}

	copy(p, s.buf[bufSize-s.left:])
	n, err = s.r.Read(p[s.left:])
	if err != nil {
		return n, err
	}
	n += s.left
	s.left = 0

	return n, nil
}

func (s *Sniffer) Restart() {
	s.left = bufSize
}
