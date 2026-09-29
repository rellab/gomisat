package csp

// Order encode: http://bach.istc.kobe-u.ac.jp/sugar/docs/order-encoding/

import (
	_ "errors"
	_ "fmt"
	"log"
)

func Encode(c Clause, baseCode map[int]int) ([][]int, bool) {
	cs := [][]int{[]int{}}
	unsat := false
	var ok bool
	for _, p := range c {
		cs, ok = p.encode(cs, baseCode)
		unsat = unsat || ok
	}
	return cs, unsat
}

func applyOr(cs [][]int, p ...int) [][]int {
	for i, _ := range cs {
		cs[i] = append(cs[i], p...)
	}
	return cs
}

func (b *BoolVar) encode(codes [][]int, baseCode map[int]int) ([][]int, bool) {
	base := baseCode[b.id]
	p := base
	return applyOr(codes, p), true
}

func (b *BoolNot) encode(codes [][]int, baseCode map[int]int) ([][]int, bool) {
	base := baseCode[b.b.id]
	p := -base
	return applyOr(codes, p), true
}

func (c *Comparator) encode(codes [][]int, baseCode map[int]int) ([][]int, bool) {
	if c.op == CSPOperatorLeZero {
		vars := sortVars(c.s.coef)
		if cs, ok := encodeIntVar([][]int{}, make([]int, 0, len(vars)), vars, c.s, -c.s.b, baseCode); ok {
			// log.Println("encode", cs)
			switch {
			case len(cs) == 0: // trueLit
				return [][]int{}, true
			case len(cs) == 1: // simple
				return applyOr(codes, cs[0]...), true
			case len(codes) == 1:
				return applyOr(cs, codes[0]...), true
			default:
				result := make([][]int, 0, len(cs)*len(codes))
				for _, x := range codes {
					for _, y := range cs {
						p := make([]int, 0, len(x)+len(y))
						p = append(p, x...)
						p = append(p, y...)
						result = append(result, p)
					}
				}
				return result, true
			}
		} else {
			// falseLit
			return codes, false
		}
	} else {
		log.Fatal("SAT encoding can be applied to LeZero")
		return codes, false
	}
}

// Encoding based on the formula: a_1 x_1 + ... a_n x_n <= b
func encodeIntVar(cs [][]int, lit []int, vars []*IntVar, s *Sum, rhs int, baseCode map[int]int) ([][]int, bool) {
	// log.Println("encodeIntVar: start", "cs", cs, "vars", vars, "rhs", rhs)
	a := s.coef[vars[0]]
	base := baseCode[vars[0].id]
	domain := vars[0].domain
	var ok bool
	if len(vars) == 1 {
		return encodeLe(cs, lit, base, domain, a, rhs)
	}
	if a > 0 {
		b := domain[0]
		// log.Println("  **start encode", vars[0], b, cs)
		cs, ok = encodeIntVar(cs, lit, vars[1:], s, rhs-a*b, baseCode)
		// log.Println("  **end encode", vars[0], b, cs)
		if ok == false {
			return cs, ok
		}
		for i, b := range domain[1:] {
			// log.Println("  **start encode", vars[0], b, cs)
			cs, ok = encodeIntVar(cs, append(lit, base+i), vars[1:], s, rhs-a*b, baseCode)
			// log.Println("  **end encode", vars[0], b, cs)
			if ok == false {
				return cs, ok
			}
		}
		// log.Println("  *encode", cs)
	} else { // assume a < 0
		for i, b := range domain[:domain.size()-1] {
			cs, ok = encodeIntVar(cs, append(lit, -(base+i)), vars[1:], s, rhs-a*b, baseCode)
			if ok == false {
				return cs, ok
			}
		}
		b := domain[domain.size()-1]
		cs, ok = encodeIntVar(cs, lit, vars[1:], s, rhs-a*b, baseCode)
		if ok == false {
			return cs, ok
		}
	}
	return cs, true
}

// SATcode for a*x <= b, x in domain
func encodeLe(cs [][]int, lit []int, base int, domain DomainSet, a, b int) ([][]int, bool) {
	if a > 0 {
		c := floorDiv(b, a)
		// log.Println("base", base, "a", a, "b", b, "c", c, "domain", domain)
		pos, ok := domain.contains(c)
		switch {
		case pos == ErrDomainLower && ok == false: // falseLit
			if len(lit) == 0 {
				return cs, false
			} else {
				// return cs, true
				return append(cs, clauseCopy(lit)), true
			}
		case pos == ErrDomainUpper && ok == false: // trueLit
			return cs, true
		default:
			lit = append(lit, base+pos)
			return append(cs, clauseCopy(lit)), true
		}
	} else {
		c := ceilDiv(b, a) - 1
		// log.Println("base", base, "a", a, "b", b, "c", c, "domain", domain)
		pos, ok := domain.contains(c)
		switch {
		case pos == ErrDomainLower && ok == false:
			return cs, true
		case pos == ErrDomainUpper && ok == false:
			if len(lit) == 0 {
				return cs, false
			} else {
				// return cs, true
				return append(cs, clauseCopy(lit)), true
			}
		default:
			lit = append(lit, -(base + pos))
			return append(cs, clauseCopy(lit)), true
		}
	}
}

func clauseCopy(s []int) []int {
	t := make([]int, len(s))
	copy(t, []int(s))
	return t
}
