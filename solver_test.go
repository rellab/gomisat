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
