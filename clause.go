package gomisat

import (
	"errors"
	"fmt"
	_ "log"
	_ "strconv"
	"strings"
)

var (
	ErrAssertError error = errors.New("Assertion is failed.")
)

type ClauseHeader struct {
	learnt   bool
	hasExtra bool
	reloced  bool
}

type Clause struct {
	header   ClauseHeader
	activity float64
	abs      uint64
	lits     []Lit
}

func (c *Clause) String() string {
	s := make([]string, 0, len(c.lits))
	for _, x := range c.lits {
		s = append(s, x.String())
	}
	return "[" + strings.Join(s, ",") + "] (" + fmt.Sprintf("%p", c) + ")"
}

func MkClause(ps []Lit, useExtra bool, learnt bool) *Clause {
	c := &Clause{
		header: ClauseHeader{
			learnt:   learnt,
			hasExtra: useExtra,
			reloced:  false,
		},
		activity: 0.0,
		abs:      0,
		lits:     ps,
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
	return LitUndef, nil
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
