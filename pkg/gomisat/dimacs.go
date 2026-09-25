package gomisat

import (
	"bytes"
	"fmt"
	"strconv"
)

// CNF is a formula read from a DIMACS file. NumVars and NumClauses are what the
// header declared, which is not necessarily what Clauses contains: a header may
// declare variables that no clause mentions, and some published instances have a
// clause count that disagrees with the file.
type CNF struct {
	NumVars    int
	NumClauses int
	Clauses    [][]int64
}

// ParseDimacsCNF reads a formula in DIMACS CNF format.
//
// Clauses are read as a stream of integers terminated by 0, so a clause may span
// several lines and a line may hold several clauses; the family of inductive
// inference instances in SATLIB needs the former. Lines starting with 'c' are
// comments and may appear anywhere. A line starting with '%' ends the formula,
// which is how the SATLIB random instances mark their end (they are followed by a
// lone 0 that is not part of the formula).
func ParseDimacsCNF(b []byte) (*CNF, error) {
	cnf := &CNF{Clauses: make([][]int64, 0, 1024)}
	var clause []int64
	headerSeen := false
	line := 0

	for len(b) > 0 {
		var row []byte
		if i := bytes.IndexByte(b, '\n'); i >= 0 {
			row, b = b[:i], b[i+1:]
		} else {
			row, b = b, nil
		}
		line++
		row = bytes.TrimSpace(row)
		if len(row) == 0 {
			continue
		}

		switch row[0] {
		case 'c':
			continue
		case '%':
			b = nil
			continue
		case 'p':
			fields := bytes.Fields(row)
			if len(fields) < 4 || string(fields[1]) != "cnf" {
				return nil, fmt.Errorf("gomisat: line %d: malformed header %q", line, row)
			}
			var err error
			if cnf.NumVars, err = strconv.Atoi(string(fields[2])); err != nil {
				return nil, fmt.Errorf("gomisat: line %d: bad variable count %q", line, fields[2])
			}
			if cnf.NumClauses, err = strconv.Atoi(string(fields[3])); err != nil {
				return nil, fmt.Errorf("gomisat: line %d: bad clause count %q", line, fields[3])
			}
			headerSeen = true
			continue
		}

		if headerSeen == false {
			return nil, fmt.Errorf("gomisat: line %d: literals before the header", line)
		}
		for _, field := range bytes.Fields(row) {
			v, err := strconv.ParseInt(string(field), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("gomisat: line %d: %q is not a literal", line, field)
			}
			if v == 0 {
				// A 0 with nothing in front of it is treated as separator noise
				// rather than as the empty clause. No benchmark ships the empty
				// clause, while stray zeros do occur.
				if len(clause) > 0 {
					cnf.Clauses = append(cnf.Clauses, clause)
					clause = nil
				}
				continue
			}
			clause = append(clause, v)
		}
	}
	if headerSeen == false {
		return nil, fmt.Errorf("gomisat: no 'p cnf' header")
	}
	// An unterminated final clause is accepted; several older files end that way.
	if len(clause) > 0 {
		cnf.Clauses = append(cnf.Clauses, clause)
	}
	return cnf, nil
}

// ParseDimacs reads a formula and returns its clauses.
func ParseDimacs(b []byte) ([][]int64, error) {
	cnf, err := ParseDimacsCNF(b)
	if err != nil {
		return nil, err
	}
	return cnf.Clauses, nil
}

// AddCNF adds every clause of a parsed formula, and creates the variables the
// header declares even when no clause mentions them. That matters as soon as
// assignments are counted rather than just found.
func (s *Solver) AddCNF(cnf *CNF, options *SolverOptions) {
	for _, clause := range cnf.Clauses {
		s.AddClauseFromCode(clause, options)
	}
	if cnf.NumVars > 0 {
		s.addVar(int64(cnf.NumVars-1), options)
	}
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
