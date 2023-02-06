package pools

import (
	"bufio"
	"sync"
)

var (
	PoolBufWriters = &sync.Pool{
		New: func() any {
			return bufio.NewWriter(nil)
		},
	}

	PoolBufReaders = &sync.Pool{
		New: func() any {
			return bufio.NewReader(nil)
		},
	}

	PoolBuf4k = &sync.Pool{
		New: func() any {
			return make([]byte, 4096)
		},
	}
)
