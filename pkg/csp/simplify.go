package csp

import (
	"fmt"
	"log"
)

type Literal interface {
	isSimple() bool
	encode([][]int, map[int]int) ([][]int, bool)
}

type Clause []Literal

func (c Clause) String() string {
	str := "["
	for _, v := range c {
		str += fmt.Sprintf("%s,", v)
	}
	return str + "]"
}

// tocnf: The function is to create CNF
func (c *Operator) tocnf(cnf []Clause, auxvars []*BoolVar) ([]Clause, []*BoolVar) {
	switch c.op {
	case opAnd:
		for _, x := range c.args {
			cnf, auxvars = x.tocnf(cnf, auxvars)
		}
		return cnf, auxvars
	case opOr:
		flattencs := c.flattenOr(make([]Constraint, 0))
		clause := make(Clause, 0)
		cs := make([]Constraint, 0)
		for _, x := range flattencs {
			clause, cs, auxvars = x.tseitin(clause, cs, auxvars)
		}
		cnf = append(cnf, clause)
		for _, x := range cs {
			cnf, auxvars = x.tocnf(cnf, auxvars)
		}
		return cnf, auxvars
	default:
		log.Fatal("Error: operator does not exist")
		return cnf, auxvars
	}
}

func (c *Comparator) tocnf(cnf []Clause, auxvars []*BoolVar) ([]Clause, []*BoolVar) {
	return append(cnf, Clause{c}), auxvars
}

func (b *BoolVar) tocnf(cnf []Clause, auxvars []*BoolVar) ([]Clause, []*BoolVar) {
	return append(cnf, Clause{b}), auxvars
}

func (b *BoolNot) tocnf(cnf []Clause, auxvars []*BoolVar) ([]Clause, []*BoolVar) {
	return append(cnf, Clause{b}), auxvars
}

// flattenOr: The function is to flatten list of literals for two or more OR operations.
// Example: Or(Or(a,b,c), AND(d, e)) -> [a, b, c, AND(d,e)]
func (c *Operator) flattenOr(cs []Constraint) []Constraint {
	switch c.op {
	case opAnd:
		return append(cs, c)
	case opOr:
		for _, x := range c.args {
			cs = x.flattenOr(cs)
		}
		return cs
	default:
		log.Fatal("Error: operator does not exist")
		return cs
	}
}

func (c *Comparator) flattenOr(cs []Constraint) []Constraint {
	return append(cs, c)
}

func (b *BoolVar) flattenOr(cs []Constraint) []Constraint {
	return append(cs, b)
}

func (b *BoolNot) flattenOr(cs []Constraint) []Constraint {
	return append(cs, b)
}

// tseitin: Tsetin transform
func (c *Operator) tseitin(first Clause, cs []Constraint, auxvars []*BoolVar) (Clause, []Constraint, []*BoolVar) {
	p := newAuxBoolVar(len(auxvars))
	auxvars = append(auxvars, p)
	first = append(first, p)
	switch c.op {
	case opAnd:
		for _, x := range c.args {
			cs = append(cs, Or(x, p.Not()))
		}
	case opOr:
		c.args = append(c.args, p.Not())
		cs = append(cs, Or(c.args...))
	default:
		log.Fatal("Error: operator does not exist")
	}
	return first, cs, auxvars
}

func (c *Comparator) tseitin(first Clause, cs []Constraint, auxvars []*BoolVar) (Clause, []Constraint, []*BoolVar) {
	return append(first, c), cs, auxvars
}

func (b *BoolVar) tseitin(first Clause, cs []Constraint, auxvars []*BoolVar) (Clause, []Constraint, []*BoolVar) {
	return append(first, b), cs, auxvars
}

func (b *BoolNot) tseitin(first Clause, cs []Constraint, auxvars []*BoolVar) (Clause, []Constraint, []*BoolVar) {
	return append(first, b), cs, auxvars
}

// simplify
func isSimple(c Clause) bool {
	count := 0
	for _, l := range c {
		if !l.isSimple() {
			if count++; count > 1 {
				return false
			}
		}
	}
	return true
}

func (c *Comparator) isSimple() bool {
	return c.s.size() <= 1
}

func (b *BoolVar) isSimple() bool {
	return true
}

func (b *BoolNot) isSimple() bool {
	return true
}

func simplify(c Constraint, cnf []Clause, auxvars []*BoolVar, tmp []Clause) ([]Clause, []*BoolVar) {
	tmp, auxvars = c.tocnf(tmp, auxvars)
	for _, clause := range tmp {
		if isSimple(clause) {
			cnf = append(cnf, clause)
		} else {
			first := make([]Literal, 0, len(clause))
			for _, lit := range clause {
				if lit.isSimple() {
					first = append(first, lit)
				} else {
					p := newAuxBoolVar(len(auxvars))
					q := &BoolNot{p}
					auxvars = append(auxvars, p)
					first = append(first, p)
					cnf = append(cnf, Clause{lit, q})
				}
			}
			cnf = append(cnf, Clause(first))
		}
	}
	return cnf, auxvars
}
