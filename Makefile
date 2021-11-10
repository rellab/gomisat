build:
	mkdir -p bin
	go build -o bin/gomisat cmd/main.go

check:
	mkdir -p bin
	go build -gcflags "-m -m" -o bin/gomisat cmd/main.go

bench:
	cd test &&\
	mkdir -p pprof &&\
	go test -bench=Bench -benchmem -benchtime 3s -o pprof/test.bin -cpuprofile pprof/cpu.out &&\
	go tool pprof --svg pprof/test.bin pprof/cpu.out > pprof/test.svg

test: test_gomisat test_sat test_dimacs

test_gomisat:
	cd test && go test -v -cover

test_sat:
	cd pkg/sat && go test -v -cover

test_dimacs:
	cd pkg/dimacs && go test -v -cover

clean:
	rm -fR pprof



