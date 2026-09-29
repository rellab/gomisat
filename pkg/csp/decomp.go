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

func sortVars(coef map[*IntVar]int) []*IntVar {
	vars := make([]*IntVar, 0, len(coef))
	for k, _ := range coef {
		vars = append(vars, k)
	}
	sort.Slice(vars, func(i, j int) bool {
		s1 := vars[i].domain.size()
		s2 := vars[j].domain.size()
		if s1 == s2 {
			k1 := abs(coef[vars[i]])
			k2 := abs(coef[vars[j]])
			return k1 > k2
		} else {
			return s1 < s2
		}
	})
	return vars
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
