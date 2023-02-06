package main

import (
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"strconv"
	"sync/atomic"
	"time"
)

type job func(in, out chan interface{})

const (
	MaxInputDataLen = 64
)

var (
	dataOverheat uint32 = 0
	Salt                = ""
)

var OverLock = func() {
	for {
		if swapped := atomic.CompareAndSwapUint32(&dataOverheat, 0, 1); !swapped {
			fmt.Println("OverheatLock happened")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

var OverUnlock = func() {
	for {
		if swapped := atomic.CompareAndSwapUint32(&dataOverheat, 1, 0); !swapped {
			fmt.Println("OverheatUnlock happened")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

var FastPredict = func(data string) string {
	OverLock()
	defer OverUnlock()
	data += Salt
	dataHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
	time.Sleep(10 * time.Millisecond)
	return dataHash
}

var SlowPredict = func(data string) string {
	data += Salt
	crcH := crc32.ChecksumIEEE([]byte(data))
	dataHash := strconv.FormatUint(uint64(crcH), 10)
	time.Sleep(time.Second)
	return dataHash
}
