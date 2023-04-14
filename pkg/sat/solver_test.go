package sat

import (
	"fmt"
	_ "io/ioutil"
	_ "log"
	"testing"
)

func TestSolver03(t *testing.T) {
	cs := [][]int{
		[]int{1, 2},
		[]int{-1},
	}
	fmt.Println(cs)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	fmt.Println(s)
	fmt.Println(s.ok)
	fmt.Println(s.assigns)
}

func TestSolver04(t *testing.T) {
	cs := [][]int{
		[]int{1, 2},
		[]int{-1, 3},
	}
	fmt.Println(cs)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	fmt.Println(s)
	fmt.Println(s.ok)
	fmt.Println(s.assigns)
}

func TestSolver05(t *testing.T) {
	cs := [][]int{
		[]int{1, 2},
		[]int{-1},
	}
	fmt.Println(cs)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	fmt.Println(s)
	fmt.Println(s.ok)
	fmt.Println(s.assigns)
	for _, c := range s.clauses {
		fmt.Println(c)
	}
	s.Simplify(options)
	fmt.Println(s)
	fmt.Println(s.ok)
	fmt.Println(s.assigns)
	for _, c := range s.clauses {
		fmt.Println(c)
	}
}

func TestSolver06(t *testing.T) {
	cs := [][]int{
		[]int{1},
		[]int{-1},
	}
	fmt.Println(cs)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
		fmt.Println(s.ok)
	}
}

func TestSolver07(t *testing.T) {
	cs := [][]int{
		[]int{1, 4, -3, 6},
		[]int{5, 2},
		[]int{-1, 3, 2},
	}
	fmt.Println(cs)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify(options)
	s.search(1, options)
}

func TestSolver08(t *testing.T) {
	cs := [][]int{
		[]int{-1, -3, -4},
		[]int{2, 3, -4},
		[]int{1, -2, 4},
		[]int{1, 3, 4},
		[]int{-1, 2, -3},
		[]int{-4},
	}
	fmt.Println(cs)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify(options)
	s.Solve(options)
	fmt.Println(s.assigns)
}
