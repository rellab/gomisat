package gomisat

import (
	"bufio"
	"bytes"
	"log"
	"strconv"
	"strings"
)

func ParseDimacs(b []byte) ([][]int64, error) {
	in := bytes.NewBuffer(b)
	s := bufio.NewScanner(in)
	var nlits, nclauses int
	for s.Scan() {
		a := strings.Fields(s.Text())
		if len(a) >= 1 && a[0] == "c" { // comment line
			continue
		}
		if len(a) >= 2 && a[0] == "p" && a[1] == "cnf" {
			nlits, _ = strconv.Atoi(a[2])
			nclauses, _ = strconv.Atoi(a[3])
			log.Println("Number of literals:", nlits)
			log.Println("Number of clauses:", nclauses)
			break
		}
	}
	clauses := make([][]int64, 0)
	for s.Scan() {
		a := strings.Fields(s.Text())
		if len(a) >= 2 {
			lits := make([]int64, 0, len(a)-1)
			if len(a) >= 1 && a[0] == "c" { // comment line
				continue
			}
			for _, x := range a {
				if v, err := strconv.Atoi(x); err == nil {
					if v != 0 {
						lits = append(lits, int64(v))
					}
				} else {
					log.Println("Atoi fails", err)
				}
			}
			clauses = append(clauses, lits)
		}
	}
	if len(clauses) != nclauses {
		log.Println("Did not match the number of clauses generated to the number of clauses defined on the header of DIMCS:", len(clauses))
	}
	return clauses, nil
}

func (s *Solver) AddClauseFromCode(codes []int64, options *SolverOptions) {
	lits := make([]Lit, 0, len(codes))
	for _, v := range codes {
		switch {
		case v > 0:
			s.addVar(v-1, options) // v starts with 0
			lits = append(lits, MkLit(Var(v-1), false))
		case v < 0:
			s.addVar(-(v + 1), options) // v starts with 0
			lits = append(lits, MkLit(Var(-(v+1)), true))
		default:
		}
	}
	s.AddClause(lits...)
}

// add a variable from a general int64
// This function is called from AddClauseFromCode only
func (s *Solver) addVar(v int64, options *SolverOptions) {
	for v >= int64(s.nextVar) {
		s.NewVar(true, options)
	}
}
