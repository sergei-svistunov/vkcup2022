package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"round-3/mypng"
)

const (
	collageWidth = 32
	imageWidth   = 512
	imageHeight  = 512
	rowImgBufSz  = 16
)

var (
	flagProfileCpu    = flag.String("profile-cpu", "", "A file for CPU pprof data")
	flagProfileMem    = flag.String("profile-mem", "", "A file for Memory pprof data")
	flagUseStdDecoder = flag.Bool("std-decoder", false, "Use standard PNG decoder")
)

func main() {
	start := time.Now()
	flag.Parse()

	if len(flag.Args()) != 2 {
		fatalf("usage %s <url> <result.png>", os.Args[0])
	}
	addr := flag.Arg(0)
	resFile := flag.Arg(1)

	fmt.Printf("Start crawling from %s and storing into %s\n", addr, resFile)

	if *flagProfileCpu != "" {
		f, err := os.Create(*flagProfileCpu)
		if err != nil {
			fatalf("could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fatalf("could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	tasksCh := make(chan ImgTask, 1024)
	c := NewCrawler(*flagUseStdDecoder)
	go func() {
		//for i := 0; i < 10; i++ {
		if err := c.Walk(addr, tasksCh); err != nil {
			fatalf("%v", err)
		}
		//}
		close(tasksCh)
	}()

	workers := runtime.NumCPU()

	f, err := os.Create(resFile)
	if err != nil {
		fatalf("%v", err)
	}
	defer f.Close()

	palette := make(color.Palette, 255)
	for i := range palette {
		palette[i] = color.NRGBA{0, 0, uint8(i), 255}
	}

	enc := mypng.NewEncoder(f)
	if err := enc.WriteIHDR(imageWidth*collageWidth, imageHeight); err != nil {
		fatalf("%v", err)
	}
	if err := enc.WritePLTE(palette); err != nil {
		fatalf("%v", err)
	}

	rowImgBufCh := make(chan *image.Paletted, rowImgBufSz)
	for i := 0; i < rowImgBufSz; i++ {
		rowImgBufCh <- image.NewPaletted(image.Rect(0, 0, imageWidth*collageWidth, imageHeight), nil)
	}
	rowImgCh := make(chan *image.Paletted, rowImgBufSz)

	type chunk struct {
		rowImg        *image.Paletted
		leftRowImages int32
	}
	chunksMap := map[uint16]*chunk{}
	chunksMtx := sync.RWMutex{}
	line := uint16(0)

	activeWorkers := int32(workers)
	for i := 0; i < workers; i++ {
		go func() {
			for task := range tasksCh {
				img, err := c.GetImage(task.Location)
				if err != nil {
					fatalf("%v", err)
				}

				chunkId := uint16(task.Id / collageWidth)

				chunksMtx.Lock()
				curChunk := chunksMap[(chunkId)]
				if curChunk == nil {
					curChunk = &chunk{
						rowImg:        <-rowImgBufCh,
						leftRowImages: int32(collageWidth),
					}
					chunksMap[chunkId] = curChunk
				}
				chunksMtx.Unlock()

				copyImage(img, curChunk.rowImg, (task.Id%collageWidth)*imageWidth)

				// Row is complete
				if atomic.AddInt32(&curChunk.leftRowImages, -1) == 0 {
					if chunkId != line {
						continue
					}

					// Store chunk
					rowImgCh <- curChunk.rowImg
					line++

					// Check if next chunks have finished earlier
					for {
						chunksMtx.RLock()
						c := chunksMap[line]
						chunksMtx.RUnlock()

						if c != nil && c.leftRowImages == 0 {
							rowImgCh <- c.rowImg
							line++
						} else {
							break
						}
					}
				}
			}

			// All workers are done
			if atomic.AddInt32(&activeWorkers, -1) == 0 {
				if lastChunk := chunksMap[line]; lastChunk != nil && lastChunk.leftRowImages != collageWidth {
					// Clean unused space
					for y := 0; y < imageHeight; y++ {
						dstPixOffset := lastChunk.rowImg.PixOffset(int((collageWidth-lastChunk.leftRowImages)*imageWidth), y)
						for x := 0; x < int(lastChunk.leftRowImages)*imageWidth; x++ {
							lastChunk.rowImg.Pix[dstPixOffset+x] = 0
						}
					}

					// Store partial chunk
					rowImgCh <- lastChunk.rowImg
					line++
				}

				close(rowImgCh)
			}
		}()
	}

	for rowImg := range rowImgCh {
		if err := enc.WriteIDAT(rowImg); err != nil {
			fatalf("%v", err)
		}
		rowImgBufCh <- rowImg
	}

	if err := enc.FinishIDAT(); err != nil {
		fatalf("%v", err)
	}

	if err := enc.WriteIEND(); err != nil {
		fatalf("%v", err)
	}

	// Fix size
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		fatalf("%v", err)
	}
	if err := enc.WriteIHDR(imageWidth*collageWidth, uint32(imageHeight*line)); err != nil {
		fatalf("%v", err)
	}

	fmt.Printf("Processed %d files in %s\n", int(line-1)*collageWidth+collageWidth-int(chunksMap[line-1].leftRowImages), time.Now().Sub(start))

	if *flagProfileMem != "" {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		fmt.Printf("Memory allocated: %d Mb\n", memStats.HeapAlloc/1024/1024)
		fmt.Printf("Memory in use: %d Mb\n", memStats.HeapInuse/1024/1024)

		f, err := os.Create(*flagProfileMem)
		if err != nil {
			fatalf("could not create memory profile: %v", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fatalf("could not write memory profile: %v", err)
		}
	}
}

func copyImage(src image.Image, dst *image.Paletted, offsetX int) {
	switch src := src.(type) {
	case *mypng.RawData:
		dstAddr := uintptr(unsafe.Pointer(&dst.Pix[0]))
		for y := 0; y < imageHeight; y++ {
			srcPix := src.Next()
			srcAddr := uintptr(unsafe.Pointer(&srcPix[0]))
			dstPixOffset := dst.PixOffset(offsetX, y)
			for x := 0; x < imageWidth; x++ {
				//dst.Pix[dstPixOffset+x] = src.Pix[srcPixOffset+x*4+1]
				dstPtr := (*uint8)((unsafe.Pointer)(dstAddr + uintptr(dstPixOffset+x)))
				srcPtr := (*uint8)((unsafe.Pointer)(srcAddr + uintptr(x*3+2)))
				*dstPtr = *srcPtr
			}
		}
		mypng.RawDataPool.Put(src)

	case *image.RGBA:
		dstAddr := uintptr(unsafe.Pointer(&dst.Pix[0]))
		srcAddr := uintptr(unsafe.Pointer(&src.Pix[0]))
		for y := 0; y < imageHeight; y++ {
			srcPixOffset := src.PixOffset(0, y)
			dstPixOffset := dst.PixOffset(offsetX, y)
			for x := 0; x < imageWidth; x++ {
				//dst.Pix[dstPixOffset+x] = src.Pix[srcPixOffset+x*4+1]
				dstPtr := (*uint8)((unsafe.Pointer)(dstAddr + uintptr(dstPixOffset+x)))
				srcPtr := (*uint8)((unsafe.Pointer)(srcAddr + uintptr(srcPixOffset+x*4+2)))
				*dstPtr = *srcPtr
			}
		}

	default:
		for y := 0; y < imageHeight; y++ {
			dstPixOffset := dst.PixOffset(offsetX, y)
			for x := 0; x < imageWidth; x++ {
				_, _, b, _ := src.At(x, y).RGBA()
				dst.Pix[dstPixOffset+x] = uint8(b)
			}
		}
	}
}

func fatalf(format string, v ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", fmt.Sprintf(format, v...))
	os.Exit(255)
}
