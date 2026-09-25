BIN       := bin
CORPUS    := testdata/satlib
BASELINE  := bench/baseline.csv
# The large corpus lives outside the repository; see scripts/fetch-corpus.sh.
EXTCORPUS := $(if $(GOMISAT_CORPUS),$(GOMISAT_CORPUS),$(HOME)/.cache/gomisat/corpus)
TIMEOUT   := 10
EXTBASE   := bench/corpus-t$(TIMEOUT)-repaired-lbd.csv
EXPECTED  := corpus/expected.tsv

.PHONY: all build test test-full check fmt vet bench bench-suite baseline \
	corpus corpus-expected corpus-sweep corpus-compare clean

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

# Download the benchmark corpus described by corpus/manifest.tsv. It is the
# corpus the committed testdata cannot replace: harder instances, several
# families, and a difficulty gradient.
corpus:
	scripts/fetch-corpus.sh

# Regenerate the table of known answers with an independent solver. Needs
# python-sat; the result is committed, so this is rarely necessary.
corpus-expected:
	scripts/expected-status.py > $(EXPECTED)

# Record the corpus results, verifying every decided answer on the way.
corpus-sweep: build
	$(BIN)/gomibench -check -expected $(EXPECTED) -timeout $(TIMEOUT) -o $(EXTBASE) $(EXTCORPUS)

# Compare against the recorded corpus results. This is the measurement that
# decides whether a change to the search is an improvement.
corpus-compare: build
	$(BIN)/gomibench -check -expected $(EXPECTED) -timeout $(TIMEOUT) -baseline $(EXTBASE) -o /dev/null $(EXTCORPUS)

clean:
	rm -fR pprof $(BIN)
