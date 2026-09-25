package gomisat

import (
	"fmt"
	"log"
	"math"
	"sort"
	"sync/atomic"
)

const (
	debug          = false
	debugOrderHeap = true
	debugAssert    = true
)

type SolverOptions struct {
	Verbosity                  int
	VarDecay                   float64
	ClauseDecay                float64
	RandomVarFreq              float64
	RandomSeed                 float64
	LubyRestart                bool
	CcminMode                  int     // Controls conflict clause minimization (0=none, 1=basic, 2=deep).
	PhaseSaving                int     // Controls the level of phase saving (0=none, 1=limited, 2=full).
	RndPol                     bool    // Use random polarities for branching heuristics.
	RndInitAct                 bool    // Initialize variable activities with a small random value.
	GarbageFrac                float64 // The fraction of wasted memory allowed before a garbage collection is triggered.
	MinLearntsLim              float64 // Minimum number to set the learnts limit to.
	RestartFirst               float64 // The initial restart limit. (default 100)
	RestartInc                 float64 // The factor with which the restart limit is multiplied in each restart. (default 1.5)
	LearntsizeFactor           float64 // The intitial limit for learnt clauses is a factor of the original clauses. (default 1 / 3)
	LearntsizeInc              float64 // The limit for learnt clauses is multiplied with this factor each restart. (default 1.1)
	LearntsizeAdjustStartConfl float64
	LearntsizeAdjustInc        float64

	// Literal block distance and tiered management of learnt clauses (lbd.go).
	// With UseLBD false the solver reduces the clause database by activity only,
	// as MiniSat does, which is what the LBD work is measured against.
	UseLBD      bool
	LBDCore     int    // clauses at or below this LBD are never deleted
	LBDTier2    int    // clauses at or below this LBD are kept while in use
	Tier2MaxAge uint64 // conflicts a mid-tier clause may go unused before demotion

	// Reduction schedule. With ReduceByConflicts the database is reduced after a
	// number of conflicts that grows by ReduceInc after every reduction, as
	// Glucose does. Without it the trigger is MiniSat's: reduce when the database
	// outgrows a budget that itself keeps growing, which does not bound the
	// database. Kept switchable because which one wins is a measurement, not an
	// opinion.
	ReduceByConflicts bool
	ReduceFirst       uint64
	ReduceInc         uint64
	// ProtectTier2 keeps the mid tier out of the deletion candidates while it is
	// still in use. Glucose only protects the core tier and takes half of
	// everything else, which is what the measurement prefers here.
	ProtectTier2 bool
	// NoReduce keeps every learnt clause. Useless as a setting, but it is the
	// arm that tells whether deleting clauses at all is what costs on a given
	// family.
	NoReduce bool
}

func DefaultSolverOptions() *SolverOptions {
	return &SolverOptions{
		Verbosity:                  0,
		VarDecay:                   0.95,
		ClauseDecay:                0.999,
		RandomVarFreq:              0,
		RandomSeed:                 91648253,
		LubyRestart:                true,
		CcminMode:                  2,
		PhaseSaving:                2,
		RndPol:                     false,
		RndInitAct:                 false,
		GarbageFrac:                0.2,
		MinLearntsLim:              0,
		RestartFirst:               100,
		RestartInc:                 2,
		LearntsizeFactor:           1.0 / 3.0,
		LearntsizeInc:              1.1,
		LearntsizeAdjustStartConfl: 100,
		LearntsizeAdjustInc:        1.5,
		UseLBD:                     true,
		LBDCore:                    2,
		LBDTier2:                   6,
		Tier2MaxAge:                30000,
		ReduceByConflicts:          true,
		ProtectTier2:               false,
		ReduceFirst:                2000,
		ReduceInc:                  300,
	}
}

type VarData struct {
	reason CRef
	level  int
}

// Watcher is an entry of a watch list. It holds a clause reference rather than a
// pointer, and a blocking literal that lets propagation skip the clause without
// touching it at all.
type Watcher struct {
	cref    CRef
	blocker Lit
}

func (w Watcher) String() string {
	return "[clause " + fmt.Sprint(uint32(w.cref)) + ", blocker " + w.blocker.String() + "]"
}

type Solver struct {
	arena       *clauseArena // Owns every clause; see clause.go.
	clauses     []CRef       // List of problem clauses.
	learnts     []CRef       // List of learnt clauses.
	trail       []Lit        // Assignment stack; stores all assigments made in the order they were made.
	trailLim    []int        // Separator indices for different decision levels in 'trail'.
	assumptions []Lit        // Current set of assumptions provided to solve by the user.

	userPol   map[Var]bool // The users preferred polarity of each variable.
	activity  []float64    // A heuristic measurement of the activity of a variable.
	assigns   []LBool      // The current assignments.
	polarity  []bool       // The preferred polarity of each variable.
	decision  []bool       // Declares if a variable is eligible for selection in the decision heuristic.
	vardata   []VarData    // Stores reason and level for each variable.
	watches   [][]Watcher  // watched[lit] is a list of constraints watching 'lit' (will go there if literal becomes true).
	orderHeap *VarHeap     // A priority queue of variables ordered with respect to the variable activity.

	maxLearnts            float64
	learntsizeAdjustConfl float64
	learntsizeAdjustCnt   int

	decVars         uint64
	numClauses      uint64
	numLearntes     uint64
	clausesLiterals uint64
	learntsLiterals uint64
	maxLiterals     uint64
	totLiterals     uint64

	// solver state
	ok              bool  // If false, the constraints are already unsatisfiable. No part of the solver state may be used
	qhead           int   // Head of queue (as index into the trail; no more explicit propagation queue in MiniSat)
	simpDBProps     int64 // Remaining number of propatations that must be made before next execution of 'simplify()'
	simpDBAssigns   int   // Number of top-level assignments since last execution of simplify()
	removeSatisfied bool

	conflict map[Lit]struct{}
	model    map[Var]LBool

	claInc float64 // Amount to bump next clause with
	varInc float64 // Amount to bump next variable with

	// Generation-stamped scratch space for computing the literal block distance
	// without allocating; see lbd.go.
	lbdStamp      []uint64
	lbdGeneration uint64

	// Conflict analysis scratch space, indexed by variable. Kept on the solver
	// because analyze used to allocate a map per conflict, which the profile put
	// at about a seventh of the whole run.
	//   0: not seen  1: source  2: removable  3: failed
	seen        []uint8
	seenToClear []Var

	// Reduction schedule (lbd.go): the conflict count at which the next
	// reduction is due, and the interval that produced it.
	reduceAt       uint64
	reduceInterval uint64

	nextVar      Var
	releasedVars []Var
	freeVars     []Var

	conflictBudget    int64
	propagationBudget int64
	asynchInterrupt   atomic.Bool // written by Interrupt, polled by withinBudget

	Solves       uint64
	Starts       uint64
	Decisions    uint64
	Propagations uint64
	Conflicts    uint64
	RndDecisions uint64
	Progress     float64
}

func NewSolver() *Solver {
	s := &Solver{
		arena:             newClauseArena(),
		activity:          make([]float64, 0),
		assigns:           make([]LBool, 0),
		polarity:          make([]bool, 0),
		userPol:           make(map[Var]bool),
		decision:          make([]bool, 0),
		vardata:           make([]VarData, 0),
		seen:              make([]uint8, 0),
		seenToClear:       make([]Var, 0, 64),
		watches:           make([][]Watcher, 0),
		releasedVars:      make([]Var, 0),
		freeVars:          make([]Var, 0),
		ok:                true,
		qhead:             0,
		simpDBAssigns:     -1,
		simpDBProps:       0,
		removeSatisfied:   true,
		nextVar:           0,
		conflictBudget:    -1,
		propagationBudget: -1,
		claInc:            1,
		varInc:            1,
	}
	s.orderHeap = NewVarHeap(func(x, y Var) bool {
		return s.activity[x] > s.activity[y]
	})
	return s
}

func (s *Solver) setDecisionVar(v Var, b bool) {
	if b && !s.decision[v] {
		s.decVars++
	} else if !b && s.decision[v] {
		s.decVars--
	}
	if debug && debugOrderHeap {
		log.Println("setDecisionVar:", s.orderHeap)
	}
}

// Add a new variable with parameters specifying variable mode.
//
//	upol: Assinged value for a variable. The default is LUndef
//	dvar: Indicator whether a variable is to be determined. The default is true.
func (s *Solver) NewVar(dvar bool, options *SolverOptions) Var {
	var v Var
	n := len(s.freeVars)
	if n > 0 {
		v = s.freeVars[n-1]
		s.freeVars = s.freeVars[:n-1]
		s.assigns[v] = LUndef
		s.vardata[v] = VarData{reason: CRefUndef, level: 0}
		s.seen[v] = 0
		if options.RndInitAct {
			s.activity[v] = drand(&options.RandomSeed) * 0.00001
		} else {
			s.activity[v] = 0
		}
		s.polarity[v] = true
		// s.userPol[v] = upol
		s.decision[v] = dvar
		s.watches[v+v] = make([]Watcher, 0)
		s.watches[v+v+1] = make([]Watcher, 0)
		s.orderHeap.indicies[v] = UndefIndex
		if dvar {
			s.orderHeap.Insert(v)
		}
	} else {
		v = s.nextVar
		s.nextVar++
		s.assigns = append(s.assigns, LUndef)
		s.vardata = append(s.vardata, VarData{reason: CRefUndef, level: 0})
		s.seen = append(s.seen, 0)
		if options.RndInitAct {
			s.activity = append(s.activity, drand(&options.RandomSeed)*0.00001)
		} else {
			s.activity = append(s.activity, 0)
		}
		s.polarity = append(s.polarity, true)
		// s.userPol = append(s.userPol, upol)
		s.decision = append(s.decision, dvar)
		s.watches = append(s.watches, make([]Watcher, 0))
		s.watches = append(s.watches, make([]Watcher, 0))
		s.orderHeap.indicies = append(s.orderHeap.indicies, UndefIndex)
		if dvar {
			s.orderHeap.Insert(v)
		}
	}

	if dvar && !s.decision[v] {
		s.decVars++
	} else if !dvar && s.decision[v] {
		s.decVars--
	}
	return v
}

func (s *Solver) AddClause(ps ...Lit) bool {
	if s.ok == false {
		return false
	}
	sort.Slice(ps, func(i, j int) bool {
		return ps[i] < ps[j]
	})
	j := 0
	p := LitUndef
	for i := 0; i < len(ps); i++ {
		if s.LitValue(ps[i]) == LTrue || ps[i] == p.Not() {
			return true
		} else if s.LitValue(ps[i]) != LFalse && ps[i] != p {
			p, ps[j] = ps[i], ps[i]
			j++
		}
	}
	ps = ps[:j] // shrink
	if len(ps) == 0 {
		s.ok = false
		if debug {
			log.Println("AddClause: ps becomes empty")
		}
		return false
	} else if len(ps) == 1 {
		s.UncheckedEnqueue(ps[0], CRefUndef)
		if confl := s.Propagate(); hasConflict(confl) == false {
			if debug {
				log.Println("AddClause: ps becomes a single literal", ps, "conflict of propagation", confl)
			}
			s.ok = true
			return true
		} else {
			if debug {
				log.Println("AddClause: ps becomes a single literal", ps, "conflict of propagation", confl)
			}
			s.ok = false
			return false
		}
	} else {
		// set clause
		c := s.arena.alloc(ps, false, true)
		s.clauses = append(s.clauses, c)
		s.AttachClause(c)
		if debug {
			log.Println("AddClause: ps becomes a clause (two or more literals)", c)
		}
		return true
	}
}

// LitValue is the value of a literal under the current assignment. The sign is
// folded in arithmetically, as in MiniSat, which is why LUndef has two
// representations (2 and 3); no caller compares against LUndef directly, they
// test against LTrue and LFalse.
func (s *Solver) LitValue(p Lit) LBool {
	return s.assigns[p.Var()] ^ LBool(p&1)
}

func (s *Solver) decisionLevel() int {
	return len(s.trailLim)
}

func luby(y float64, x int) float64 {
	// Find the finite subsequence that contains index 'x', and the
	// size of that subsequence:

	var size, seq int
	seq = 0
	for size = 1; size < x+1; size = 2*size + 1 {
		seq++
	}

	for size-1 != x {
		size = (size - 1) >> 1
		seq--
		x = x % size
	}

	return math.Pow(y, float64(seq))
}

func (s *Solver) UncheckedEnqueue(p Lit, c CRef) {
	s.assigns[p.Var()] = NewLBool(!p.Sign())
	s.vardata[p.Var()] = VarData{reason: c, level: s.decisionLevel()}
	s.trail = append(s.trail, p)
	if debug {
		log.Println("UncheckedEnqueue: Variable", p.Var(), "is assinged as", s.assigns[p.Var()])
	}
}

// RemoveSatisfied deletes the clauses satisfied at the root level and strips the
// literals that are already false from the ones that remain. It returns the list
// of surviving clauses.
func (s *Solver) RemoveSatisfied(cs []CRef) []CRef {
	j := 0
	for _, c := range cs {
		if s.Satisfied(c) {
			s.RemoveClause(c)
			continue
		}
		// Trim the clause. Only literals beyond the two watched ones may go.
		lits := s.arena.Lits(c)
		for k := 2; k < len(lits); k++ {
			if s.LitValue(lits[k]) == LFalse {
				if debug {
					log.Println("RemoveSatisfied: Remove a literal that becomes false", lits[k])
				}
				lits[k] = lits[len(lits)-1]
				lits = lits[:len(lits)-1]
				s.arena.shrink(c, 1)
				// The literal moved into position k has not been looked at yet.
				k--
			}
		}
		cs[j] = c
		j++
	}
	// Keep the survivors. The original code truncated to the number of deleted
	// clauses instead, which emptied the list whenever nothing was satisfied.
	return cs[:j]
}

// RemoveClause deletes a clause. The watch lists are not searched for it: the
// clause is flagged, propagation drops the entries it meets, and collectGarbage
// sweeps whatever is left. Detaching strictly, which is what this used to do,
// costs a linear scan of two watch lists per deleted clause, and a reduction
// deletes thousands at a time.
func (s *Solver) RemoveClause(c CRef) {
	if debug {
		log.Println("RemoveCluase: Remove the clause", s.arena.String(c))
	}
	if s.Locked(c) {
		v := s.arena.Lits(c)[0].Var()
		s.vardata[v] = VarData{reason: CRefUndef, level: s.vardata[v].level}
	}
	if s.arena.markDead(c) == false {
		return
	}
	size := uint64(s.arena.Size(c))
	if s.arena.Learnt(c) {
		s.numLearntes--
		s.learntsLiterals -= size
	} else {
		s.numClauses--
		s.clausesLiterals -= size
	}
}

// sweepWatches drops every watcher of a deleted clause.
func (s *Solver) sweepWatches() {
	for p, ws := range s.watches {
		j := 0
		for i := 0; i < len(ws); i++ {
			if s.arena.Dead(ws[i].cref) {
				continue
			}
			ws[j] = ws[i]
			j++
		}
		s.watches[p] = ws[:j]
	}
}

func (s *Solver) AttachClause(c CRef) {
	lits := s.arena.Lits(c)
	s.watches[lits[0].Not()] = append(s.watches[lits[0].Not()], Watcher{c, lits[1]})
	s.watches[lits[1].Not()] = append(s.watches[lits[1].Not()], Watcher{c, lits[0]})
	if s.arena.Learnt(c) {
		s.numLearntes++
		s.learntsLiterals += uint64(len(lits))
	} else {
		s.numClauses++
		s.clausesLiterals += uint64(len(lits))
	}
}

// Return true if a clause is a reason for some implication in the currrent state
func (s *Solver) Locked(c CRef) bool {
	first := s.arena.Lits(c)[0]
	return s.LitValue(first) == LTrue && s.vardata[first.Var()].reason == c
}

// Return true if a clause is satisfied in the current state
func (s *Solver) Satisfied(c CRef) bool {
	for _, lit := range s.arena.Lits(c) {
		if s.LitValue(lit) == LTrue {
			return true
		}
	}
	return false
}

// Perform unit propagation. Return possibly conflicting clause.
// Propagate all enqueued facts. If a conflict arises, the conflicting clause is returned,
// otherwise nil (CRef_Undef)
//
// Post condition
//
//	the propagation queue is empty, even if there was a conflict.
func (s *Solver) Propagate() CRef {
	confl := CRefUndef
	numProps := 0

	for s.qhead < len(s.trail) {
		p := s.trail[s.qhead]
		s.qhead++
		ws := s.watches[p]
		numProps++
		if debug {
			log.Println("Propagate: Check an assigned literal", p, " ", ws)
		}

		i := 0
		j := 0
		for i < len(ws) {
			if debug {
				log.Println("Propagate: Check a watcher", ws[i])
			}
			// A deleted clause is dropped from the list here rather than by
			// searching for it when it was deleted.
			if s.arena.Dead(ws[i].cref) {
				i++
				continue
			}

			// Try to avoid inspecting the clause
			blocker := ws[i].blocker
			if s.LitValue(blocker) == LTrue {
				ws[j] = ws[i]
				i++
				j++
				continue
			}

			// Make sure the false literal is lits[1]
			c := ws[i].cref
			lits := s.arena.Lits(c)
			falseLit := p.Not()
			if lits[0] == falseLit {
				lits[0], lits[1] = lits[1], falseLit
			}
			i++

			// If 0th watch is true, then clause is already satisfied.
			first := lits[0]
			w := Watcher{c, first}
			if first != blocker && s.LitValue(first) == LTrue {
				ws[j] = w
				if debug {
					log.Println("Propagate: Attach a new watcher", w, "to a literal", p)
				}
				j++
				continue
			}

			// Look for new watch
			for k := 2; k < len(lits); k++ {
				if s.LitValue(lits[k]) != LFalse {
					lits[1], lits[k] = lits[k], falseLit
					s.watches[lits[1].Not()] = append(s.watches[lits[1].Not()], w)
					goto nextClause
				}
			}

			// Did not find watch -- clause is unit under assignment
			ws[j] = w
			j++
			if s.LitValue(first) == LFalse {
				confl = c
				s.qhead = len(s.trail)
				// copy the remaining watches
				for i < len(ws) {
					ws[j] = ws[i]
					j++
					i++
				}
			} else {
				s.UncheckedEnqueue(first, c)
			}

		nextClause:
		}
		if debug {
			log.Printf("Propagate: shrink watchers; lit %s i-j %d\n", p.String(), len(ws)-j)
		}
		s.watches[p] = ws[:j]
	}
	s.Propagations += uint64(numProps)
	s.simpDBProps -= int64(numProps)
	return confl
}

// hasConflict reports whether a propagation result is a conflict.
func hasConflict(c CRef) bool { return c != CRefUndef }

//
// simplify
// Simplify the clause database according to the current top-level assignment.
// Currently, the only thing done here is the removal of satisfied clauses, but
// more things can be put here.

func (s *Solver) Simplify() bool {
	if s.ok == false || hasConflict(s.Propagate()) {
		s.ok = false
		return false
	}

	if len(s.trail) == s.simpDBAssigns || s.simpDBProps > 0 {
		//		log.Println("Simplify: The result is true because len(s.trail) == s.simpDBAssigns || s.simpDBProps > 0")
		return true
	}

	seen := make(map[Var]struct{})

	// Remove satisfied clauses
	s.learnts = s.RemoveSatisfied(s.learnts)
	if s.removeSatisfied { // s.removeSatisfied is always true for the time being
		s.clauses = s.RemoveSatisfied(s.clauses)

		if debug {
			log.Println("Simplify: The released variables: ", s.releasedVars)
		}
		// Remove all released variables from the trail
		for _, v := range s.releasedVars {
			seen[v] = struct{}{}
		}

		// Keep the literals whose variable was *not* released. The original
		// condition was inverted, which emptied the root-level trail.
		j := 0
		for _, lit := range s.trail {
			if _, released := seen[lit.Var()]; released == false {
				s.trail[j] = lit
				j++
			}
		}
		s.trail = s.trail[:j]
		s.qhead = len(s.trail)
		s.freeVars = append(s.freeVars, s.releasedVars...)
		s.releasedVars = s.releasedVars[:0]
	}
	s.collectGarbage()
	s.rebuildOrderHeap()

	s.simpDBAssigns = len(s.trail)
	s.simpDBProps = int64(s.clausesLiterals) + int64(s.learntsLiterals) // shouldn't depend on stats really, but it will do for now

	return true
}

func (s *Solver) rebuildOrderHeap() {
	vs := make([]Var, 0, s.nextVar)
	for v := Var(0); v < s.nextVar; v++ {
		if s.decision[v] && (s.assigns[v] != LTrue && s.assigns[v] != LFalse) {
			vs = append(vs, Var(v))
		}
	}
	s.orderHeap.Build(vs)
	if debug && debugOrderHeap {
		log.Println("rebuildOrderHeap: ", s.orderHeap.heap)
	}
}

// Solve searches for a model of the current clause set without assumptions.
// Assumptions left over from a previous call are discarded.
func (s *Solver) Solve(options *SolverOptions) LBool {
	return s.SolveWithAssumptions(nil, options)
}

// SolveWithAssumptions searches for a model in which every literal of
// assumptions is true. The assumptions hold for this call only; they are
// replaced on every call, so a solver can be reused for a sequence of queries.
//
// On LTrue the assignment is available through Model / ModelValue.
// On LFalse, if assumptions were given, UnsatCore reports the subset of them
// that is already sufficient for unsatisfiability. An empty core means the
// clause set itself is unsatisfiable and the solver is permanently UNSAT.
func (s *Solver) SolveWithAssumptions(assumptions []Lit, options *SolverOptions) LBool {
	s.assumptions = append(s.assumptions[:0], assumptions...)
	// An assumption may name a variable that no clause mentions, which is a
	// legitimate query: the variable simply is free. Create it rather than
	// indexing past the end of the assignment arrays.
	for _, p := range s.assumptions {
		s.addVar(int64(p.Var()), options)
	}
	return s.solve(options)
}

func (s *Solver) solve(options *SolverOptions) LBool {
	s.model = make(map[Var]LBool)
	s.conflict = make(map[Lit]struct{})

	if s.ok == false {
		return LFalse
	}

	s.Solves++

	s.maxLearnts = float64(s.numClauses) * options.LearntsizeFactor
	if s.maxLearnts < options.MinLearntsLim {
		s.maxLearnts = options.MinLearntsLim
	}
	if debug {
		log.Println("Solve: maxLearnts", s.numClauses, options.LearntsizeFactor, s.maxLearnts)
	}

	s.learntsizeAdjustConfl = options.LearntsizeAdjustStartConfl
	s.learntsizeAdjustCnt = int(s.learntsizeAdjustConfl)
	status := LUndef

	//	log.Println("==== Search Statistics ====")

	// Search
	currRestarts := 0
	for status != LTrue && status != LFalse { // this means status == LUndef
		var resetBase float64
		if options.LubyRestart {
			resetBase = luby(options.RestartInc, currRestarts)
		} else {
			resetBase = math.Pow(options.RestartInc, float64(currRestarts))
		}

		status = s.search(int(resetBase*options.RestartFirst), options)
		if s.withinBudget() == false {
			break
		}
		currRestarts++
	}

	if status == LTrue {
		// Extend & copy model
		for k, v := range s.assigns {
			s.model[Var(k)] = v
		}
	} else if status == LFalse && len(s.conflict) == 0 {
		s.ok = false
	}
	// Return to the root level so that the solver can be reused by the next
	// call. Without this the trail is left inside a decision level and any
	// subsequent solve starts from a corrupted state.
	s.cancelUntil(0, options)
	return status
}

// search
// Search for a model the specified number of conflicts.
// Note: Use negative value for nof_conflicts indicate infinity
//
// Output
//
//	LTrue if a partial assigment that is consistent with respect to the clauseset if found.
//	If all variables are decision variables, this means that the clause set is satisfiable.
//	LFalse if the clause set is insatisfiable. LUndef if the bound on number of conflicts is reached.
func (s *Solver) search(nofConflicts int, options *SolverOptions) LBool {
	// backtranckLevel := 0
	conflictC := 0
	s.Starts++

	// for k := 0; k < 5; k++ { // for test
	for {
		if confl := s.Propagate(); hasConflict(confl) {
			if debug {
				log.Println("search: Find a conflict", confl)
			}
			s.Conflicts++
			conflictC++
			if s.decisionLevel() == 0 {
				return LFalse
			}
			learntClause, backtranckLevel, lbd := s.analyze(confl, options)
			s.cancelUntil(backtranckLevel, options)
			if debug {
				log.Println("Propagete: The result of analyze", learntClause, backtranckLevel)
			}

			if len(learntClause) == 1 {
				s.UncheckedEnqueue(learntClause[0], CRefUndef)
			} else {
				c := s.arena.alloc(learntClause, true, false)
				s.noteLearnt(c, lbd, options)
				s.learnts = append(s.learnts, c)
				s.AttachClause(c)
				s.claBumpActivity(c)
				s.UncheckedEnqueue(learntClause[0], c)
			}

			s.varDecayActivity(options)
			s.claDecayActivity(options)

			s.learntsizeAdjustCnt--
			if s.learntsizeAdjustCnt == 0 {
				s.learntsizeAdjustConfl *= options.LearntsizeAdjustInc
				s.learntsizeAdjustCnt = int(s.learntsizeAdjustConfl)
				s.maxLearnts *= options.LearntsizeInc

				//				log.Println("||")
			}
		} else {
			if debug {
				log.Println("search: No conflict")
			}

			if (nofConflicts >= 0 && conflictC >= nofConflicts) || !s.withinBudget() {
				if debug {
					log.Println("search: Reached bound on number of conflicts")
				}
				s.Progress = s.progressEstimate()
				s.cancelUntil(0, options)
				return LUndef
			}

			//simplify the set of problem clauses
			if s.decisionLevel() == 0 && s.Simplify() == false {
				if debug {
					log.Println("search: Simplified problem has a conflict (UNSAT)")
				}
				return LFalse
			}

			if s.reductionDue(options) {
				if debug {
					log.Println("search: Reduce the set of learnt clauses", len(s.learnts), len(s.trail), s.maxLearnts)
				}
				s.reduceDB(options)
			}

			next := LitUndef
			// NOTE: this loop must be labelled. In Go a bare 'break' inside a
			// switch leaves the switch only, not the enclosing for, so the
			// unassigned-assumption case would spin forever re-reading the same
			// literal. The C++ original relies on 'break' leaving the while.
		placeAssumptions:
			for s.decisionLevel() < len(s.assumptions) {
				if debug {
					log.Println("search: Perform user provided assumption")
				}
				p := s.assumptions[s.decisionLevel()]
				switch s.LitValue(p) {
				case LTrue:
					// Dummy decision level
					s.newDecisionLevel()
				case LFalse:
					// An assumption is falsified: record the subset of the
					// assumptions responsible for it, i.e. the UNSAT core.
					s.conflict = s.analyzeFinal(p.Not())
					return LFalse
				default:
					next = p
					break placeAssumptions
				}
			}

			if next == LitUndef {
				if debug {
					log.Println("search: New variable decision")
				}
				s.Decisions++
				next = s.pickBranchLit(options)
				if next == LitUndef {
					if debug {
						log.Println("search: Model found", s.assigns)
					}
					return LTrue
				}
			}

			if debug {
				log.Println("search: Increase decision level and enqueue next", next)
			}
			s.newDecisionLevel()
			s.UncheckedEnqueue(next, CRefUndef)
		}
	}
}

func (s *Solver) pickBranchLit(options *SolverOptions) Lit {
	next := VarUndef

	// Random decision
	if drand(&options.RandomSeed) < options.RandomVarFreq && s.orderHeap.IsEmpty() == false {
		next = s.orderHeap.heap[irand(&options.RandomSeed, len(s.orderHeap.heap))]
		// The condition has to be 'unassigned and eligible'; the disjunction it
		// replaces was true for every value of next.
		if s.assigns[next] == LUndef && s.decision[next] == true {
			s.RndDecisions++
		}
	}
	if debug {
		log.Println("pickBranchLit: Random choose", next)
	}

	// Activity based decision
	if debug && debugOrderHeap {
		log.Println("pickBranchLit: orderheap at starting activity based decision", s.orderHeap.heap)
	}
	for next == VarUndef || s.assigns[next] == LTrue || s.assigns[next] == LFalse || s.decision[next] == false {
		if s.orderHeap.IsEmpty() {
			next = VarUndef
			break
		} else {
			next = s.orderHeap.RemoveMin()
			if debug && debugOrderHeap {
				log.Println("pickBranchLit: orderheap after removeMin", s.orderHeap)
			}
		}
	}
	if debug {
		if next == VarUndef {
			log.Println("pickBranchLit: Active based choose", next, s.activity[next])
		} else {
			log.Println("pickBranchLit: Active based choose is VarUndef")
		}
	}
	if debug && debugOrderHeap {
		log.Println("pickBranchLit: orderheap after selection", s.orderHeap.heap)
	}

	// Choose polarity based on different polarity modes (global or per-variable)
	if next == VarUndef {
		return LitUndef
	} else if upol, ok := s.userPol[next]; ok {
		return MkLit(next, upol)
	} else if options.RndPol {
		return MkLit(next, drand(&options.RandomSeed) < 0.5)
	} else {
		return MkLit(next, s.polarity[next])
	}
}

func (s *Solver) newDecisionLevel() {
	s.trailLim = append(s.trailLim, len(s.trail))
	if debug {
		log.Println("newDecisionLevel: decision level", s.decisionLevel())
	}
}

func (s *Solver) progressEstimate() float64 {
	progress := 0.0
	F := 1.0 / float64(s.nextVar)
	for i := 0; i < s.decisionLevel(); i++ {
		var beg, end int
		if i == 0 {
			beg = 0
		} else {
			beg = s.trailLim[i-1]
		}
		if i == s.decisionLevel() {
			end = len(s.trail)
		} else {
			end = s.trailLim[i]
		}
		progress += math.Pow(F, float64(i)) * float64(end-beg)
	}
	return progress / float64(s.nextVar)
}

// reduceDB
// Remove half of the learnt clauses, minus the clauses locked by the current assignment. Locked
// clauses are clauses that are reason to some assignment. Binary clauses are never removed.
// reduceDB shrinks the learnt clause database and reports how many clauses it
// deleted.
func (s *Solver) reduceDB(options *SolverOptions) int {
	if options.UseLBD {
		return s.reduceDBTiered(options)
	}
	if len(s.learnts) == 0 {
		return 0
	}
	extraLim := float32(s.claInc / float64(len(s.learnts)))
	sort.Slice(s.learnts, func(i, j int) bool {
		mi, mj := &s.arena.meta[s.learnts[i]], &s.arena.meta[s.learnts[j]]
		return mi.size > 2 && (mj.size == 2 || mi.activity < mj.activity)
	})
	// Do not delete binary or locked clauses. From the rest, delete clauses from the first half
	// and clauses with activity smaller than extraLim
	j := 0
	removed := 0
	for i := 0; i < len(s.learnts); i++ {
		c := s.learnts[i]
		m := &s.arena.meta[c]
		if m.size > 2 && !s.Locked(c) && (i < len(s.learnts)/2 || m.activity < extraLim) {
			s.RemoveClause(c)
			removed++
		} else {
			s.learnts[j] = s.learnts[i]
			j++
		}
	}
	s.learnts = s.learnts[:j]
	if debug {
		log.Println("reduceDB: The number of new learnts", j)
	}
	s.collectGarbage()
	return removed
}

func (s *Solver) withinBudget() bool {
	return !s.asynchInterrupt.Load() && (s.conflictBudget < 0 || s.Conflicts < uint64(s.conflictBudget)) && (s.propagationBudget < 0 || s.Propagations < uint64(s.propagationBudget))
}

// Increase a clause with the current bump value
func (s *Solver) claBumpActivity(c CRef) {
	m := &s.arena.meta[c]
	m.activity += float32(s.claInc)
	if m.activity > 1e20 {
		// rescale
		for _, c := range s.learnts {
			s.arena.meta[c].activity *= 1e-20
		}
		s.claInc *= 1e-20
	}
}

func (s *Solver) varBumpActivity(v Var) {
	s.activity[v] += s.varInc
	if s.activity[v] > 1e100 {
		// rescale
		for k, _ := range s.activity {
			s.activity[k] *= 1e-100
		}
		s.varInc *= 1e-100
	}
	// update orderheap with respect to new activity
	if s.orderHeap.InHeap(v) {
		s.orderHeap.Decrease(v)
		if debug && debugOrderHeap {
			log.Println("varBumpActivity: orderheap", s.orderHeap)
		}
	}
}

func (s *Solver) varDecayActivity(options *SolverOptions) {
	s.varInc *= 1.0 / options.VarDecay
}

func (s *Solver) claDecayActivity(options *SolverOptions) {
	s.claInc *= 1.0 / options.ClauseDecay
}

// Revert to the state at given level (keeping all assignment at level but not beyond)
func (s *Solver) cancelUntil(level int, options *SolverOptions) {
	if s.decisionLevel() > level {
		for i := len(s.trail) - 1; i >= s.trailLim[level]; i-- {
			x := s.trail[i].Var()
			s.assigns[x] = LUndef
			if options.PhaseSaving > 1 || (options.PhaseSaving == 1 && i > s.trailLim[len(s.trailLim)-1]) {
				s.polarity[x] = s.trail[i].Sign()
			}
			s.orderHeap.Insert(x)
			if debug && debugOrderHeap {
				log.Println("cancelUntil: orderheap", s.orderHeap)
			}
		}
		s.qhead = s.trailLim[level]
		s.trail = s.trail[:s.trailLim[level]]
		s.trailLim = s.trailLim[:level]
	}
}

// analyze
// Analyze conflict and produce a reason clause
//
// Pre-conditions:
//   - outLeant is assumed to be cleared
//   - current decision level must be greather than root level
//
// Post-conditions:
//   - outLeant[0] is the asserting literal at level 'outBtlevel'
//   - If outLearnt.size() > 1 then outLearnt[1] has the greatest decision level of the
//     rest of literals. There may be others from the same level through.

func (s *Solver) analyze(c CRef, options *SolverOptions) ([]Lit, int, int) {
	pathC := 0
	p := LitUndef
	outLearnt := make([]Lit, 1, s.arena.Size(c)) // outLeant[0] will be put at the end of this function

	// Generate conflict clause
	index := len(s.trail) - 1
	for {
		if s.arena.Learnt(c) {
			s.claBumpActivity(c)
			s.noteUsed(c, options)
		}

		lits := s.arena.Lits(c)
		var j int
		if p == LitUndef {
			j = 0
		} else {
			j = 1
		}
		for ; j < len(lits); j++ {
			v := lits[j].Var()
			if s.seen[v] == 0 {
				if s.vardata[v].level > 0 {
					s.varBumpActivity(v)
					s.markSeen(v, 1)
					if s.vardata[v].level >= s.decisionLevel() {
						pathC++
					} else {
						outLearnt = append(outLearnt, lits[j])
					}
				}
			}
		}

		for {
			p = s.trail[index]
			if s.seen[p.Var()] != 0 {
				c = s.vardata[p.Var()].reason
				s.seen[p.Var()] = 0
				break
			} else {
				index--
			}
		}
		pathC--

		if debug {
			log.Println("analyze: pathC in the loop", pathC)
		}
		if pathC == 0 {
			break
		}
	}
	outLearnt[0] = p.Not()
	if debug {
		log.Println("analyze: outLearnt", outLearnt)
	}

	// simplify conflict clause
	switch {
	case options.CcminMode == 2:
		j := 1
		for i := 1; i < len(outLearnt); i++ {
			p := outLearnt[i]
			if s.vardata[p.Var()].reason == CRefUndef || s.litRedundant(p) == false {
				outLearnt[j] = p
				j++
			}
		}
		s.maxLiterals += uint64(len(outLearnt))
		outLearnt = outLearnt[:j]
		s.totLiterals += uint64(len(outLearnt))
	case options.CcminMode == 1:
		j := 1
		for i := 1; i < len(outLearnt); i++ {
			p := outLearnt[i]
			if c := s.vardata[p.Var()].reason; c == CRefUndef {
				outLearnt[j] = p
				j++
			} else {
				lits := s.arena.Lits(c)
				for k := 1; k < len(lits); k++ {
					v := lits[k].Var()
					if s.seen[v] == 0 && s.vardata[v].level > 0 {
						outLearnt[j] = p
						j++
						break
					}
				}
			}
		}
		s.maxLiterals += uint64(len(outLearnt))
		outLearnt = outLearnt[:j]
		s.totLiterals += uint64(len(outLearnt))
	default:
		s.maxLiterals += uint64(len(outLearnt))
		s.totLiterals += uint64(len(outLearnt))
	}

	// Find correct backtrack level
	var outBtlevel int
	if len(outLearnt) == 1 {
		outBtlevel = 0
	} else {
		maxi := 1
		maxlevel := s.vardata[outLearnt[maxi].Var()].level
		for i := 2; i < len(outLearnt); i++ {
			if l := s.vardata[outLearnt[i].Var()].level; l > maxlevel {
				maxi = i
				maxlevel = l
			}
		}
		outLearnt[1], outLearnt[maxi] = outLearnt[maxi], outLearnt[1]
		outBtlevel = maxlevel
	}

	// The LBD has to be measured here, before the caller backjumps: after that
	// the decision levels no longer describe the trail this clause came from.
	lbd := s.computeLBD(outLearnt)
	s.clearSeen()
	return outLearnt, outBtlevel, lbd
}

// markSeen records a value in the analysis scratch space and remembers the
// variable so that clearSeen can reset only what was touched.
func (s *Solver) markSeen(v Var, value uint8) {
	if s.seen[v] == 0 {
		s.seenToClear = append(s.seenToClear, v)
	}
	s.seen[v] = value
}

func (s *Solver) clearSeen() {
	for _, v := range s.seenToClear {
		s.seen[v] = 0
	}
	s.seenToClear = s.seenToClear[:0]
}

// This is used in litRedundant
type redundantStackElem struct {
	i int
	l Lit
}

// Check if p can be removed from a conflict clause
func (s *Solver) litRedundant(p Lit) bool {
	// seen
	//   0: undef (key does not exist)
	//   1: seen_source
	//   2: seen_removable
	//   3: seen_failed
	//

	if debug && debugAssert {
		log.Println("litRedundant assertion (seen[var(p)] == seen_undef || seen[var(p)] == seen_source):", s.seen[p.Var()] == 0 || s.seen[p.Var()] == 1)
		log.Println("litRedundant assertion (reason(var(p)) != CRefUndef):", s.vardata[p.Var()].reason != CRefUndef)
	}

	stack := make([]redundantStackElem, 0, 10)
	c := s.vardata[p.Var()].reason
	i := 1
	for {
		lits := s.arena.Lits(c)
		if i < len(lits) {
			l := lits[i]

			// Variable at level 0 or previsouly removable
			if s.vardata[l.Var()].level == 0 || s.seen[l.Var()] == 1 || s.seen[l.Var()] == 2 {
				goto nextLoop
			}

			// Check variable cannot be removed for some local reason
			if s.vardata[l.Var()].reason == CRefUndef || s.seen[l.Var()] == 3 {
				stack = append(stack, redundantStackElem{0, p})
				for _, ss := range stack {
					if s.seen[ss.l.Var()] == 0 {
						s.markSeen(ss.l.Var(), 3)
					}
				}
				return false
			}

			// Recursively check l.
			// NOTE: the reason has to be the one of l. Go evaluates the whole
			// right-hand side before assigning, so writing this as a single
			// parallel assignment would take the reason of the old p, unlike the
			// sequential assignments of the C++ original.
			stack = append(stack, redundantStackElem{i, p})
			i, p = 0, l
			c = s.vardata[l.Var()].reason
		} else {
			// Finished with current element p and reason c
			if s.seen[p.Var()] == 0 {
				s.markSeen(p.Var(), 2)
			}

			// Terminate with success if stack is empty
			if len(stack) == 0 {
				return true
			}

			// Continue with top element on stack
			i, p = stack[len(stack)-1].i, stack[len(stack)-1].l
			c = s.vardata[p.Var()].reason
			stack = stack[:len(stack)-1]
		}
	nextLoop:
		i++
	}
}
