package gomisat

import (
	"fmt"
	"testing"
)

func TestSolver01(t *testing.T) {
	cs, _ := ParseDimacs([]byte(`
	p cnf 5 6
	4 5 6 5 0
	-1 2 1 0
	`))
	fmt.Println(cs)
	s := NewSolver()
	fmt.Println(s)
	s.AddClauseFromCode(cs[0])
	fmt.Println(s)
}

func TestSolver02(t *testing.T) {
	cs, _ := ParseDimacs([]byte(`
	p cnf 1 1
	1 2 0
	-1 0
	`))
	fmt.Println(cs)
	s := NewSolver()
	for _, x := range cs {
		s.AddClauseFromCode(x)
	}
	fmt.Println(s)
	fmt.Println(s.ok)
	fmt.Println(s.assigns)
}

func TestSolver03(t *testing.T) {
	cs := [][]int64{
		[]int64{1, 2},
		[]int64{-1},
	}
	fmt.Println(cs)
	s := NewSolver()
	for _, x := range cs {
		s.AddClauseFromCode(x)
	}
	fmt.Println(s)
	fmt.Println(s.ok)
	fmt.Println(s.assigns)
}

func TestSolver04(t *testing.T) {
	cs := [][]int64{
		[]int64{1, 2},
		[]int64{-1, 3},
	}
	fmt.Println(cs)
	s := NewSolver()
	for _, x := range cs {
		s.AddClauseFromCode(x)
	}
	fmt.Println(s)
	fmt.Println(s.ok)
	fmt.Println(s.assigns)
}
