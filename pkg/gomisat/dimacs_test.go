package gomisat

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

func TestDimacs01(t *testing.T) {
	cs, _ := ParseDimacs([]byte(`
	p cnf 5 6
	4 5 6 3 0
	-1 2 1 0
	`))
	for _, x := range cs {
		fmt.Println(x)
	}
}

func TestDimacs02(t *testing.T) {
	file, _ := os.Open("../../testdata/aim-100-1_6-no-1.cnf")
	defer file.Close()
	b, _ := io.ReadAll(file)
	cs, _ := ParseDimacs(b)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	result := s.Solve(options)
	fmt.Println("Result", result)
}

func TestDimacs03(t *testing.T) {
	file, _ := os.Open("../../testdata/aim-50-1_6-yes1-4.cnf")
	defer file.Close()
	b, _ := io.ReadAll(file)
	cs, _ := ParseDimacs(b)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	result := s.Solve(options)
	fmt.Println("Result", result)
	fmt.Println("  ", s.assigns)
}

func TestDimacs04(t *testing.T) {
	log.SetOutput(io.Discard)
	file, _ := os.Open("../../testdata/bf0432-007.cnf")
	defer file.Close()
	b, _ := io.ReadAll(file)
	cs, _ := ParseDimacs(b)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	result := s.Solve(options)
	fmt.Println("Result", result)
	fmt.Println("  ", s.assigns)
}
