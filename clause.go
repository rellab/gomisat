package gomisat

import (
	"errors"
	_ "fmt"
	_ "log"
	_ "strconv"
)

var (
	ErrAssertError error = errors.New("Assertion is failed.")
)

type ClauseHeader struct {
	mark     byte
	learnt   bool
	hasExtra bool
	reloced  bool
}

type Clause struct {
	header ClauseHeader
	act    float32
	abs    uint64
	lits   []Lit
}

func MkClause(ps []Lit, useExtra bool, learnt bool) *Clause {
	c := &Clause{
		header: ClauseHeader{
			mark:     0,
			learnt:   learnt,
			hasExtra: useExtra,
			reloced:  false,
		},
		act:  0.0,
		abs:  0,
		lits: ps,
	}
	c.CalcAbstraction()
	return c
}

// abst: it likes a hash value for the clause
func (c *Clause) CalcAbstraction() {
	abst := uint64(0)
	if c.header.hasExtra {
		for _, x := range c.lits {
			abst |= 0x01 << (x.Var() & 0x3f)
		}
	}
	c.abs = abst
}

func (c *Clause) Size() int {
	return len(c.lits)
}

func (c *Clause) Shrink(i int) {
	c.lits = c.lits[:len(c.lits)-i]
}

func (c *Clause) Pop() {
	c.Shrink(1)
}

func (c *Clause) Last() Lit {
	return c.lits[len(c.lits)-1]
}

func (c *Clause) Subsumes(d *Clause) (Lit, error) {
	if c.header.learnt == true ||
		d.header.learnt == true ||
		c.header.hasExtra == false ||
		d.header.hasExtra == false {
		return 0, ErrAssertError
	}
	if len(c.lits) < len(d.lits) || c.abs & ^d.abs != 0 {
		return 0, ErrLitError
	}
	for _, x := range c.lits {
		p, n := findLit(x, d.lits)
		switch {
		case p == false && n == true:
			return x, nil
		case p == false && n == false:
			return 0, ErrLitError
		default:
		}
	}
	return 0, ErrLitUndef
}

// The function is called in the Subsumes only
func findLit(x Lit, ps []Lit) (bool, bool) {
	for _, y := range ps {
		if x == y {
			return true, false
		} else if x == y.Not() {
			return true, false
		}
	}
	return false, false
}
