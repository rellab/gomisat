build:
	mkdir -p bin
	go build -o bin/gomisat cmd/main.go

bench:
	mkdir -p pprof
	go test -bench=. -benchmem -benchtime 3s -o pprof/test.bin -cpuprofile pprof/cpu.out
	go tool pprof --svg pprof/test.bin pprof/cpu.out > pprof/test.svg

test:
	cd pkg/gomisat/ && go test -v -cover

clean:
	rm -fR pprof



