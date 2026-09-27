package gomisat

import "math/big"

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
	cache   map[cacheKey]*big.Float
	exact   map[string]*big.Float // used instead when CountOptions.ExactCache is set
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

	// Component storage. The searches at one node and below it are the only users
	// of a component, so the arenas follow the recursion: mark on entry, release
	// on the way out. See arena_scratch.go.
	varArena  *chunkArena[Var]
	refArena  *chunkArena[CRef]
	compArena *chunkArena[component]
	groupOf   []int32 // variable -> index of its group during a split
	rootGroup []int32 // union-find root -> index of its group
	groupSize []int32 // scratch, indexed by group: clause and variable counts
	groupVars []int32

	// varSum[v] is the weight of v being either value, the factor a free variable
	// contributes. Precomputed because the weights do not change during a count.
	varSum []*big.Float
}

// WeightedCount returns the sum over the models of the product of their literal
// weights. With unit weights it is the model count.
func (s *Solver) WeightedCount(options *SolverOptions, copt *CountOptions, weights *Weights) (*big.Float, CountStats) {
	return s.weightedCount(options, copt, weights, nil)
}

func (s *Solver) weightedCount(options *SolverOptions, copt *CountOptions, weights *Weights,
	cache map[cacheKey]*big.Float) (*big.Float, CountStats) {
	c := newWeightedCounter(s, options, copt, weights, cache)
	return c.count1()
}

// count1 answers one query with this counter, which may be reused for the next.
func (c *weightedCounter) count1() (*big.Float, CountStats) {
	c.stats = CountStats{}
	c.varArena.release(arenaMark{})
	c.refArena.release(arenaMark{})
	c.compArena.release(arenaMark{})
	total := c.run()
	c.stats.CacheSize = c.cacheSize()
	return total, c.stats
}

func (c *weightedCounter) cacheSize() int {
	if c.copt.ExactCache {
		return len(c.exact)
	}
	return len(c.cache)
}

// lookup and store hide which of the two cache representations is in use.
func (c *weightedCounter) lookup(comp *component) (*big.Float, bool) {
	if c.copt.ExactCache {
		hit, ok := c.exact[string(c.keyBytes(comp))]
		return hit, ok
	}
	hit, ok := c.cache[hashComponent(comp.vars, comp.clauses)]
	return hit, ok
}

func (c *weightedCounter) store(comp *component, value *big.Float) {
	if c.copt.ExactCache {
		c.exact[c.key(comp)] = value
		return
	}
	c.cache[hashComponent(comp.vars, comp.clauses)] = value
}

// newWeightedCounter builds the scratch state a counter needs. It is separated from
// the counting so that a Study can keep one and pay for this once rather than once
// per query, which was where the time went after the search itself had been cut.
func newWeightedCounter(s *Solver, options *SolverOptions, copt *CountOptions, weights *Weights,
	cache map[cacheKey]*big.Float) *weightedCounter {
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
	if cache == nil {
		cache = make(map[cacheKey]*big.Float)
	}
	c := &weightedCounter{
		s:       s,
		options: options,
		copt:    copt,
		weights: weights,
		cache:   cache,
		uf:      make([]Var, n),
		seenVar: make([]uint64, n),
		occ:     make([]int32, n),
		covered: make([]uint64, n),
		exact:   make(map[string]*big.Float),
		keyBuf:  make([]byte, 0, 256),
		allVars: make([]Var, n),
		prec:    prec,
	}
	chunk := 4096
	if 4*n > chunk {
		chunk = 4 * n
	}
	c.varArena = newChunkArena[Var](chunk)
	c.refArena = newChunkArena[CRef](chunk)
	c.compArena = newChunkArena[component](256)
	c.groupOf = make([]int32, n)
	c.rootGroup = make([]int32, n)
	for i := range c.groupOf {
		c.groupOf[i] = -1
		c.rootGroup[i] = -1
	}
	c.varSum = make([]*big.Float, n)
	for v := 0; v < n; v++ {
		c.varSum[v] = new(big.Float).SetPrec(prec).Add(
			weights.Of(MkLit(Var(v), false)), weights.Of(MkLit(Var(v), true)))
	}
	for v := 0; v < n; v++ {
		c.allVars[v] = Var(v)
	}
	if copt.Branching == BranchEliminationOrder {
		c.order = s.countingOrder()
	}

	return c
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

	root := c.residual(s.clauses, c.allVars)
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
	// The arenas are unwound in line, without a defer and without a closure: this
	// is the hot recursive function.
	varMark, refMark, compMark := c.varArena.mark(), c.refArena.mark(), c.compArena.mark()

	parts := c.split(comp)

	// The free variables and their factor have to be settled before recursing: the
	// mark array is shared with the nested calls, which raise the mark for their own
	// components.
	c.coverMark++
	mark := c.coverMark
	for i := range parts {
		for _, v := range parts[i].vars {
			c.covered[v] = mark
		}
	}
	total := c.one()
	for _, v := range scope {
		if c.s.assigns[v] == LUndef && c.covered[v] != mark {
			total.Mul(total, c.varSum[v])
		}
	}

	for i := range parts {
		n := c.count(&parts[i])
		if n.Sign() == 0 {
			c.varArena.release(varMark)
			c.refArena.release(refMark)
			c.compArena.release(compMark)
			return c.zero()
		}
		total.Mul(total, n)
	}
	c.varArena.release(varMark)
	c.refArena.release(refMark)
	c.compArena.release(compMark)
	return total
}

func (c *weightedCounter) count(comp *component) *big.Float {
	if len(comp.clauses) == 0 {
		total := c.one()
		for _, v := range comp.vars {
			total.Mul(total, c.varSum[v])
		}
		return total
	}

	if c.copt.UseCache {
		if hit, ok := c.lookup(comp); ok {
			c.stats.CacheHits++
			// Returned without copying. Callers only ever read the result into an
			// accumulator of their own, never write through it; countAll is the
			// only caller and it multiplies rather than assigns.
			return hit
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
		varMark, refMark, compMark := c.varArena.mark(), c.refArena.mark(), c.compArena.mark()
		sub := c.residual(comp.clauses, comp.vars)
		branch.Mul(branch, c.countAll(sub, comp.vars))
		c.varArena.release(varMark)
		c.refArena.release(refMark)
		c.compArena.release(compMark)
		total.Add(total, branch)
		c.s.cancelUntil(level, c.options)
	}

	if c.copt.UseCache {
		c.store(comp, new(big.Float).Set(total))
	}
	return total
}

// The structural helpers below are the same as the unweighted counter's. They are
// repeated rather than shared so that the two counters stay independent; see the
// note at the top of the file.

// residual keeps the clauses that are not satisfied yet and collects their
// unassigned variables, taking its storage from the arenas.
//
// It scans the clauses once, taking storage at an upper bound -- at most one entry
// per clause, at most one variable per variable of scope -- and leaves the unused
// tail behind, because the arena reuses it as soon as the mark is restored.
// Counting the exact sizes first would double the most-executed loop in the
// counter, which measured 20 % slower than paying for the waste.
//
// scope is the sorted variable list of the component these clauses came from. The
// variables that survive are collected by walking it rather than by sorting what
// the clauses yield: the cache key needs a canonical order, and taking it from an
// already-ordered list costs a linear pass instead of a sort. The sort was
// sort.Slice, which goes through reflection, and it was a third of the counter's
// running time.
func (c *weightedCounter) residual(clauses []CRef, scope []Var) *component {
	out := &c.compArena.alloc(1)[0]
	out.clauses = c.refArena.alloc(len(clauses))[:0]
	out.vars = c.varArena.alloc(len(scope))[:0]

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
			if c.assigned(p) {
				continue
			}
			c.seenVar[p.Var()] = c.mark
		}
	}
	for _, v := range scope {
		if c.seenVar[v] == c.mark {
			out.vars = append(out.vars, v)
		}
	}
	return out
}

func (c *weightedCounter) assigned(p Lit) bool {
	v := c.s.LitValue(p)
	return v == LTrue || v == LFalse
}

// split partitions a component into connected components: two clauses belong
// together when they share an unassigned variable.
//
// The grouping is done by counting rather than with a map, and the storage comes
// from the arenas, because this was three quarters of everything the counter
// allocated. Groups are numbered by the first appearance of one of their variables
// in comp.vars, which is sorted, so the result is ordered by smallest variable and
// each group's variables come out sorted without another pass.
func (c *weightedCounter) split(comp *component) []component {
	if c.copt.UseDecomposition == false || len(comp.clauses) <= 1 {
		if len(comp.clauses) == 0 {
			return nil
		}
		c.stats.Components++
		out := c.compArena.alloc(1)
		out[0] = *comp
		return out
	}

	for _, v := range comp.vars {
		c.uf[v] = v
	}
	for _, ref := range comp.clauses {
		first := VarUndef
		for _, p := range c.s.arena.Lits(ref) {
			if c.assigned(p) {
				continue
			}
			if first == VarUndef {
				first = p.Var()
				continue
			}
			c.union(first, p.Var())
		}
	}

	// One find per variable, and none per clause: the group of a variable is
	// resolved once and read from an array afterwards. Calling find again for every
	// clause and every counting pass was a third of the counter's time.
	ngroups := 0
	for _, v := range comp.vars {
		r := c.find(v)
		if c.rootGroup[r] < 0 {
			c.rootGroup[r] = int32(ngroups)
			ngroups++
		}
		c.groupOf[v] = c.rootGroup[r]
	}
	if cap(c.groupSize) < ngroups {
		c.groupSize = make([]int32, ngroups)
		c.groupVars = make([]int32, ngroups)
	}
	c.groupSize, c.groupVars = c.groupSize[:ngroups], c.groupVars[:ngroups]
	for g := range c.groupSize {
		c.groupSize[g], c.groupVars[g] = 0, 0
	}
	for _, ref := range comp.clauses {
		c.groupSize[c.groupOfClause(ref)]++
	}
	for _, v := range comp.vars {
		c.groupVars[c.groupOf[v]]++
	}

	parts := c.compArena.alloc(ngroups)
	for g := range parts {
		parts[g].clauses = c.refArena.alloc(int(c.groupSize[g]))[:0]
		parts[g].vars = c.varArena.alloc(int(c.groupVars[g]))[:0]
	}
	for _, ref := range comp.clauses {
		g := c.groupOfClause(ref)
		parts[g].clauses = append(parts[g].clauses, ref)
	}
	for _, v := range comp.vars {
		g := c.groupOf[v]
		parts[g].vars = append(parts[g].vars, v)
	}
	for _, v := range comp.vars {
		c.groupOf[v] = -1
		c.rootGroup[v] = -1
	}

	c.stats.Components += uint64(ngroups)
	return parts
}

// groupOfClause is the group of the first unassigned variable of a clause; every
// unassigned variable of a clause is in the same group by construction.
func (c *weightedCounter) groupOfClause(ref CRef) int32 {
	for _, p := range c.s.arena.Lits(ref) {
		if c.assigned(p) {
			continue
		}
		return c.groupOf[p.Var()]
	}
	return 0
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

// keyBytes builds the key in the reusable buffer. Looking up a map with
// string(buf) does not copy the bytes, so a hit costs no allocation; only an
// insert needs the string.
func (c *weightedCounter) keyBytes(comp *component) []byte {
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
	return buf
}

func (c *weightedCounter) key(comp *component) string {
	return string(c.keyBytes(comp))
}
