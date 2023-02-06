package main

import (
	"flag"
	"log"

	"round-2/datagen"
)

var (
	flagDir    = flag.String("dir", "", "The directory where files will be stores")
	flagFiles  = flag.Int("files", 1000, "The number of files")
	flagRows   = flag.Int("rows", 1000, "The number of rows in each file")
	flagPrefix = flag.String("prefix", "", "The prefix in filename <prefix><num>.txt")
)

type Logger struct{}

func (l Logger) Logf(format string, args ...any) { log.Printf(format, args...) }

func main() {
	flag.Parse()

	if *flagDir == "" {
		log.Fatal("Missed required parameter -dir")
	}

	if _, err := datagen.GenerateTestDir(*flagDir, *flagFiles, *flagRows, Logger{}); err != nil {
		log.Fatalf("Cannot generate data: %v", err)
	}
}
