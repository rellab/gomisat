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

	// Learnt clauses only. See lbd.go.
	lbd     int        // number of distinct decision levels, when it was last measured
	tier    clauseTier // which tier the clause is managed in
	touched uint64     // conflict count when it last took part in conflict analysis
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
	// c can only subsume d if c is the shorter clause.
	if len(d.lits) < len(c.lits) || c.abs & ^d.abs != 0 {
		return LitUndef, ErrLitError
	}
	ret := LitUndef
	for _, x := range c.lits {
		pos, neg := findLit(x, d.lits)
		switch {
		case pos:
			// x occurs in d as is, nothing to record.
		case neg && ret == LitUndef:
			// x occurs negated: c subsumes d only after resolving on x
			// (self-subsuming resolution). At most one such literal is
			// allowed; a second one means c does not subsume d.
			ret = x
		default:
			return LitUndef, ErrLitError
		}
	}
	return ret, nil
}

// findLit reports whether x occurs in ps as is (pos) and whether it occurs
// negated (neg). It is called from Subsumes only.
func findLit(x Lit, ps []Lit) (pos bool, neg bool) {
	for _, y := range ps {
		if x == y {
			pos = true
		} else if x == y.Not() {
			neg = true
		}
	}
	return pos, neg
}
