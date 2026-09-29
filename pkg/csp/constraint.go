package csp

import (
	"fmt"
	"log"
	"strconv"
)

type operatorType int

const (
	opAnd operatorType = iota + 1
	opOr
	CSPOperatorLeZero
	CSPOperatorGeZero
	CSPOperatorEqZero
	CSPOperatorNeZero
)

type Constraint interface {
	Not() Constraint
	ToLeZero() Constraint
	Decomp([]*IntVar) (Constraint, []*IntVar)
	tocnf(cnf []Clause, auxvars []*BoolVar) ([]Clause, []*BoolVar)
	flattenOr(cs []Constraint) []Constraint
	tseitin(first Clause, cs []Constraint, auxvars []*BoolVar) (Clause, []Constraint, []*BoolVar)
}

// Comparator
// This is one constraint using one of the comparators <= (Le), >= (Ge), == (Eq), != (Ne).
// The fundamental form becomes
//
//	a_1 x_1 + a_2 x_2 + ... + a_n x_n (comparator) 0
//
// where the right-hand side is Sum in the program
type Comparator struct {
	op operatorType
	s  *Sum
}

func LeZero(s *Sum) *Comparator {
	return &Comparator{
		op: CSPOperatorLeZero,
		s:  s,
	}
}

func GeZero(s *Sum) *Comparator {
	return &Comparator{
		op: CSPOperatorGeZero,
		s:  s,
	}
}

func EqZero(s *Sum) *Comparator {
	return &Comparator{
		op: CSPOperatorEqZero,
		s:  s,
	}
}

func NeZero(s *Sum) *Comparator {
	return &Comparator{
		op: CSPOperatorNeZero,
		s:  s,
	}
}

func (c *Comparator) String() string {
	switch c.op {
	case CSPOperatorEqZero:
		return fmt.Sprintf("EqZero(%s)", c.s)
	case CSPOperatorNeZero:
		return fmt.Sprintf("NeZero(%s)", c.s)
	case CSPOperatorLeZero:
		return fmt.Sprintf("LeZero(%s)", c.s)
	case CSPOperatorGeZero:
		return fmt.Sprintf("GeZero(%s)", c.s)
	default:
		log.Fatal("Error: operator does not exist")
		return ""
	}
}

// Operator
// This is one constraint using logical operators (AND, OR, IMP, IFF)
//

type Operator struct {
	op   operatorType
	args []Constraint
}

func And(args ...Constraint) *Operator {
	return &Operator{
		op:   opAnd,
		args: args,
	}
}

func Or(args ...Constraint) *Operator {
	return &Operator{
		op:   opOr,
		args: args,
	}
}

func Imp(x, y Constraint) Constraint {
	return Or(x.Not(), y)
}

func Iff(x, y Constraint) Constraint {
	return And(Or(x.Not(), y), Or(x, y.Not()))
}

func (c Operator) String() string {
	switch c.op {
	case opAnd:
		str := "AND("
		for _, x := range c.args {
			str += fmt.Sprintf("%s,", x)
		}
		return str + ")"
	case opOr:
		str := "OR("
		for _, x := range c.args {
			str += fmt.Sprintf("%s,", x)
		}
		return str + ")"
	default:
		log.Fatal("Error: operator does not exist")
		return ""
	}
}

// BoolNot
// This is a negation of bool var

type BoolNot struct {
	b *BoolVar
}

func (x *BoolNot) String() string {
	var s string
	if x.b.aux == false {
		s = "!b" + strconv.Itoa(int(x.b.id))
	} else {
		s = "!ab" + strconv.Itoa(int(x.b.id))
	}
	return s
}

// Not
// The method is to take the negation
func (c *Comparator) Not() Constraint {
	s := c.s.copy()
	switch c.op {
	case CSPOperatorLeZero:
		s.addConst(-1)
		return GeZero(s)
	case CSPOperatorGeZero:
		s.addConst(1)
		return LeZero(s)
	case CSPOperatorEqZero:
		return NeZero(s)
	case CSPOperatorNeZero:
		return EqZero(s)
	default:
		log.Fatal("Model Operator is invalid.")
		return nil
	}
}

func (c *Operator) Not() Constraint {
	newargs := make([]Constraint, len(c.args))
	for i, x := range c.args {
		newargs[i] = x.Not()
	}
	switch c.op {
	case opAnd:
		return Or(newargs...)
	case opOr:
		return And(newargs...)
	default:
		log.Fatal("Model Operator is invalid.")
		return nil
	}
}

func (b *BoolVar) Not() Constraint {
	return &BoolNot{b}
}

func (b *BoolNot) Not() Constraint {
	return b.b
}

// ToLeZero
// The method is to change the Comparator to the forms using only Sum <= 0
func (c *Comparator) ToLeZero() Constraint {
	switch c.op {
	case CSPOperatorEqZero:
		s1 := c.s.copy()
		s2 := c.s.copy()
		return And(LeZero(s1), LeZero(s2.neg()))
	case CSPOperatorNeZero:
		s1 := c.s.copy()
		s2 := c.s.copy()
		return Or(LeZero(s1.addConst(1)), LeZero(s2.neg().addConst(1)))
	case CSPOperatorGeZero:
		s1 := c.s.copy()
		return LeZero(s1.neg())
	case CSPOperatorLeZero:
		s1 := c.s.copy()
		return LeZero(s1)
	default:
		log.Fatal("Model Operator is invalid.")
		return nil
	}
}

func (c *Operator) ToLeZero() Constraint {
	newargs := make([]Constraint, 0, len(c.args))
	for _, x := range c.args {
		newargs = append(newargs, x.ToLeZero())
	}
	switch c.op {
	case opAnd:
		return And(newargs...)
	case opOr:
		return Or(newargs...)
	default:
		log.Fatal("Model Operator is invalid.")
		return nil
	}
}

func (b *BoolVar) ToLeZero() Constraint {
	return b
}

func (b *BoolNot) ToLeZero() Constraint {
	return b
}
