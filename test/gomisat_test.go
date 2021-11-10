package test

import (
	"com.github/rellab/gomisat/pkg/dimacs"
	"com.github/rellab/gomisat/pkg/sat"
	"fmt"
	"io"
	_"log"
	"os"
	"testing"
)

func TestSolver01(t *testing.T) {
	cs, _ := dimacs.ParseDimacs([]byte(`
	p cnf 5 6
	4 5 6 5 0
	-1 2 1 0
	`))
	fmt.Println(cs)
	s := sat.NewSolver()
	options := sat.DefaultSolverOptions()
	fmt.Println(s)
	s.AddClauseFromCode(cs[0], options)
	fmt.Println(s)
}

func TestSolver02(t *testing.T) {
	cs, _ := dimacs.ParseDimacs([]byte(`
	p cnf 1 1
	1 2 0
	-1 0
	`))
	fmt.Println(cs)
	s := sat.NewSolver()
	options := sat.DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	fmt.Println(s)
}

func TestDimacs02(t *testing.T) {
	file, _ := os.Open("testdata/aim-100-1_6-no-1.cnf")
	defer file.Close()
	b, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(b)
	s := sat.NewSolver()
	options := sat.DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	result, _ := s.Solve(options)
	fmt.Println("Result", result)
}

func TestDimacs03(t *testing.T) {
	file, _ := os.Open("testdata/aim-50-1_6-yes1-4.cnf")
	defer file.Close()
	b, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(b)
	s := sat.NewSolver()
	options := sat.DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	result, _ := s.Solve(options)
	fmt.Println("Result", result)
	// fmt.Println("  ", s.assigns)
}

func TestDimacs04(t *testing.T) {
	file, _ := os.Open("testdata/bf0432-007.cnf")
	defer file.Close()
	b, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(b)
	s := sat.NewSolver()
	options := sat.DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	result, _ := s.Solve(options)
	fmt.Println("Result", result)
	// fmt.Println("  ", s.assigns)
}

func TestDimacs05(t *testing.T) {
	file, _ := os.Open("../testdata/satlib/unsat-dimacs-dubois/dubois50.cnf")
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(buf)
	s := sat.NewSolver()
	options := sat.DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	fmt.Println(s.Solve(options))
}
