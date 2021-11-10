package main

import (
	"com.github/rellab/gomisat/pkg/dimacs"
	"com.github/rellab/gomisat/pkg/sat"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	fname := os.Args[1]
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(buf)
	s := sat.NewSolver()
	options := sat.DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	start := time.Now()
	s.Simplify()
	fmt.Println(s.Solve(options))
	end := time.Now()
	fmt.Println(s.Conflicts)
	fmt.Println(s.Propagations)
	fmt.Printf("computation time : %.8f (sec)\n", (end.Sub(start)).Seconds())
}
