package gomisat

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrAssertError error = errors.New("Assertion is failed.")
)

// CRef refers to a clause held by the solver's arena. It is an index, not a
// pointer, and it never changes for the lifetime of the clause.
type CRef uint32

// CRefUndef stands for the absence of a clause.
const CRefUndef CRef = ^CRef(0)

const (
	clauseLearnt uint8 = 1 << iota
	clauseHasExtra
	clauseDead
)

// clauseMeta is the fixed-size part of a clause. The literals are not here: they
// live in the arena's literal store, so the database is two contiguous arrays
// rather than one heap object plus one literal array per clause.
type clauseMeta struct {
	begin    uint32 // index of the first literal in clauseArena.lits
	size     uint32 // number of literals
	activity float32
	lbd      int32
	touched  uint32 // conflict count when the clause last took part in analysis
	flags    uint8
	tier     clauseTier
}

// clauseArena owns every clause, problem and learnt alike.
//
// A CRef indexes meta and stays valid until the clause is deleted, which is what
// makes compaction cheap: only the literal store is rewritten, so no watcher, no
// reason and no clause list has to be fixed up. MiniSat, which hands out offsets
// into the literal store itself, has to relocate every reference instead.
//
// The slice returned by Lits aliases the literal store, so it must not be held
// across an alloc or a compact.
type clauseArena struct {
	meta []clauseMeta
	lits []Lit
	// Deletion is lazy: a deleted clause is flagged and its slot is put aside,
	// but the slot may only be handed out again once no watch list refers to it
	// any more. Recycling it earlier would silently turn a stale watcher into a
	// watcher of a different clause.
	pending []CRef
	free    []CRef
	wasted  uint32 // literal slots held by deleted clauses
}

func newClauseArena() *clauseArena {
	return &clauseArena{
		meta: make([]clauseMeta, 0, 1024),
		lits: make([]Lit, 0, 8192),
	}
}

// alloc copies the literals into the store and returns a reference to the new
// clause.
func (a *clauseArena) alloc(lits []Lit, learnt bool, hasExtra bool) CRef {
	begin := uint32(len(a.lits))
	a.lits = append(a.lits, lits...)

	var flags uint8
	if learnt {
		flags |= clauseLearnt
	}
	if hasExtra {
		flags |= clauseHasExtra
	}
	m := clauseMeta{begin: begin, size: uint32(len(lits)), flags: flags}

	if n := len(a.free); n > 0 {
		c := a.free[n-1]
		a.free = a.free[:n-1]
		a.meta[c] = m
		return c
	}
	a.meta = append(a.meta, m)
	return CRef(len(a.meta) - 1)
}

// Lits returns the literals of a clause as a mutable view into the literal
// store. Reordering them in place, which propagation does, is intended.
func (a *clauseArena) Lits(c CRef) []Lit {
	m := &a.meta[c]
	return a.lits[m.begin : m.begin+m.size : m.begin+m.size]
}

func (a *clauseArena) Size(c CRef) int    { return int(a.meta[c].size) }
func (a *clauseArena) Learnt(c CRef) bool { return a.meta[c].flags&clauseLearnt != 0 }
func (a *clauseArena) Dead(c CRef) bool   { return a.meta[c].flags&clauseDead != 0 }

// markDead flags a clause as deleted without touching any watch list. The
// literals stay where they are until the next compaction, and the metadata slot
// stays reserved until the watch lists have been swept.
func (a *clauseArena) markDead(c CRef) bool {
	m := &a.meta[c]
	if m.flags&clauseDead != 0 {
		return false
	}
	m.flags |= clauseDead
	a.wasted += m.size
	a.pending = append(a.pending, c)
	return true
}

// recycle makes the slots of the deleted clauses available again. It is only safe
// once nothing refers to them any more, which is why the arena does not do it on
// its own; see Solver.collectGarbage.
func (a *clauseArena) recycle() {
	a.free = append(a.free, a.pending...)
	a.pending = a.pending[:0]
}

// wastedFraction is the share of the literal store held by deleted clauses.
func (a *clauseArena) wastedFraction() float64 {
	if len(a.lits) == 0 {
		return 0
	}
	return float64(a.wasted) / float64(len(a.lits))
}

// compact rewrites the literal store, dropping the literals of deleted clauses.
// References survive untouched because they are meta indices.
func (a *clauseArena) compact() {
	lits := make([]Lit, 0, len(a.lits)-int(a.wasted))
	for c := range a.meta {
		m := &a.meta[c]
		if m.flags&clauseDead != 0 {
			continue
		}
		begin := uint32(len(lits))
		lits = append(lits, a.lits[m.begin:m.begin+m.size]...)
		m.begin = begin
	}
	a.lits = lits
	a.wasted = 0
}

func (a *clauseArena) String(c CRef) string {
	if c == CRefUndef {
		return "[]"
	}
	lits := a.Lits(c)
	parts := make([]string, 0, len(lits))
	for _, p := range lits {
		parts = append(parts, p.String())
	}
	return "[" + strings.Join(parts, ",") + "] (" + fmt.Sprint(uint32(c)) + ")"
}

// abstraction is a one-word summary of the variables of a clause, used to reject
// most subsumption candidates without looking at their literals. MiniSat keeps it
// in the clause; here it is recomputed, because it is only needed by subsumption
// and not on any hot path.
func (a *clauseArena) abstraction(c CRef) uint64 {
	abst := uint64(0)
	for _, p := range a.Lits(c) {
		abst |= 0x01 << (p.Var() & 0x3f)
	}
	return abst
}

// subsumes reports whether clause c subsumes clause d.
//
// It returns LitUndef when c subsumes d outright, the literal to resolve on when
// c subsumes d after resolving on it (self-subsuming resolution), and ErrLitError
// when c does not subsume d.
func (a *clauseArena) subsumes(c, d CRef) (Lit, error) {
	if a.Learnt(c) || a.Learnt(d) ||
		a.meta[c].flags&clauseHasExtra == 0 || a.meta[d].flags&clauseHasExtra == 0 {
		return LitUndef, ErrAssertError
	}
	// c can only subsume d if c is the shorter clause.
	if a.Size(d) < a.Size(c) || a.abstraction(c) & ^a.abstraction(d) != 0 {
		return LitUndef, ErrLitError
	}
	dLits := a.Lits(d)
	ret := LitUndef
	for _, x := range a.Lits(c) {
		pos, neg := findLit(x, dLits)
		switch {
		case pos:
			// x occurs in d as is, nothing to record.
		case neg && ret == LitUndef:
			// x occurs negated: c subsumes d only after resolving on x. At most
			// one such literal is allowed; a second one means c does not subsume
			// d at all.
			ret = x
		default:
			return LitUndef, ErrLitError
		}
	}
	return ret, nil
}

// findLit reports whether x occurs in ps as is (pos) and whether it occurs
// negated (neg). It is called from subsumes only.
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

// shrink drops the last n literals of a clause. The slots they occupied become
// waste until the next compaction.
func (a *clauseArena) shrink(c CRef, n int) {
	m := &a.meta[c]
	if n <= 0 || uint32(n) > m.size {
		return
	}
	m.size -= uint32(n)
	a.wasted += uint32(n)
}
