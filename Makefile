bench:
	mkdir -p pprof
	go test -bench=. -benchmem -benchtime 10s -o pprof/test.bin -cpuprofile pprof/cpu.out
	go tool pprof --svg pprof/test.bin pprof/cpu.out > pprof/test.svg

test:
	cd pkg/gomisat/ && go test -v -cover

clean:
	rm -fR pprof



