package test

import (
	"com.github/rellab/gomisat/pkg/dimacs"
	"com.github/rellab/gomisat/pkg/sat"
	_"fmt"
	"io"
	"log"
	"os"
	"testing"
)

func BenchmarkDimacs01(b *testing.B) {
	log.SetOutput(io.Discard)
	file, _ := os.Open("../testdata/aim-100-1_6-no-1.cnf")
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(buf)
	for i := 0; i < b.N; i++ {
		s := sat.NewSolver()
		options := sat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs02(b *testing.B) {
	log.SetOutput(io.Discard)
	file, _ := os.Open("../testdata/aim-50-1_6-yes1-4.cnf")
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(buf)
	for i := 0; i < b.N; i++ {
		s := sat.NewSolver()
		options := sat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs03(b *testing.B) {
	log.SetOutput(io.Discard)
	file, _ := os.Open("../testdata/bf0432-007.cnf")
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(buf)
	for i := 0; i < b.N; i++ {
		s := sat.NewSolver()
		options := sat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}
