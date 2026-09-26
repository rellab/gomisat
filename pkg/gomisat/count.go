package gomisat

import (
	"math/big"
	"sort"
)

// Exact model counting by search with component decomposition and caching.
//
// The counter branches like a DPLL search, but after every propagation it splits
// what is left of the formula into connected components -- groups of clauses that
// share no unassigned variable -- and counts each of them separately. The models
// of the whole are then the product of the models of the parts, which is where the
// leverage of counting comes from: a formula that falls apart into k independent
// pieces costs the sum of their searches rather than the product.
//
// Every component is also memoised. The same sub-formula turns up under many
// different partial assignments, and the count only depends on which clauses are
// left and which of their variables are still unassigned, so it can be looked up.
//
// Clause learning is deliberately not used here. A learnt clause may span two
// components, and then assigning inside one of them can propagate into the other,
// which breaks the independence the product relies on. Getting that right needs
// care about which clauses may take part in propagation; this stage is about
// establishing a correct baseline that a brute-force oracle can check, so it
// propagates with the problem clauses only. See DESIGN.md, phase 2.

// CountOptions controls the counter.
type CountOptions struct {
	// UseCache memoises component counts. Off, the counter still decomposes; the
	// two must agree, which is what the regression tests check.
	UseCache bool
	// UseDecomposition splits the residual formula into components. Off, the
	// counter is a plain DPLL model counter, which is the other reference point.
	UseDecomposition bool
}

func DefaultCountOptions() *CountOptions {
	return &CountOptions{UseCache: true, UseDecomposition: true}
}

// CountStats reports what the search did.
type CountStats struct {
	Decisions  uint64
	Conflicts  uint64
	CacheHits  uint64
	CacheMiss  uint64
	Components uint64 // components produced by decomposition
	CacheSize  int
}

// component is a set of clauses that are not yet satisfied, together with the
// variables of theirs that are still unassigned.
type component struct {
	vars    []Var
	clauses []CRef
}

type counter struct {
	s       *Solver
	options *SolverOptions
	copt    *CountOptions
	cache   map[string]*big.Int
	stats   CountStats

	// scratch, reused across nodes to keep the allocation out of the hot path
	uf        []Var // union-find parent, indexed by variable
	seenVar   []uint64
	mark      uint64
	occ       []int32 // occurrences per variable, for the branching heuristic
	covered   []uint64
	coverMark uint64
	keyBuf    []byte
	allVars   []Var
}

// CountModels returns the number of assignments of all variables that satisfy
// every problem clause. The solver must not have learnt any clauses; the counter
// propagates with the problem clauses alone.
func (s *Solver) CountModels(options *SolverOptions, copt *CountOptions) (*big.Int, CountStats) {
	if copt == nil {
		copt = DefaultCountOptions()
	}
	n := s.NumVars()
	c := &counter{
		s:       s,
		options: options,
		copt:    copt,
		cache:   make(map[string]*big.Int),
		uf:      make([]Var, n),
		seenVar: make([]uint64, n),
		occ:     make([]int32, n),
		covered: make([]uint64, n),
		keyBuf:  make([]byte, 0, 256),
		allVars: make([]Var, n),
	}
	for v := 0; v < n; v++ {
		c.allVars[v] = Var(v)
	}
	total := c.run()
	c.stats.CacheSize = len(c.cache)
	return total, c.stats
}

func (c *counter) run() *big.Int {
	s := c.s
	if s.ok == false {
		return big.NewInt(0)
	}
	// Fix whatever the unit clauses force before anything else.
	if hasConflict(s.Propagate()) {
		return big.NewInt(0)
	}

	root := c.residual(s.clauses)
	return c.countAll(root, c.allVars)
}

// countAll splits a clause set into components, counts each and multiplies the
// results, then doubles the product once for every variable of scope that is
// still unassigned and no longer occurs in any clause. Those variables are free:
// both of their values extend every model of the rest.
func (c *counter) countAll(comp *component, scope []Var) *big.Int {
	parts := c.split(comp)

	// The free variables have to be settled before recursing. The mark array is
	// shared with the nested calls, which raise the mark for their own components,
	// so reading it afterwards would see their marks and not this node's.
	c.coverMark++
	mark := c.coverMark
	for _, part := range parts {
		for _, v := range part.vars {
			c.covered[v] = mark
		}
	}
	free := 0
	for _, v := range scope {
		if c.s.assigns[v] == LUndef && c.covered[v] != mark {
			free++
		}
	}

	total := big.NewInt(1)
	for _, part := range parts {
		n := c.count(part)
		if n.Sign() == 0 {
			return big.NewInt(0)
		}
		total.Mul(total, n)
	}
	if free > 0 {
		total.Lsh(total, uint(free))
	}
	return total
}

// count returns the number of assignments of comp.vars that satisfy comp.clauses
// under the current partial assignment.
func (c *counter) count(comp *component) *big.Int {
	if len(comp.clauses) == 0 {
		// Nothing left to satisfy: every assignment of the free variables counts.
		return new(big.Int).Lsh(big.NewInt(1), uint(len(comp.vars)))
	}

	var key string
	if c.copt.UseCache {
		key = c.key(comp)
		if hit, ok := c.cache[key]; ok {
			c.stats.CacheHits++
			return new(big.Int).Set(hit)
		}
		c.stats.CacheMiss++
	}

	v := c.branchVar(comp)
	total := big.NewInt(0)
	level := c.s.decisionLevel()
	for _, sign := range []bool{false, true} {
		c.stats.Decisions++
		c.s.newDecisionLevel()
		c.s.UncheckedEnqueue(MkLit(v, sign), CRefUndef)
		if hasConflict(c.s.Propagate()) {
			c.stats.Conflicts++
			c.s.cancelUntil(level, c.options)
			continue
		}
		// The branch fixed v and whatever propagation reached. What is left of
		// comp.vars is either free or inside one of the sub-components.
		sub := c.residual(comp.clauses)
		total.Add(total, c.countAll(sub, comp.vars))
		c.s.cancelUntil(level, c.options)
	}

	if c.copt.UseCache {
		c.cache[key] = new(big.Int).Set(total)
	}
	return total
}

// residual keeps the clauses that are not satisfied yet and collects their
// unassigned variables.
func (c *counter) residual(clauses []CRef) *component {
	out := &component{
		clauses: make([]CRef, 0, len(clauses)),
		vars:    make([]Var, 0, 16),
	}
	c.mark++
	for _, ref := range clauses {
		if c.s.arena.Dead(ref) {
			continue
		}
		satisfied := false
		count := 0
		for _, p := range c.s.arena.Lits(ref) {
			switch c.s.LitValue(p) {
			case LTrue:
				satisfied = true
			case LFalse:
			default:
				count++
			}
			if satisfied {
				break
			}
		}
		if satisfied || count == 0 {
			continue
		}
		out.clauses = append(out.clauses, ref)
		for _, p := range c.s.arena.Lits(ref) {
			if c.s.LitValue(p) == LTrue || c.s.LitValue(p) == LFalse {
				continue
			}
			if c.seenVar[p.Var()] != c.mark {
				c.seenVar[p.Var()] = c.mark
				out.vars = append(out.vars, p.Var())
			}
		}
	}
	sort.Slice(out.vars, func(i, j int) bool { return out.vars[i] < out.vars[j] })
	return out
}

// split partitions a component into connected components: two clauses belong
// together when they share an unassigned variable.
func (c *counter) split(comp *component) []*component {
	if c.copt.UseDecomposition == false || len(comp.clauses) <= 1 {
		if len(comp.clauses) == 0 {
			return nil
		}
		c.stats.Components++
		return []*component{comp}
	}

	for _, v := range comp.vars {
		c.uf[v] = v
	}
	for _, ref := range comp.clauses {
		first := VarUndef
		for _, p := range c.s.arena.Lits(ref) {
			if c.s.LitValue(p) == LTrue || c.s.LitValue(p) == LFalse {
				continue
			}
			if first == VarUndef {
				first = p.Var()
				continue
			}
			c.union(first, p.Var())
		}
	}

	groups := make(map[Var]*component, 4)
	for _, ref := range comp.clauses {
		root := VarUndef
		for _, p := range c.s.arena.Lits(ref) {
			if c.s.LitValue(p) == LTrue || c.s.LitValue(p) == LFalse {
				continue
			}
			root = c.find(p.Var())
			break
		}
		g := groups[root]
		if g == nil {
			g = &component{}
			groups[root] = g
		}
		g.clauses = append(g.clauses, ref)
	}
	for _, v := range comp.vars {
		root := c.find(v)
		if g := groups[root]; g != nil {
			g.vars = append(g.vars, v)
		}
	}

	out := make([]*component, 0, len(groups))
	for _, g := range groups {
		sort.Slice(g.vars, func(i, j int) bool { return g.vars[i] < g.vars[j] })
		out = append(out, g)
	}
	// A deterministic order keeps runs reproducible.
	sort.Slice(out, func(i, j int) bool { return out[i].vars[0] < out[j].vars[0] })
	c.stats.Components += uint64(len(out))
	return out
}

func (c *counter) find(v Var) Var {
	for c.uf[v] != v {
		c.uf[v] = c.uf[c.uf[v]]
		v = c.uf[v]
	}
	return v
}

func (c *counter) union(a, b Var) {
	ra, rb := c.find(a), c.find(b)
	if ra != rb {
		c.uf[ra] = rb
	}
}

// branchVar picks the variable of the component that occurs in the most residual
// clauses, which tends to break the component apart soonest.
func (c *counter) branchVar(comp *component) Var {
	for _, v := range comp.vars {
		c.occ[v] = 0
	}
	for _, ref := range comp.clauses {
		for _, p := range c.s.arena.Lits(ref) {
			if c.s.LitValue(p) == LTrue || c.s.LitValue(p) == LFalse {
				continue
			}
			c.occ[p.Var()]++
		}
	}
	best, bestCount := comp.vars[0], int32(-1)
	for _, v := range comp.vars {
		if c.occ[v] > bestCount {
			best, bestCount = v, c.occ[v]
		}
	}
	return best
}

// key identifies a component: the clauses that are left and the variables of
// theirs that are still unassigned. Those two together determine the residual
// formula, so they determine the count.
func (c *counter) key(comp *component) string {
	buf := c.keyBuf[:0]
	buf = appendUvarint(buf, uint64(len(comp.vars)))
	for _, v := range comp.vars {
		buf = appendUvarint(buf, uint64(v))
	}
	buf = appendUvarint(buf, uint64(len(comp.clauses)))
	for _, ref := range comp.clauses {
		buf = appendUvarint(buf, uint64(ref))
	}
	c.keyBuf = buf
	return string(buf)
}

func appendUvarint(buf []byte, x uint64) []byte {
	for x >= 0x80 {
		buf = append(buf, byte(x)|0x80)
		x >>= 7
	}
	return append(buf, byte(x))
}
