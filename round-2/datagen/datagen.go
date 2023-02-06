package datagen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"round-2/taskio/crc"
	"round-2/taskio/rand"
	"round-2/taskio/textfile"
)

type Logger interface {
	Logf(format string, args ...any)
}

func GenerateTestFile(filename string, n int) (int64, error) {
	crcFilter := crc.NewReader(rand.Reader(n))

	err := textfile.NewWriter(filename).WriteData(crcFilter)

	return crcFilter.Sum(), err
}

func GenerateTestDir(path string, filesCount, fileRows int, logger Logger) (int64, error) {
	logger.Logf("Start data generation in %s", path)
	start := time.Now()

	if err := os.MkdirAll(path, 0755); err != nil {
		return 0, fmt.Errorf("cannot create destination dir: %w", err)
	}

	targetCrcSum := int64(0)
	for i := 0; i < filesCount; i++ {
		testFilename := filepath.Join(path, fmt.Sprintf("%d.txt", i))
		crcSum, err := GenerateTestFile(testFilename, fileRows)
		if err != nil {
			return 0, err
		}

		targetCrcSum += crcSum
	}

	logger.Logf("Done: %s", time.Now().Sub(start))
	
	return targetCrcSum, nil
}

func GenerateTmpTestDir(filesCount, fileRows int, logger Logger) (string, int64, error) {
	tmpDir, err := ioutil.TempDir("", "gotest*")
	if err != nil {
		return "", 0, err
	}

	crcSum, err := GenerateTestDir(tmpDir, filesCount, fileRows, logger)

	return tmpDir, crcSum, err
}
