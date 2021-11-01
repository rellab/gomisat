package gomisat

import (
	"log"
	"math"
	"sort"
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
	RestartFirst               int     // The initial restart limit. (default 100)
	RestartInc                 float64 // The factor with which the restart limit is multiplied in each restart. (default 1.5)
	LearntsizeFactor           float64 // The intitial limit for learnt clauses is a factor of the original clauses. (default 1 / 3)
	LearntsizeInc              float64 // The limit for learnt clauses is multiplied with this factor each restart. (default 1.1)
	LearntsizeAdjustStartConfl float64
	LearntsizeAdjustInc        float64
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
		LearntsizeFactor:           1 / 3,
		LearntsizeInc:              1.1,
		LearntsizeAdjustStartConfl: 100,
		LearntsizeAdjustInc:        1.5,
	}
}

type VarData struct {
	reason *Clause
	level  int
}

// TODO: Make watcher manager to delete unused clauses by GC
type Watcher struct {
	clause  *Clause
	blocker Lit
}

func (w Watcher) String() string {
	return "[" + w.clause.String() + ", blocker " + w.blocker.String() + "]"
}

type Solver struct {
	clauses     []*Clause // List of problem clauses.
	learnts     []*Clause // List of learnt clauses.
	trail       []Lit     // Assignment stack; stores all assigments made in the order they were made.
	trailLim    []int     // Separator indices for different decision levels in 'trail'.
	assumptions []Lit     // Current set of assumptions provided to solve by the user.

	activity  map[Var]float64   // A heuristic measurement of the activity of a variable.
	assigns   map[Var]LBool     // The current assignments.
	polarity  map[Var]bool      // The preferred polarity of each variable.
	userPol   map[Var]bool      // The users preferred polarity of each variable.
	decision  map[Var]bool      // Declares if a variable is eligible for selection in the decision heuristic.
	vardata   map[Var]VarData   // Stores reason and level for each variable.
	watches   map[Lit][]Watcher // watched[lit] is a list of constraints watching 'lit' (will go there if literal becomes true).
	orderHeap *VarHeap          // A priority queue of variables ordered with respect to the variable activity.

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

	nextVar      Var
	releasedVars []Var
	freeVars     []Var

	conflictBudget    int64
	propagationBudget int64
	asynchInterrupt   bool

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
		activity:          make(map[Var]float64),
		assigns:           make(map[Var]LBool),
		polarity:          make(map[Var]bool),
		userPol:           make(map[Var]bool),
		decision:          make(map[Var]bool),
		vardata:           make(map[Var]VarData),
		watches:           make(map[Lit][]Watcher),
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
		asynchInterrupt:   false,
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
	s.decision[v] = b
	if b {
		s.orderHeap.Insert(v)
	}
	log.Println("setDecisionVar:", s.orderHeap)
}

// Add a new variable with parameters specifying variable mode.
//   upol: Assinged value for a variable. The default is LUndef
//   dvar: Indicator whether a variable is to be determined. The default is true.
func (s *Solver) NewVar(dvar bool, options *SolverOptions) Var {
	var v Var
	n := len(s.freeVars)
	if n > 0 {
		v = s.freeVars[n-1]
		s.freeVars = s.freeVars[:n-1]
	} else {
		v = s.nextVar
		s.nextVar++
	}

	s.assigns[v] = LUndef
	s.vardata[v] = VarData{reason: nil, level: 0}
	if options.RndInitAct {
		s.activity[v] = drand(&options.RandomSeed) * 0.00001
	} else {
		s.activity[v] = 0
	}
	s.polarity[v] = true
	// s.userPol[v] = upol
	s.setDecisionVar(v, dvar)
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
		log.Println("AddClause: ps becomes empty")
		return false
	} else if len(ps) == 1 {
		s.UncheckedEnqueue(ps[0], nil)
		if confl := s.Propagate(); confl == nil {
			log.Println("AddClause: ps becomes a single literal", ps, "conflict of propagation", confl)
			s.ok = true
			return true
		} else {
			log.Println("AddClause: ps becomes a single literal", ps, "conflict of propagation", confl)
			s.ok = false
			return false
		}
	} else {
		// set clause
		c := MkClause(ps, true, false)
		s.clauses = append(s.clauses, c)
		s.AttachClause(c)
		log.Println("AddClause: ps becomes a clause (two or more literals)", c)
		return true
	}
}

func (s *Solver) LitValue(p Lit) LBool {
	return s.assigns[p.Var()].Flip(p.Sign())
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

func (s *Solver) UncheckedEnqueue(p Lit, c *Clause) {
	s.assigns[p.Var()] = NewLBool(!p.Sign())
	s.vardata[p.Var()] = VarData{reason: c, level: s.decisionLevel()}
	s.trail = append(s.trail, p)
	log.Println("UncheckedEnqueue: Variable", p.Var(), "is assinged as", s.assigns[p.Var()])
}

func (s *Solver) RemoveSatisfied(cs []*Clause) []*Clause {
	j := 0
	for _, c := range cs {
		if s.Satisfied(c) {
			s.RemoveClause(c)
		} else {
			// Trim clause
			log.Println("RemoveSatisfied assertion: c.LitValue(c.lits[0]) == LUndef && c.LitValue(c.lits[1]) == LUndef", (s.LitValue(c.lits[0]) != LTrue && s.LitValue(c.lits[0]) != LFalse) && (s.LitValue(c.lits[1]) != LTrue && s.LitValue(c.lits[1]) != LFalse))
			for k := 2; k < len(c.lits); k++ {
				if s.LitValue(c.lits[k]) == LFalse {
					log.Println("RemoveSatisfied: Remove a literal that becomes false", c.lits[k])
					c.lits[k] = c.lits[len(c.lits)-1]
					c.lits = c.lits[:len(c.lits)-1]
				}
			}
			cs[j] = c
			j++
		}
	}
	cs = cs[:len(cs)-j]
	return cs
}

func (s *Solver) RemoveClause(c *Clause) {
	log.Println("RemoveCluase: Remove the clause", c)
	s.DetachClause(c, false)
	if s.Locked(c) {
		vdat := s.vardata[c.lits[0].Var()]
		s.vardata[c.lits[0].Var()] = VarData{reason: nil, level: vdat.level}
	}
}

func (s *Solver) AttachClause(c *Clause) {
	s.watches[c.lits[0].Not()] = append(s.watches[c.lits[0].Not()], Watcher{c, c.lits[1]})
	log.Println("AttachClause: Attach a watcher", Watcher{c, c.lits[1]}, "to a literal", c.lits[0].Not())
	s.watches[c.lits[1].Not()] = append(s.watches[c.lits[1].Not()], Watcher{c, c.lits[0]})
	log.Println("AttachClause: Attach a watcher", Watcher{c, c.lits[0]}, "to a literal", c.lits[1].Not())
	if c.header.learnt {
		s.numLearntes++
		s.learntsLiterals += uint64(len(c.lits))
	} else {
		s.numClauses++
		s.clausesLiterals += uint64(len(c.lits))
	}
}

// Detach a clause to watcher lists. If strict = true, the clause is immediately removed.
// Otherwise, the clause may be removed when GC runs. The default is strict = false
func (s *Solver) DetachClause(c *Clause, strict bool) {
	// TODO: strict or lazy detaching
	if strict {
		// remove(s.watches[c.lits[0].Not()], Watcher{c, c.lits[1]})
		// remove(s.watches[c.lits[1].Not()], Watcher{c, c.lits[0]})
	} else {
		// // check dirtybit
		// s.watches.smudge(c.lits[0].Not())
		// s.watches.smudge(c.lits[1].Not())
	}

	if c.header.learnt {
		s.numLearntes--
		s.learntsLiterals -= uint64(len(c.lits))
	} else {
		s.numClauses--
		s.clausesLiterals -= uint64(len(c.lits))
	}
}

// Return true if a clause is a reason for some implication in the currrent state
func (s *Solver) Locked(c *Clause) bool {
	if vdat, ok := s.vardata[c.lits[0].Var()]; ok {
		return s.LitValue(c.lits[0]) == LTrue && vdat.reason == c
	} else {
		return false
	}
}

// Return true if a clause is satisfied in the current state
func (s *Solver) Satisfied(c *Clause) bool {
	for _, lit := range c.lits {
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
//  the propagation queue is empty, even if there was a conflict.
//
func (s *Solver) Propagate() *Clause {
	var confl *Clause = nil
	numProps := 0

	for s.qhead < len(s.trail) {
		p := s.trail[s.qhead]
		s.qhead++
		ws := s.watches[p]
		numProps++
		log.Println("Propagate: Check an assigned literal", p, " ", ws)

		i := 0
		j := 0
		for i < len(ws) {
			log.Println("Propagate: Check a watcher", ws[i])
			// Try to avoid inspecting the clause
			blocker := ws[i].blocker
			if s.LitValue(blocker) == LTrue {
				ws[j] = ws[i]
				i++
				j++
				continue
			}

			// Make sure the false literal is data[1]
			c := ws[i].clause
			falseLit := p.Not()
			if c.lits[0] == falseLit {
				c.lits[0], c.lits[1] = c.lits[1], falseLit
			}
			i++

			// If 0th watch is true, then clause is already satisfied.
			first := c.lits[0]
			w := Watcher{c, first}
			if first != blocker && s.LitValue(first) == LTrue {
				ws[j] = w
				log.Println("Propagate: Attach a new watcher", w, "to a literal", p)
				j++
				continue
			}

			// Look for new watch
			for k := 2; k < len(c.lits); k++ {
				if s.LitValue(c.lits[k]) != LFalse {
					c.lits[1], c.lits[k] = c.lits[k], falseLit
					s.watches[c.lits[1].Not()] = append(s.watches[c.lits[1].Not()], w)
					log.Println("Propagate: Attach a new watcher", w, "to a literal", c.lits[1].Not())
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
		s.watches[p] = ws[:j]
	}
	s.Propagations += uint64(numProps)
	s.simpDBProps -= int64(numProps)
	return confl
}

//
// simplify
// Simplify the clause database according to the current top-level assignment.
// Currently, the only thing done here is the removal of satisfied clauses, but
// more things can be put here.

func (s *Solver) Simplify() bool {
	if s.ok == false || s.Propagate() != nil {
		s.ok = false
		return false
	}

	if len(s.trail) == s.simpDBAssigns || s.simpDBProps > 0 {
		log.Println("Simplify: The result is true because len(s.trail) == s.simpDBAssigns || s.simpDBProps > 0")
		return true
	}

	seen := make(map[Var]struct{})

	// Remove satisfied clauses
	s.learnts = s.RemoveSatisfied(s.learnts)
	if s.removeSatisfied { // s.removeSatisfied is always true for the time being
		s.clauses = s.RemoveSatisfied(s.clauses)

		log.Println("Simplify: The released variables: ", s.releasedVars)
		// Remove all released variables from the trail
		for _, v := range s.releasedVars {
			seen[v] = struct{}{}
		}

		j := 0
		for _, lit := range s.trail {
			if _, ok := seen[lit.Var()]; ok {
				s.trail[j] = lit
				j++
			}
		}
		s.trail = s.trail[:j]
		s.qhead = len(s.trail)
		s.freeVars = append(s.freeVars, s.releasedVars...)
		s.releasedVars = s.releasedVars[:0]
	}
	// checkGarbage()
	// rebuildOrderHeap()

	s.simpDBAssigns = len(s.trail)
	s.simpDBProps = int64(s.clausesLiterals) + int64(s.learntsLiterals) // shouldn't depend on stats really, but it will do for now

	return true
}

func (s *Solver) Solve(options *SolverOptions) LBool {
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

	s.learntsizeAdjustConfl = options.LearntsizeAdjustStartConfl
	s.learntsizeAdjustCnt = int(s.learntsizeAdjustConfl)
	status := LUndef

	log.Println("==== Search Statistics ====")

	// Search
	currRestarts := 0
	for status != LTrue && status != LFalse { // this means status == LUndef
		var resetBase float64
		if options.LubyRestart {
			resetBase = luby(options.RestartInc, currRestarts)
		} else {
			resetBase = math.Pow(options.RestartInc, float64(currRestarts))
		}

		status = s.search(int(resetBase*float64(options.RestartFirst)), options)
		if s.withinBudget() == false {
			break
		}
		currRestarts++
	}

	if status == LTrue {
		// Extend & copy model
		for k, v := range s.assigns {
			s.model[k] = v
		}
	} else if status == LFalse && len(s.conflict) == 0 {
		s.ok = false
	}
	return status
}

// search
// Search for a model the specified number of conflicts.
// Note: Use negative value for nof_conflicts indicate infinity
//
// Output
//  LTrue if a partial assigment that is consistent with respect to the clauseset if found.
//  If all variables are decision variables, this means that the clause set is satisfiable.
//  LFalse if the clause set is insatisfiable. LUndef if the bound on number of conflicts is reached.
//
func (s *Solver) search(nofConflicts int, options *SolverOptions) LBool {
	// backtranckLevel := 0
	conflictC := 0
	s.Starts++

	// for k := 0; k < 5; k++ { // for test
	for {
		if confl := s.Propagate(); confl != nil {
			log.Println("search: Find a conflict", confl)
			// conflict
			//
			s.Conflicts++
			conflictC++
			if s.decisionLevel() == 0 {
				return LFalse
			}
			learntClause, backtranckLevel := s.analyze(confl, options)
			s.cancelUntil(backtranckLevel, options)
			log.Println("Propagete: The result of analyze", learntClause, backtranckLevel)

			if len(learntClause) == 1 {
				s.UncheckedEnqueue(learntClause[0], nil)
			} else {
				c := MkClause(learntClause, false, true) // learnt: ture
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

				log.Println("||")
			}
		} else {
			log.Println("search: No conflict")

			if (nofConflicts >= 0 && conflictC >= nofConflicts) || !s.withinBudget() {
				log.Println("search: Reached bound on number of conflicts")
				s.Progress = s.progressEstimate()
				s.cancelUntil(0, options)
				return LUndef
			}

			//simplify the set of problem clauses
			if s.decisionLevel() == 0 && s.Simplify() == false {
				log.Println("search: Simplified problem has a conflict (UNSAT)")
				return LFalse
			}

			if float64(len(s.learnts)-len(s.trail)) >= s.maxLearnts {
				log.Println("search: Reduce the set of learnt clauses")
				s.reduceDB()
			}

			next := LitUndef
			for s.decisionLevel() < len(s.assumptions) {
				log.Println("search: Perform user provided assumption")
				p := s.assumptions[s.decisionLevel()]
				switch s.LitValue(p) {
				case LTrue:
					// Dummy decision level
					s.newDecisionLevel()
				case LFalse:
					//s.analyzeFinal(p.Not(), conflict)
					return LFalse
				default:
					next = p
					break
				}
			}

			if next == LitUndef {
				log.Println("search: New variable decision")
				s.Decisions++
				next = s.pickBranchLit(options)
				if next == LitUndef {
					log.Println("search: Model found", s.assigns)
					return LTrue
				}
			}

			log.Println("search: Increase decision level and enqueue next", next)
			s.newDecisionLevel()
			s.UncheckedEnqueue(next, nil)
		}
	}
	return LUndef
}

func (s *Solver) pickBranchLit(options *SolverOptions) Lit {
	next := VarUndef

	// Random decision
	if drand(&options.RandomSeed) < options.RandomVarFreq && s.orderHeap.IsEmpty() == false {
		next = s.orderHeap.heap[irand(&options.RandomSeed, len(s.orderHeap.heap))]
		if (s.assigns[next] != LTrue || s.assigns[next] != LFalse) && s.decision[next] == true {
			s.RndDecisions++
		}
	}
	log.Println("pickBranchLit: Random choose", next)

	// Activity based decision
	for next == VarUndef || s.assigns[next] == LTrue || s.assigns[next] == LFalse || s.decision[next] == false {
		if s.orderHeap.IsEmpty() {
			next = VarUndef
			break
		} else {
			next = s.orderHeap.RemoveMin()
			// log.Println("pickBranchLit:", s.orderHeap)
		}
	}
	log.Println("pickBranchLit: Active based choose", next)

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
	log.Println("newDecisionLevel: decision level", s.decisionLevel())
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
func (s *Solver) reduceDB() {
	extraLim := s.claInc / float64(len(s.learnts))
	sort.Slice(s.learnts, func(i, j int) bool {
		return len(s.learnts[i].lits) > 2 && (len(s.learnts[j].lits) == 2 || s.learnts[i].activity < s.learnts[j].activity)
	})
	// Do not delete binary or locked clauses. From the rest, delete clauses from the first half
	// and clauses with activity smaller than extraLim
	j := 0
	for i := 0; i < len(s.learnts); i++ {
		c := s.learnts[i]
		if len(c.lits) > 2 && !s.Locked(c) && (i < len(s.learnts)/2 || c.activity < extraLim) {
			s.RemoveClause(c)
		} else {
			s.learnts[j] = s.learnts[i]
			j++
		}
	}
	s.learnts = s.learnts[:j]
	// checkGarbage()
}

func (s *Solver) withinBudget() bool {
	return !s.asynchInterrupt && (s.conflictBudget < 0 || s.Conflicts < uint64(s.conflictBudget)) && (s.propagationBudget < 0 || s.Propagations < uint64(s.propagationBudget))
}

// Increase a clause with the current bump value
func (s *Solver) claBumpActivity(c *Clause) {
	c.activity += s.claInc
	if c.activity > 1e20 {
		// rescale
		for i, _ := range s.learnts {
			s.learnts[i].activity *= 1e-20
		}
		s.claInc *= 1e-20
	}
}

//
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
		// log.Println("varBumpActivity:", s.orderHeap)
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
			// log.Println("cancelUntil:", s.orderHeap)
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

func (s *Solver) analyze(c *Clause, options *SolverOptions) ([]Lit, int) {
	pathC := 0
	p := LitUndef
	outLearnt := make([]Lit, 1, len(c.lits)) // outLeant[0] will be put at the end of this function

	seen := make(map[Var]byte)

	// Generate conflict clause
	index := len(s.trail) - 1
	for {
		if c.header.learnt {
			s.claBumpActivity(c)
		}

		var j int
		if p == LitUndef {
			j = 0
		} else {
			j = 1
		}
		for ; j < len(c.lits); j++ {
			v := c.lits[j].Var()
			if _, ok := seen[v]; ok == false {
				if s.vardata[v].level > 0 {
					s.varBumpActivity(v)
					seen[v] = 1
					if s.vardata[v].level >= s.decisionLevel() {
						pathC++
					} else {
						outLearnt = append(outLearnt, c.lits[j])
					}
				}
			}
		}

		for {
			p = s.trail[index]
			if _, ok := seen[p.Var()]; ok {
				c = s.vardata[p.Var()].reason
				delete(seen, p.Var())
				break
			} else {
				index--
			}
		}
		pathC--

		// log.Println("analyze: pathC", pathC)
		if pathC == 0 {
			break
		}
	}
	outLearnt[0] = p.Not()

	log.Println("analyze: outLearnt", outLearnt)

	// simplify conflict clause
	switch {
	case options.CcminMode == 2:
		// log.Println("analyze: ccmode 2")
		j := 1
		for i := 1; i < len(outLearnt); i++ {
			p := outLearnt[i]
			if s.vardata[p.Var()].reason == nil || s.litRedundant(p, seen) == false {
				outLearnt[j] = p
				j++
			}
		}
		s.maxLiterals += uint64(len(outLearnt))
		outLearnt = outLearnt[:j]
		s.totLiterals += uint64(len(outLearnt))
	case options.CcminMode == 1:
		// log.Println("analyze: ccmode 1")
		j := 1
		for i := 1; i < len(outLearnt); i++ {
			p := outLearnt[i]
			if c := s.vardata[p.Var()].reason; c == nil {
				outLearnt[j] = p
				j++
			} else {
				for k := 1; k < len(c.lits); k++ {
					v := c.lits[k].Var()
					if _, ok := seen[v]; ok == false && s.vardata[v].level > 0 {
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
		log.Println("analyze: ccmode 0")
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

	return outLearnt, outBtlevel
}

// This is used in litRedundant
type redundantStackElem struct {
	i int
	l Lit
}

// Check if p can be removed from a conflict clause
func (s *Solver) litRedundant(p Lit, seen map[Var]byte) bool {
	// seen
	//   0: undef (key does not exist)
	//   1: seen_source
	//   2: seen_removable
	//   3: seen_failed
	//

	// log.Println("litRedundant assertion (seen[var(p)] == seen_undef || seen[var(p)] == seen_source):", seen[p.Var()] == 0 || seen[p.Var()] == 1)
	// log.Println("litRedundant assertion (reason(var(p)) != nil):", s.vardata[p.Var()].reason != nil)

	stack := make([]redundantStackElem, 0, 10)
	c := s.vardata[p.Var()].reason
	i := 1
	for {
		if i < len(c.lits) {
			l := c.lits[i]

			// Variable at level 0 or previsouly removable
			if s.vardata[l.Var()].level == 0 || seen[l.Var()] == 1 || seen[l.Var()] == 2 {
				goto nextLoop
			}

			// Check variable cannot be removed for some local reason
			if s.vardata[l.Var()].reason == nil || seen[l.Var()] == 3 {
				stack = append(stack, redundantStackElem{0, p})
				for _, ss := range stack {
					if _, ok := seen[ss.l.Var()]; ok == false {
						seen[ss.l.Var()] = 3
					}
				}
				return false
			}

			// Recursively check l
			stack = append(stack, redundantStackElem{i, p})
			i, p, c = 0, l, s.vardata[p.Var()].reason
		} else {
			// Finished with current element p and reason c
			if _, ok := seen[p.Var()]; ok == false {
				seen[p.Var()] = 2
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

	return true
}