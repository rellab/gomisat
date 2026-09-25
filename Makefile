BIN       := bin
CORPUS    := testdata/satlib
BASELINE  := bench/baseline.csv

.PHONY: all build test test-full check fmt vet bench bench-suite baseline clean

all: build

build:
	mkdir -p $(BIN)
	go build -o $(BIN)/gomisat ./cmd
	go build -o $(BIN)/gomibench ./cmd/gomibench

# Fast regression run: a sample of every SATLIB family plus the brute-force and
# reuse checks.
test:
	go test ./...

# The same suite over all 2186 instances of the corpus.
test-full:
	GOMISAT_FULL=1 go test ./... -timeout 30m

check: vet test

vet:
	go vet ./...

fmt:
	gofmt -l -w .

# Micro benchmarks with a CPU profile.
bench:
	mkdir -p pprof
	go test -bench=. -benchmem -benchtime 3s -o pprof/test.bin -cpuprofile pprof/cpu.out
	go tool pprof --svg pprof/test.bin pprof/cpu.out > pprof/test.svg

# Sweep the corpus and compare against the recorded baseline. Answers that
# changed make this fail; timings are reported as a median ratio.
bench-suite: build
	$(BIN)/gomibench -check -timeout 60 -baseline $(BASELINE) -o /dev/null $(CORPUS)

# Record a new baseline. Commit the result together with the change that
# justifies it.
baseline: build
	$(BIN)/gomibench -check -timeout 60 -o $(BASELINE) $(CORPUS)

clean:
	rm -fR pprof $(BIN)
