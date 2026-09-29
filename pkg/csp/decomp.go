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

// decompSum breaks a linear constraint into terms of at most three variables by
// repeatedly replacing the two variables with the smallest domains by one auxiliary
// variable standing for their weighted sum.
//
// op is the comparison the sum is under, and it is used to throw away the part of
// each auxiliary domain that cannot lead to a solution. Without that the encoding
// ignores the bound almost entirely: a ten-variable sum over 0..9 took 3368 clauses
// whether the bound was 1 or 200, when a bound of 1 rules out every partial sum
// above it.
//
// Dropping those values keeps both the solutions and their number. A pair whose
// weighted sum lies above the cap makes the constraint fail for every value of the
// remaining terms, so no solution is lost; and the auxiliary stays *equal* to the
// sum it stands for rather than merely bounding it, so each solution still extends
// to exactly one assignment of the auxiliaries and the count is unchanged. The
// weaker "auxiliary bounds the sum" decomposition would allow a tighter domain and
// would break counting, which is why it is not used.
func decompSum(s *Sum, auxvars []*IntVar, op operatorType) ([]Constraint, []*IntVar) {
	cs := make([]Constraint, 0)
	for s.size() > 3 {
		vars := sortVars(s.coef)
		x, y := vars[0], vars[1]
		a, b := s.coef[x], s.coef[y]
		d := crossApply(x.domain, y.domain, func(x, y int) int {
			return a*x + b*y
		})
		d = narrow(d, s, x, y, op)
		z := newAuxIntVar(len(auxvars), d)
		auxvars = append(auxvars, z)
		coef := map[*IntVar]int{x: a, y: b, z: -1}
		f := NewSum(coef, 0)
		s.sub(f)
		cs = append(cs, EqZero(f))
	}
	return cs, auxvars
}

// narrow drops the values of a prospective auxiliary domain that cannot take part
// in a solution. After the substitution the auxiliary appears in the sum with
// coefficient one, so the constraint reads "auxiliary + rest <= 0" or the same with
// >=, where rest is what the sum holds apart from the two variables being replaced.
//
// At least one value is always kept: if even the best of them fails, the constraint
// is unsatisfiable and the encoder reports that in its own way.
func narrow(d []int, s *Sum, x, y *IntVar, op operatorType) []int {
	restLo, restHi := s.bounds(x, y)
	keep := func(ok func(v int) bool) []int {
		out := make([]int, 0, len(d))
		for _, v := range d {
			if ok(v) {
				out = append(out, v)
			}
		}
		if len(out) == 0 {
			return d[:1]
		}
		return out
	}
	switch op {
	case CSPOperatorLeZero:
		// v + rest <= 0 has to be possible, so v <= -restLo.
		return keep(func(v int) bool { return v+restLo <= 0 })
	case CSPOperatorGeZero:
		return keep(func(v int) bool { return v+restHi >= 0 })
	case CSPOperatorEqZero:
		return keep(func(v int) bool { return v+restLo <= 0 && v+restHi >= 0 })
	}
	// A disequality rules nothing out.
	return d
}

func (c *Comparator) Decomp(auxvars []*IntVar) (Constraint, []*IntVar) {
	var cs []Constraint
	switch c.op {
	case CSPOperatorEqZero:
		cs, auxvars = decompSum(c.s, auxvars, CSPOperatorEqZero)
		cs = append(cs, EqZero(c.s))
	case CSPOperatorNeZero:
		cs, auxvars = decompSum(c.s, auxvars, CSPOperatorNeZero)
		cs = append(cs, NeZero(c.s))
	case CSPOperatorLeZero:
		cs, auxvars = decompSum(c.s, auxvars, CSPOperatorLeZero)
		cs = append(cs, LeZero(c.s))
	case CSPOperatorGeZero:
		cs, auxvars = decompSum(c.s, auxvars, CSPOperatorGeZero)
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
