package csp

import (
	// "fmt"
	"log"
	"sort"
)

func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

// sortVars orders the variables of a sum: smaller domains first, then larger
// coefficients, then by identity.
//
// That last tie-break is not cosmetic. The variables come out of a map, whose
// iteration order Go randomises, and the first two keys leave many ties in a
// typical model -- equal domains and coefficients from a small set. Without a total
// order the decomposition differed from run to run: ten encodings of the same
// twenty-variable model produced nine different formulas, between 20 190 and 35 060
// clauses. An encoding that changes size by three quarters between runs cannot be
// measured, and a library whose output depends on allocation order is a poor
// neighbour.
func sortVars(coef map[*IntVar]int) []*IntVar {
	vars := make([]*IntVar, 0, len(coef))
	for k := range coef {
		vars = append(vars, k)
	}
	sortVarSlice(vars, coef)
	return vars
}

// sortVarSlice is the ordering itself, shared with the encoder so that the two
// cannot drift apart.
func sortVarSlice(vars []*IntVar, coef map[*IntVar]int) {
	sort.Slice(vars, func(i, j int) bool {
		if s1, s2 := vars[i].domain.size(), vars[j].domain.size(); s1 != s2 {
			return s1 < s2
		}
		if k1, k2 := abs(coef[vars[i]]), abs(coef[vars[j]]); k1 != k2 {
			return k1 > k2
		}
		return vars[i].id < vars[j].id
	})
}

// decompSum
// The function is to decompose a linear constraint to up to three terms
func decompSum(s *Sum, auxvars []*IntVar) ([]Constraint, []*IntVar) {
	cs := make([]Constraint, 0)
	for s.size() > 3 {
		vars := sortVars(s.coef)
		x, y := vars[0], vars[1]
		a, b := s.coef[x], s.coef[y]
		// d := x.domain.copy()
		d := crossApply(x.domain, y.domain, func(x, y int) int {
			return a*x + b*y
		})
		z := newAuxIntVar(len(auxvars), d)
		auxvars = append(auxvars, z)
		coef := map[*IntVar]int{x: a, y: b, z: -1}
		f := NewSum(coef, 0)
		s.sub(f)
		cs = append(cs, EqZero(f))
	}
	return cs, auxvars
}

func (c *Comparator) Decomp(auxvars []*IntVar) (Constraint, []*IntVar) {
	var cs []Constraint
	switch c.op {
	case CSPOperatorEqZero:
		cs, auxvars = decompSum(c.s, auxvars)
		cs = append(cs, EqZero(c.s))
	case CSPOperatorNeZero:
		cs, auxvars = decompSum(c.s, auxvars)
		cs = append(cs, NeZero(c.s))
	case CSPOperatorLeZero:
		cs, auxvars = decompSum(c.s, auxvars)
		cs = append(cs, LeZero(c.s))
	case CSPOperatorGeZero:
		cs, auxvars = decompSum(c.s, auxvars)
		cs = append(cs, GeZero(c.s))
	default:
		log.Fatal("Error: operator does not exist")
		return nil, auxvars
	}
	if len(cs) == 1 {
		return cs[0], auxvars
	} else {
		return And(cs...), auxvars
	}
}

func (c *Operator) Decomp(auxvars []*IntVar) (Constraint, []*IntVar) {
	switch c.op {
	case opAnd:
		args := make([]Constraint, len(c.args))
		for i, x := range c.args {
			args[i], auxvars = x.Decomp(auxvars)
		}
		return And(args...), auxvars
	case opOr:
		args := make([]Constraint, len(c.args))
		for i, x := range c.args {
			args[i], auxvars = x.Decomp(auxvars)
		}
		return Or(args...), auxvars
	default:
		log.Fatal("Error: operator does not exist")
		return nil, auxvars
	}
}

func (b *BoolVar) Decomp(auxvars []*IntVar) (Constraint, []*IntVar) {
	return b, auxvars
}

func (b *BoolNot) Decomp(auxvars []*IntVar) (Constraint, []*IntVar) {
	return b, auxvars
}
