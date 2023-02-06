GO_FILES = $(shell find ./ -type f -name '*.go' -not -path ".var")

.PHONY: build test e2etest e2etest_sort e2etest_uniq e2etest_heap clean

build: bin/solution bin/datagen bin/heapcheck

test:
	go test ./...

e2etest: e2etest_sort e2etest_uniq e2etest_heap

clean:
	@rm -rf .var bin

bin/%: ${GO_FILES}
	@mkdir -p bin
	@go build -o $@ ./cmd/$(notdir $@)/main.go

e2etest_sort: bin/solution .var/testdata
	@echo "Generating res.txt with the -sort key"
	@systemd-run --scope --user -q -p MemoryMax=500M bin/solution -sort -dir .var/testdata -no-cleanup
	@echo "	Checking if res.txt is sorted"
	@sort --check -n .var/testdata/res.txt

e2etest_uniq: bin/solution .var/testdata
	@echo "Generating res.txt with the -uniq key"
	@systemd-run --scope --user -q -p MemoryMax=500M bin/solution -uniq -dir .var/testdata
	@echo "	Checking if res.txt is sorted and uniq"
	@sort --check -u -n .var/testdata/res.txt

e2etest_heap: bin/solution .var/testdata
	@echo "Generating res.txt with -heap key"
	@systemd-run --scope --user -q -p MemoryMax=500M bin/solution -heap -dir .var/testdata
	@echo "	Checking if res.txt is a valid heap"
	@go run ./cmd/heapcheck/main.go -file .var/testdata/res.txt

.var/testdata:
	@echo "Preparing test data"
	@mkdir -p .var/testdata
	@echo "	Generating 499'999 small files..."
	@go run ./cmd/datagen/main.go -dir $@ -files 499_999 -rows 400
	@echo "	Generating a huge file..."
	@go run ./cmd/datagen/main.go -dir $@ -files 1 -rows 66_000_000 -prefix huge
