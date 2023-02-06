package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"round-2/processor"
	"round-2/taskio"
	"round-2/taskio/binaryfile"
	"round-2/taskio/heapfile"
	"round-2/taskio/sorteddir"
	"round-2/taskio/textfile"
	"round-2/taskio/txtdir"
	"round-2/taskio/uniq"
)

var (
	flagSort       = flag.Bool("sort", false, "Write sorted numbers in res.txt")
	flagUniq       = flag.Bool("uniq", false, "Write unique numbers in res.txt")
	flagHeap       = flag.Bool("heap", false, "Write \"heapyfied\" numbers in res.txt")
	flagDir        = flag.String("dir", ".", "The directory with *.txt files")
	flagNoCleanup  = flag.Bool("no-cleanup", false, "Do not remove temporary files")
	flagMemLimit   = flag.Int("mem-limit", 380, "Memory limit in Mb for sorting data, not whole application")
	flagProfileCpu = flag.String("profile-cpu", "", "A file for CPU pprof data")
	flagProfileMem = flag.String("profile-mem", "", "A file for Memory pprof data")
)

func main() {
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

	flag.Parse()

	*flagMemLimit *= 1024 * 1024

	if !(*flagSort || *flagUniq || *flagHeap) {
		fatalf("no mod has been selected. Use -sort, -uniq or -heap")
	}

	if *flagHeap && (*flagSort || *flagUniq) {
		fatalf("cannot use -heap with -sort or -uniq")
	}

	taskProcessor := processor.New()

	srcTxtDir := txtdir.NewReader(*flagDir, numCPU(1))
	var resReader taskio.Reader

	// -sort / -uniq
	if *flagSort || *flagUniq {
		srtDirPath := filepath.Join(*flagDir, ".srt")
		if !*flagNoCleanup {
			defer os.RemoveAll(srtDirPath)
		}

		taskProcessor.AddStage(
			srcTxtDir,
			sorteddir.NewWriter(srtDirPath, numCPU(2), *flagMemLimit),
			"Preparing sorted chunks",
			processor.WithDoneFile(filepath.Join(srtDirPath, ".done")),
		)

		resReader = sorteddir.NewReader(srtDirPath)

		if *flagUniq {
			resReader = uniq.NewReader(resReader)
		}
	}

	// -heap
	if *flagHeap {
		heapDir := filepath.Join(*flagDir, ".heap")
		if err := os.MkdirAll(heapDir, 0700); err != nil {
			fatalf("%v", err)
		}
		if !*flagNoCleanup {
			defer os.RemoveAll(heapDir)
		}

		heapFilename := filepath.Join(heapDir, "bin")

		taskProcessor.AddStage(
			srcTxtDir,
			heapfile.NewWriter(heapFilename, *flagMemLimit),
			"Preparing heap chunks",
			processor.WithDoneFile(filepath.Join(heapDir, ".done")),
		)

		taskProcessor.AddStage(
			nil,
			heapfile.NewFixer(heapFilename, *flagMemLimit),
			"Merging heap chunks",
			processor.WithDoneFile(filepath.Join(heapDir, ".fixed")),
		)

		resReader = binaryfile.NewReader(heapFilename)
	}

	// Final stage
	taskProcessor.AddStage(
		resReader,
		textfile.NewWriter(filepath.Join(*flagDir, "res.txt")),
		"Writing res.txt",
	)

	if err := taskProcessor.Run(); err != nil {
		fatalf("%v", err)
	}

	if *flagProfileMem != "" {
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

func fatalf(format string, args ...any) {
	log.Printf("Error: %s", fmt.Sprintf(format, args...))
	os.Exit(1)
}

func numCPU(divider int) int {
	n := runtime.NumCPU() / divider
	if n == 0 {
		return 1
	}
	return n
}
