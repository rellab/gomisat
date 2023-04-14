package esat

import (
	"com.github/rellab/gomisat/pkg/dimacs"
	"fmt"
	"io"
	_ "io/ioutil"
	"log"
	"os"
	"testing"
)

func TestEsat01(t *testing.T) {
	cs := [][]int{
		[]int{1, 2},
		[]int{-1},
	}
	fmt.Println(cs)
	s := MkESat(cs)
	fmt.Println(s)
	for _, c := range s.clauses {
		fmt.Println(c)
	}
}

func TestEsat02(t *testing.T) {
	cs := [][]int{
		[]int{1, 2},
		[]int{-1},
	}
	fmt.Println(cs)
	s := MkESat(cs)
	fmt.Println(s)
	for _, c := range s.clauses {
		fmt.Println(c)
	}
	ans := s.MkAssign([]int{1, 2})
	fmt.Println(ans)
}

func TestEsat03(t *testing.T) {
	cs := [][]int{
		[]int{1, 2},
		[]int{-1},
	}
	fmt.Println(cs)
	s := MkESat(cs)
	fmt.Println(s)
	for _, c := range s.clauses {
		fmt.Println(c)
	}
	ans := s.MkAssign([]int{2, 1})
	fmt.Println(ans)
}

func TestEsat04(t *testing.T) {
	cs := [][]int{
		[]int{-1, -3, -4},
		[]int{2, 3, -4},
		[]int{1, -2, 4},
		[]int{1, 3, 4},
		[]int{-1, 2, -3},
		[]int{-4},
	}
	fmt.Println(cs)
	s := MkESat(cs)
	fmt.Println(s)
	for _, c := range s.clauses {
		fmt.Println(c)
	}
	ans := s.MkAssign([]int{1, 2, 3, 4})
	fmt.Println(ans)
}

func TestEsat05(t *testing.T) {
	file, _ := os.Open("../../testdata/satlib/unsat-dimacs-dubois/dubois50.cnf")
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(buf)
	s := MkESat(cs)
	// fmt.Println(s)
	ans := s.MkAssign(s.vars)
	fmt.Println(ans)
	ok := 0
	for _, c := range s.clauses {
		for _, p := range c.lits {
			if p > 0 {
				if ans[p] == true {
					ok += 1
					goto NextClause
				}
			} else {
				if ans[-p] == false {
					ok += 1
					goto NextClause
				}
			}
		}
		fmt.Println("fail ", c)
	NextClause:
	}
	fmt.Println(ok)
}

func TestEsat06(t *testing.T) {
	// file, _ := os.Open("../../testdata/satlib/unsat-dimacs-dubois/dubois50.cnf")
	file, _ := os.Open("../../testdata/satlib/f3200.cnf")
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(buf)
	s := MkESat(cs)
	// fmt.Println(s)
	v := make(map[int]struct{})
	for _, x := range s.vars {
		v[x] = struct{}{}
	}
	ans := s.MkAssign2(v)
	fmt.Println(ans)
	ok := 0
	for _, c := range s.clauses {
		for _, p := range c.lits {
			if p > 0 {
				if ans[p] == true {
					ok += 1
					goto NextClause
				}
			} else {
				if ans[-p] == false {
					ok += 1
					goto NextClause
				}
			}
		}
		// fmt.Println("fail ", c)
	NextClause:
	}
	fmt.Println(ok)
}

func TestEsat07(t *testing.T) {
	file, _ := os.Open("../../testdata/satlib/f3200.cnf")
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := dimacs.ParseDimacs(buf)
	// for _, c := range cs {
	// 	fmt.Println(c)
	// }
	s := MkESat(cs)
	// fmt.Println(s)
	ans := s.MkAssign3(s.vars)
	fmt.Println(ans)
	ok := 0
	for _, c := range s.clauses {
		// fmt.Print(c)
		for _, p := range c.lits {
			if p > 0 {
				if ans[p] == true {
					ok += 1
					// fmt.Println("Clause is True with", p)
					goto NextClause
				}
			} else if p < 0 {
				if ans[-p] == false {
					ok += 1
					// fmt.Println("Clause is True with", p)
					goto NextClause
				}
			} else {
				log.Println("error")
			}
		}
		// fmt.Println("Clause is False", c)
	NextClause:
	}
	fmt.Println(ok)
}
