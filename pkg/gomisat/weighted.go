package gomisat

import (
	"math/big"
	"sort"
)

// Weighted model counting.
//
// Every literal carries a weight; the weight of an assignment is the product of
// the weights of its literals, and the weighted count is the sum over the models.
// With the weight of a component's "works" literal set to its probability and the
// complement to one minus it, the weighted count of the structure function is the
// reliability of the system. That identity is the reason this project exists.
//
// The accumulator is a big.Float rather than a float64: the product of a hundred
// component probabilities leaves the range of a double, and a reliability that
// silently underflows to zero is worse than no answer. It is also not a float64
// summed in log space, because the sums here are over models and log-space
// addition would cost accuracy on every one of them.
//
// This is a second implementation of the same recursion as count.go rather than a
// generalisation of it. That is deliberate: with unit weights the two must agree,
// which is a check no shared implementation could provide.

// Weights assigns a weight to every literal. The default is one, which makes the
// weighted count equal the model count.
type Weights struct {
	w []*big.Float
}

// NewWeights returns unit weights for numVars variables.
func NewWeights(numVars int) *Weights {
	w := &Weights{w: make([]*big.Float, 2*numVars)}
	for i := range w.w {
		w.w[i] = big.NewFloat(1)
	}
	return w
}

// Set gives variable v the weight whenTrue when it is true and whenFalse when it
// is false.
func (w *Weights) Set(v Var, whenTrue, whenFalse *big.Float) {
	w.w[MkLit(v, false)] = new(big.Float).Set(whenTrue)
	w.w[MkLit(v, true)] = new(big.Float).Set(whenFalse)
}

// SetProbability gives variable v the weight p when true and 1-p when false, which
// is how a component's reliability enters the formula.
func (w *Weights) SetProbability(v Var, p float64) {
	w.Set(v, big.NewFloat(p), big.NewFloat(1-p))
}

// Of returns the weight of a literal.
func (w *Weights) Of(p Lit) *big.Float {
	if int(p) >= len(w.w) {
		return big.NewFloat(1)
	}
	return w.w[p]
}

// sum is the weight of a variable being either value, the factor a free variable
// contributes.
func (w *Weights) sum(v Var) *big.Float {
	return new(big.Float).Add(w.Of(MkLit(v, false)), w.Of(MkLit(v, true)))
}

type weightedCounter struct {
	s       *Solver
	options *SolverOptions
	copt    *CountOptions
	weights *Weights
	cache   map[string]*big.Float
	stats   CountStats

	uf        []Var
	seenVar   []uint64
	mark      uint64
	occ       []int32
	order     []int32
	covered   []uint64
	coverMark uint64
	keyBuf    []byte
	allVars   []Var
	prec      uint
}

// WeightedCount returns the sum over the models of the product of their literal
// weights. With unit weights it is the model count.
func (s *Solver) WeightedCount(options *SolverOptions, copt *CountOptions, weights *Weights) (*big.Float, CountStats) {
	if copt == nil {
		copt = DefaultCountOptions()
	}
	n := s.NumVars()
	if weights == nil {
		weights = NewWeights(n)
	}
	prec := copt.Precision
	if prec == 0 {
		prec = 256
	}
	c := &weightedCounter{
		s:       s,
		options: options,
		copt:    copt,
		weights: weights,
		cache:   make(map[string]*big.Float),
		uf:      make([]Var, n),
		seenVar: make([]uint64, n),
		occ:     make([]int32, n),
		covered: make([]uint64, n),
		keyBuf:  make([]byte, 0, 256),
		allVars: make([]Var, n),
		prec:    prec,
	}
	for v := 0; v < n; v++ {
		c.allVars[v] = Var(v)
	}
	if copt.Branching == BranchEliminationOrder {
		c.order = s.eliminationScores(s.clauses)
	}

	total := c.run()
	c.stats.CacheSize = len(c.cache)
	return total, c.stats
}

func (c *weightedCounter) zero() *big.Float { return new(big.Float).SetPrec(c.prec) }
func (c *weightedCounter) one() *big.Float  { return new(big.Float).SetPrec(c.prec).SetInt64(1) }

func (c *weightedCounter) run() *big.Float {
	s := c.s
	if s.ok == false {
		return c.zero()
	}
	if hasConflict(s.Propagate()) {
		return c.zero()
	}
	// Whatever is fixed at the root is part of every model, so its weight is a
	// factor of the whole answer. The trail has to be taken from its beginning,
	// not from where this function found it: AddClause propagates a unit clause as
	// soon as it is added, so the root assignments are already there.
	forced := c.trailWeight(0)

	root := c.residual(s.clauses)
	total := c.countAll(root, c.allVars)
	return total.Mul(total, forced)
}

// trailWeight is the product of the weights of the literals put on the trail from
// position from onwards: the decision of the current branch and everything that
// propagation derived from it.
func (c *weightedCounter) trailWeight(from int) *big.Float {
	product := c.one()
	for i := from; i < len(c.s.trail); i++ {
		product.Mul(product, c.weights.Of(c.s.trail[i]))
	}
	return product
}

func (c *weightedCounter) countAll(comp *component, scope []Var) *big.Float {
	parts := c.split(comp)

	c.coverMark++
	mark := c.coverMark
	for _, part := range parts {
		for _, v := range part.vars {
			c.covered[v] = mark
		}
	}
	free := make([]Var, 0, 8)
	for _, v := range scope {
		if c.s.assigns[v] == LUndef && c.covered[v] != mark {
			free = append(free, v)
		}
	}

	total := c.one()
	for _, part := range parts {
		n := c.count(part)
		if n.Sign() == 0 {
			return c.zero()
		}
		total.Mul(total, n)
	}
	for _, v := range free {
		total.Mul(total, c.weights.sum(v))
	}
	return total
}

func (c *weightedCounter) count(comp *component) *big.Float {
	if len(comp.clauses) == 0 {
		total := c.one()
		for _, v := range comp.vars {
			total.Mul(total, c.weights.sum(v))
		}
		return total
	}

	var key string
	if c.copt.UseCache {
		key = c.key(comp)
		if hit, ok := c.cache[key]; ok {
			c.stats.CacheHits++
			return new(big.Float).Set(hit)
		}
		c.stats.CacheMiss++
	}

	v := c.branchVar(comp)
	total := c.zero()
	level := c.s.decisionLevel()
	for _, sign := range []bool{false, true} {
		c.stats.Decisions++
		before := len(c.s.trail)
		c.s.newDecisionLevel()
		c.s.UncheckedEnqueue(MkLit(v, sign), CRefUndef)
		if hasConflict(c.s.Propagate()) {
			c.stats.Conflicts++
			c.s.cancelUntil(level, c.options)
			continue
		}
		// The decision and everything propagation derived from it are fixed in
		// this branch, so their weights multiply the branch.
		branch := c.trailWeight(before)
		sub := c.residual(comp.clauses)
		branch.Mul(branch, c.countAll(sub, comp.vars))
		total.Add(total, branch)
		c.s.cancelUntil(level, c.options)
	}

	if c.copt.UseCache {
		c.cache[key] = new(big.Float).Set(total)
	}
	return total
}

// The structural helpers below are the same as the unweighted counter's. They are
// repeated rather than shared so that the two counters stay independent; see the
// note at the top of the file.

func (c *weightedCounter) residual(clauses []CRef) *component {
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
		open := 0
		for _, p := range c.s.arena.Lits(ref) {
			switch c.s.LitValue(p) {
			case LTrue:
				satisfied = true
			case LFalse:
			default:
				open++
			}
			if satisfied {
				break
			}
		}
		if satisfied || open == 0 {
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

func (c *weightedCounter) split(comp *component) []*component {
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
		if g := groups[c.find(v)]; g != nil {
			g.vars = append(g.vars, v)
		}
	}
	out := make([]*component, 0, len(groups))
	for _, g := range groups {
		sort.Slice(g.vars, func(i, j int) bool { return g.vars[i] < g.vars[j] })
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].vars[0] < out[j].vars[0] })
	c.stats.Components += uint64(len(out))
	return out
}

func (c *weightedCounter) find(v Var) Var {
	for c.uf[v] != v {
		c.uf[v] = c.uf[c.uf[v]]
		v = c.uf[v]
	}
	return v
}

func (c *weightedCounter) union(a, b Var) {
	if ra, rb := c.find(a), c.find(b); ra != rb {
		c.uf[ra] = rb
	}
}

func (c *weightedCounter) branchVar(comp *component) Var {
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
	if c.order != nil {
		best, bestOrder, bestCount := comp.vars[0], int32(-1), int32(-1)
		for _, v := range comp.vars {
			if c.order[v] > bestOrder || (c.order[v] == bestOrder && c.occ[v] > bestCount) {
				best, bestOrder, bestCount = v, c.order[v], c.occ[v]
			}
		}
		return best
	}
	best, bestCount := comp.vars[0], int32(-1)
	for _, v := range comp.vars {
		if c.occ[v] > bestCount {
			best, bestCount = v, c.occ[v]
		}
	}
	return best
}

func (c *weightedCounter) key(comp *component) string {
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
