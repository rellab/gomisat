package sat

import (
	"errors"
	_ "fmt"
	"log"
	"math"
	"sort"
	_ "sync"
)

const (
	debug          = false
	debugOrderHeap = true
	debugAssert    = true
)

var (
	ErrReachedBound = errors.New("search: Reached bound on number of conflicts")
	ErrOutOfBudget  = errors.New("search: Out of budget")
)

type SolverOptions struct {
	Verbosity                  int
	VarDecay                   float64
	ClauseDecay                float64
	RandomVarFreq              float64
	RandomSeed                 float64
	LubyRestart                bool
	CCMinMode                  int     // Controls conflict clause minimization (0=none, 1=basic, 2=deep).
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
	StrictDetachClause         bool
}

func DefaultSolverOptions() *SolverOptions {
	return &SolverOptions{
		Verbosity:                  0,
		VarDecay:                   0.95,
		ClauseDecay:                0.999,
		RandomVarFreq:              0,
		RandomSeed:                 91648253,
		LubyRestart:                true,
		CCMinMode:                  2,
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
		StrictDetachClause:         false,
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

	userPol   map[Var]bool  // The users preferred polarity of each variable.
	activity  []float64     // A heuristic measurement of the activity of a variable.
	assigns   LitAssignment // The current assignments.
	polarity  []bool        // The preferred polarity of each variable.
	decision  []bool        // Declares if a variable is eligible for selection in the decision heuristic.
	vardata   []VarData     // Stores reason and level for each variable.
	watches   [][]*Watcher  // watched[lit] is a list of constraints watching 'lit' (will go there if literal becomes true).
	orderHeap *VarHeap      // A priority queue of variables ordered with respect to the variable activity.

	maxLearnts            float64
	learntsizeAdjustConfl float64
	learntsizeAdjustCnt   int

	decVars         uint
	numClauses      uint
	numLearntes     uint
	clausesLiterals uint
	learntsLiterals uint
	maxLiterals     uint
	totLiterals     uint

	// solver state
	ok            bool // If false, the constraints are already unsatisfiable. No part of the solver state may be used
	qhead         int  // Head of queue (as index into the trail; no more explicit propagation queue in MiniSat)
	simpDBProps   int  // Remaining number of propatations that must be made before next execution of 'simplify()'
	simpDBAssigns int  // Number of top-level assignments since last execution of simplify()

	conflict map[Lit]struct{}
	// model    map[Var]bool

	claInc float64 // Amount to bump next clause with
	varInc float64 // Amount to bump next variable with

	nextVar      Var
	releasedVars []Var
	freeVars     []Var

	conflictBudget    int
	propagationBudget int
	asynchInterrupt   bool

	Solves       uint
	Starts       uint
	Decisions    uint
	Propagations uint
	Conflicts    uint
	RndDecisions uint
	Progress     float64

	stack []redundantStackElem
}

func NewSolver() *Solver {
	s := &Solver{
		clauses:           make([]*Clause, 0),
		learnts:           make([]*Clause, 0),
		trail:             make([]Lit, 0),
		trailLim:          make([]int, 0),
		assumptions:       make([]Lit, 0),
		activity:          make([]float64, 0),
		assigns:           make(LitAssignment, 0),
		polarity:          make([]bool, 0),
		userPol:           make(map[Var]bool),
		decision:          make([]bool, 0),
		vardata:           make([]VarData, 0),
		watches:           make([][]*Watcher, 0),
		releasedVars:      make([]Var, 0),
		freeVars:          make([]Var, 0),
		ok:                true,
		qhead:             0,
		simpDBAssigns:     -1,
		simpDBProps:       0,
		nextVar:           0,
		conflictBudget:    -1,
		propagationBudget: -1,
		asynchInterrupt:   false,
		claInc:            1,
		varInc:            1,
		stack:             make([]redundantStackElem, 0, 10),
	}
	s.orderHeap = NewVarHeap(func(x, y Var) bool {
		return s.activity[x] > s.activity[y]
	})
	return s
}

func (s Solver) setDecisionVar(v Var, b bool) {
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
//   upol: Assinged value for a variable. The default is LUndef
//   dvar: Indicator whether a variable is to be determined. The default is true.
func (s *Solver) NewVar(dvar bool, options *SolverOptions) Var {
	var v Var
	n := len(s.freeVars)
	if n > 0 {
		v = s.freeVars[n-1]
		s.freeVars = s.freeVars[:n-1]
		s.assigns[v] = LUndef
		s.vardata[v] = VarData{reason: nil, level: 0}
		if options.RndInitAct {
			s.activity[v] = drand(&options.RandomSeed) * 0.00001
		} else {
			s.activity[v] = 0
		}
		s.polarity[v] = true
		// s.userPol[v] = upol
		s.decision[v] = dvar
		s.watches[v+v] = make([]*Watcher, 0)
		s.watches[v+v+1] = make([]*Watcher, 0)
		s.orderHeap.indicies[v] = UndefIndex
		if dvar {
			s.orderHeap.Insert(v)
		}
	} else {
		v = s.nextVar
		s.nextVar++
		s.assigns = append(s.assigns, LUndef)
		s.vardata = append(s.vardata, VarData{reason: nil, level: 0})
		if options.RndInitAct {
			s.activity = append(s.activity, drand(&options.RandomSeed)*0.00001)
		} else {
			s.activity = append(s.activity, 0)
		}
		s.polarity = append(s.polarity, true)
		// s.userPol = append(s.userPol, upol)
		s.decision = append(s.decision, dvar)
		s.watches = append(s.watches, make([]*Watcher, 0))
		s.watches = append(s.watches, make([]*Watcher, 0))
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

func simplifyLiteral(a []LBool, ps []Lit) ([]Lit, bool) {
	sort.Slice(ps, func(i, j int) bool {
		return ps[i] < ps[j]
	})
	j := 0
	p := ps[0]
	for _, x := range ps[1:] {
		v := a[p.Var()]
		if v == LTrue || p.Not() == x {
			return nil, true
		} else if v != LFalse && p != x {
			p = x
			ps[j] = x
			j++
		}
	}
	return ps[:j], false
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
		if s.assigns.IsTrue(ps[i]) || ps[i] == p.Not() {
			return true
		} else if s.assigns.IsUndef(ps[i]) && ps[i] != p {
			p, ps[j] = ps[i], ps[i]
			j++
		}
	}
	ps = ps[:j] // shrink
	// var ok bool
	// if ps, ok = simplifyLiteral(s.assigns, ps); ok == true {
	// 	return true
	// }
	switch {
	case len(ps) == 0:
		s.ok = false
		if debug {
			log.Println("AddClause: ps becomes empty")
		}
		return false
	case len(ps) == 1:
		s.uncheckedEnqueue(ps[0], nil)
		if confl := s.Propagate(); confl == nil {
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
	default:
		// set clause
		c := MkClause(ps, false)
		s.clauses = append(s.clauses, c)
		s.AttachClause(c)
		if debug {
			log.Println("AddClause: ps becomes a clause (two or more literals)", c)
		}
		return true
	}
}

// func (s *Solver) litValue(p Lit) LBool {
// 	if p.Sign() == true {
// 		return s.assigns[p.Var()].Not()
// 	} else {
// 		return s.assigns[p.Var()]
// 	}
// }

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

func (s *Solver) uncheckedEnqueue(p Lit, c *Clause) {
	s.assigns[p.Var()] = NewLBool(!p.Sign())
	s.vardata[p.Var()] = VarData{reason: c, level: s.decisionLevel()}
	s.trail = append(s.trail, p)
	if debug {
		log.Println("uncheckedEnqueue: Variable", p.Var(), "is assinged as", s.assigns[p.Var()])
	}
}

func (s *Solver) removeSatisfied(cs []*Clause, options *SolverOptions) []*Clause {
	j := 0
	for _, c := range cs {
		if s.Satisfied(c) {
			s.RemoveClause(c, options)
		} else {
			// Trim clause
			if debug && debugAssert {
				log.Println("removeSatisfied assertion: c.litValue(c.lits[0]) == LUndef && c.litValue(c.lits[1]) == LUndef", s.assigns.IsUndef(c.lits[0]) && s.assigns.IsUndef(c.lits[1]))
			}
			k := 2
			for k < len(c.lits) {
				if s.assigns.IsFalse(c.lits[k]) {
					if debug {
						log.Println("removeSatisfied: Remove a literal that becomes false", c.lits[k])
					}
					c.lits[k] = c.lits[len(c.lits)-1]
					c.lits = c.lits[:len(c.lits)-1]
				} else {
					k++
				}
			}
			cs[j] = c
			j++
		}
	}
	cs = cs[:len(cs)-j]
	return cs
}

func (s *Solver) RemoveClause(c *Clause, options *SolverOptions) {
	if debug {
		log.Println("RemoveCluase: Remove the clause", c)
	}
	s.DetachClause(c, options)
	if s.Locked(c) {
		vdat := s.vardata[c.lits[0].Var()]
		s.vardata[c.lits[0].Var()] = VarData{reason: nil, level: vdat.level}
	}
}

func (s *Solver) AttachClause(c *Clause) {
	s.watches[c.lits[0].Not()] = append(s.watches[c.lits[0].Not()], &Watcher{c, c.lits[1]})
	s.watches[c.lits[1].Not()] = append(s.watches[c.lits[1].Not()], &Watcher{c, c.lits[0]})
	if debug {
		log.Println("AttachClause: Attach a watcher", Watcher{c, c.lits[1]}, "to a literal", c.lits[0].Not())
		log.Println("AttachClause: Attach a watcher", Watcher{c, c.lits[0]}, "to a literal", c.lits[1].Not())
	}
	s.numClauses++
	s.clausesLiterals += uint(len(c.lits))
}

func (s *Solver) AttachLearntClause(c *Clause) {
	s.watches[c.lits[0].Not()] = append(s.watches[c.lits[0].Not()], &Watcher{c, c.lits[1]})
	s.watches[c.lits[1].Not()] = append(s.watches[c.lits[1].Not()], &Watcher{c, c.lits[0]})
	if debug {
		log.Println("AttachLearntClause: Attach a watcher", Watcher{c, c.lits[1]}, "to a literal", c.lits[0].Not())
		log.Println("AttachLearntClause: Attach a watcher", Watcher{c, c.lits[0]}, "to a literal", c.lits[1].Not())
	}
	s.numLearntes++
	s.learntsLiterals += uint(len(c.lits))
}

// Detach a clause to watcher lists. If strict = true, the clause is immediately removed.
// Otherwise, the clause may be removed when GC runs. The default is strict = false
func (s *Solver) DetachClause(c *Clause, options *SolverOptions) {
	if options.StrictDetachClause {
		s.watches[c.lits[0].Not()] = RemoveWatcher(s.watches[c.lits[0].Not()], c) //Watcher{c, c.lits[1]})
		s.watches[c.lits[1].Not()] = RemoveWatcher(s.watches[c.lits[1].Not()], c) //Watcher{c, c.lits[0]})
	} else {
		// check dirtybit
		c.header.durty = true
	}

	if c.header.learnt {
		s.numLearntes--
		s.learntsLiterals -= uint(len(c.lits))
	} else {
		s.numClauses--
		s.clausesLiterals -= uint(len(c.lits))
	}
}

func RemoveWatcher(ws []*Watcher, c *Clause) []*Watcher {
	j := 0
	for _, w := range ws {
		if w.clause != c {
			ws[j] = w
			j++
		}
	}
	return ws[:j]
}

// Return true if a clause is a reason for some implication in the currrent state
func (s *Solver) Locked(c *Clause) bool {
	vdat := s.vardata[c.lits[0].Var()]
	return s.assigns.IsTrue(c.lits[0]) && vdat.reason == c
}

// Return true if a clause is satisfied in the current state
func (s *Solver) Satisfied(c *Clause) bool {
	for _, p := range c.lits {
		if s.assigns.IsTrue(p) {
			return true
		}
	}
	return false
}

// func onePropagation(a LitAssignment, queue []Lit, p Lit, ws []*Watcher) ([]Lit, []*Watcher, *Clause) {
// 		nws := ws[:0]
// 		for i, w := range ws {
// 			if debug {
// 				log.Println("Propagate: Check a watcher", w)
// 			}

// 			// Try to avoid inspecting the clause
// 			blocker := w.blocker
// 			if a.IsTrue(blocker) {
// 				nws = append(nws, w)
// 				continue
// 			}

// 			// Make sure the false literal is data[1]
// 			c := w.clause
// 			falseLit := p.Not()
// 			if c.lits[0] == falseLit {
// 				c.lits[0], c.lits[1] = c.lits[1], falseLit
// 			}

// 			// If 0th watch is true, then clause is already satisfied.
// 			first := c.lits[0]
// 			w.blocker = first
// 			if first != blocker && a.IsTrue(first) {
// 				if debug {
// 					log.Println("Propagate: Attach a new watcher", w, "to a literal", p)
// 				}
// 				nws = append(nws, w)
// 				continue
// 			}

// 			// Look for new watch
// 			for k := 2; k < len(c.lits); k++ {
// 				if a.IsFalse == false {
// 					c.lits[1], c.lits[k] = c.lits[k], falseLit
// 					s.watches[c.lits[1].Not()] = append(s.watches[c.lits[1].Not()], w)
// 					if debug {
// 						log.Println("Propagate: Attach a new watcher", w, "to a literal", c.lits[1].Not())
// 					}
// 					goto nextClause
// 				}
// 			}

// 			// Did not find watch -- clause is unit under assignment
// 			nws = append(nws, w)

// 			if a.IsFalse(first) {
// 				for _, w := range ws[i:] {
// 					nws = append(nws, w)
// 				}
// 				return queue[:0], nws, c
// 			} else {
// 				// queue = append(queue, first)s.uncheckedEnqueue(first, c)
// 			}
// 		}
// 		return queue, nws, nil
// 	}
// }

// Perform unit propagation. Return possibly conflicting clause.
// Propagate all enqueued facts. If a conflict arises, the conflicting clause is returned,
// otherwise nil (CRef_Undef)
//
// Post condition
//  the propagation queue is empty, even if there was a conflict.
//

// func (s *Solver) Propagate() *Clause {
// 	var confl *Clause = nil
// 	numProps := 0

// 	for s.qhead < len(s.trail) {
// 		p := s.trail[s.qhead]
// 		s.qhead++
// 		ws := s.watches[p]
// 		numProps++
// 		if debug {
// 			log.Println("Propagate: Check an assigned literal", p, " ", ws)
// 		}

// 		i := 0
// 		j := 0
// 		for i < len(ws) {
// 			if debug {
// 				log.Println("Propagate: Check a watcher", ws[i])
// 			}

// 			if ws[i].clause.header.durty {
// 				i++
// 				continue
// 			}

// 			// Try to avoid inspecting the clause
// 			blocker := ws[i].blocker
// 			if s.assigns.IsTrue(blocker) {
// 				ws[j] = ws[i]
// 				i++
// 				j++
// 				continue
// 			}

// 			// Make sure the false literal is data[1]
// 			c := ws[i].clause
// 			falseLit := p.Not()
// 			if c.lits[0] == falseLit {
// 				c.lits[0], c.lits[1] = c.lits[1], falseLit
// 			}
// 			i++

// 			// If 0th watch is true, then clause is already satisfied.
// 			first := c.lits[0]
// 			ws[i-1].blocker = first
// 			w := ws[i-1] // Watcher{c, first}
// 			if first != blocker && s.assigns.IsTrue(first) {
// 				ws[j] = w
// 				if debug {
// 					log.Println("Propagate: Attach a new watcher", w, "to a literal", p)
// 				}
// 				j++
// 				continue
// 			}

// 			// Look for new watch
// 			for k := 2; k < len(c.lits); k++ {
// 				if s.assigns.IsFalse(c.lits[k]) == false {
// 					c.lits[1], c.lits[k] = c.lits[k], falseLit
// 					s.watches[c.lits[1].Not()] = append(s.watches[c.lits[1].Not()], w)
// 					if debug {
// 						log.Println("Propagate: Attach a new watcher", w, "to a literal", c.lits[1].Not())
// 					}
// 					goto nextClause
// 				}
// 			}

// 			// Did not find watch -- clause is unit under assignment
// 			ws[j] = w
// 			j++
// 			if s.assigns.IsFalse(first) {
// 				confl = c
// 				s.qhead = len(s.trail)
// 				// copy the remaining watches
// 				for i < len(ws) {
// 					ws[j] = ws[i]
// 					j++
// 					i++
// 				}
// 			} else {
// 				s.uncheckedEnqueue(first, c)
// 			}

// 		nextClause:
// 		}
// 		if debug {
// 			log.Printf("Propagate: shrink watchers; lit %s i-j %d\n", p.String(), len(ws)-j)
// 		}
// 		s.watches[p] = ws[:j]
// 	}
// 	s.Propagations += uint(numProps)
// 	s.simpDBProps -= int(numProps)
// 	return confl
// }

func (s *Solver) Propagate() *Clause {
	var confl *Clause = nil
	numProps := 0

	for s.qhead < len(s.trail) {
		p := s.trail[s.qhead]
		s.qhead++
		ws := s.watches[p]
		numProps++
		if debug {
			log.Println("Propagate: Check an assigned literal", p, " ", ws)
		}

		nws := ws[:0]
		for _, w := range ws {
			if debug {
				log.Println("Propagate: Check a watcher", w)
			}

			blocker := w.blocker
			c := w.clause

			switch {
			case w.clause.header.durty:
				continue
			case confl != nil:
				nws = append(nws, w)
				continue
			case s.assigns.IsTrue(blocker):
				nws = append(nws, w)
				continue
			}

			// Make sure the false literal is data[1]
			falseLit := p.Not()
			if c.lits[0] == falseLit {
				c.lits[0], c.lits[1] = c.lits[1], falseLit
			}

			// If 0th watch is true, then clause is already satisfied.
			first := c.lits[0]
			w.blocker = first
			if first != blocker && s.assigns.IsTrue(first) {
				nws = append(nws, w)
				if debug {
					log.Println("Propagate: Attach a new watcher", w, "to a literal", p)
				}
				continue
			}

			// Look for new watch
			for k := 2; k < len(c.lits); k++ {
				if s.assigns.IsFalse(c.lits[k]) == false {
					c.lits[1], c.lits[k] = c.lits[k], falseLit
					s.watches[c.lits[1].Not()] = append(s.watches[c.lits[1].Not()], w)
					if debug {
						log.Println("Propagate: Attach a new watcher", w, "to a literal", c.lits[1].Not())
					}
					goto nextClause
				}
			}

			// Did not find watch -- clause is unit under assignment
			nws = append(nws, w)
			if s.assigns.IsFalse(first) {
				confl = c
				s.qhead = len(s.trail)
				// // copy the remaining watches
				// for _, w := range ws[i+1:] {
				// 	nws = append(nws, w)
				// }
			} else {
				s.uncheckedEnqueue(first, c)
			}

		nextClause:
		}
		s.watches[p] = nws
	}
	s.Propagations += uint(numProps)
	s.simpDBProps -= int(numProps)
	return confl
}

//
// simplify
// Simplify the clause database according to the current top-level assignment.
// Currently, the only thing done here is the removal of satisfied clauses, but
// more things can be put here.

func (s *Solver) Simplify(options *SolverOptions) bool {
	if s.ok == false || s.Propagate() != nil {
		s.ok = false
		return false
	}

	if len(s.trail) == s.simpDBAssigns || s.simpDBProps > 0 {
		//		log.Println("Simplify: The result is true because len(s.trail) == s.simpDBAssigns || s.simpDBProps > 0")
		return true
	}

	seen := make(map[Var]struct{})

	// Remove satisfied clauses
	s.learnts = s.removeSatisfied(s.learnts, options)
	s.clauses = s.removeSatisfied(s.clauses, options)

	if debug {
		log.Println("Simplify: The released variables: ", s.releasedVars)
	}
	// Remove all released variables from the trail
	for _, v := range s.releasedVars {
		seen[v] = struct{}{}
	}

	j := 0
	for _, p := range s.trail {
		if _, ok := seen[p.Var()]; ok {
			s.trail[j] = p
			j++
		}
	}
	s.trail = s.trail[:j]
	s.qhead = len(s.trail)
	s.freeVars = append(s.freeVars, s.releasedVars...)
	s.releasedVars = s.releasedVars[:0]

	// checkGarbage()
	s.rebuildOrderHeap()

	s.simpDBAssigns = len(s.trail)
	s.simpDBProps = int(s.clausesLiterals) + int(s.learntsLiterals) // shouldn't depend on stats really, but it will do for now

	return true
}

func (s *Solver) rebuildOrderHeap() {
	vs := make([]Var, 0, s.nextVar)
	for v := Var(0); v < s.nextVar; v++ {
		if s.decision[v] && s.assigns.IsVarUndef(v) {
			vs = append(vs, v)
		}
	}
	s.orderHeap.Build(vs)
	if debug && debugOrderHeap {
		log.Println("rebuildOrderHeap: ", s.orderHeap.heap)
	}
}

func resetBase(currRestarts int, options *SolverOptions) int {
	if options.LubyRestart {
		return int(luby(options.RestartInc, currRestarts) * options.RestartFirst)
	} else {
		return int(math.Pow(options.RestartInc, float64(currRestarts)) * options.RestartFirst)
	}
}

func (s *Solver) Solve(options *SolverOptions) (bool, error) {
	s.trail = make([]Lit, 0, s.nextVar)
	s.trailLim = make([]int, 0, s.nextVar)
	model := make(map[Var]bool)
	s.conflict = make(map[Lit]struct{})

	if s.ok == false {
		return false, nil
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

	//	log.Println("==== Search Statistics ====")

	// Search
	currRestarts := 0
	for {
		result, err := s.search(resetBase(currRestarts, options), options)
		switch {
		case result == true && err == nil:
			// Extend & copy model
			for k, _ := range s.assigns {
				if s.assigns.IsVarTrue(Var(k)) {
					model[Var(k)] = true
				} else if s.assigns.IsVarFalse(Var(k)) {
					model[Var(k)] = false
				}
			}
			return true, nil
		case result == false && err == nil:
			s.ok = false
			return false, nil
		case err == ErrOutOfBudget:
			return false, err
		case err == ErrReachedBound:
			s.Progress = s.progressEstimate()
			s.cancelUntil(0, options)
		}
		currRestarts++
	}
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
func (s *Solver) search(nofConflicts int, options *SolverOptions) (bool, error) {
	conflictC := 0
	s.Starts++

	for {
		if confl := s.Propagate(); confl != nil {
			if debug {
				log.Println("search: Find a conflict", confl)
			}
			s.Conflicts++
			conflictC++
			if s.decisionLevel() == 0 {
				return false, nil
			}
			learntClause, backtranckLevel := s.analyze(confl, options)
			s.cancelUntil(backtranckLevel, options)
			if debug {
				log.Println("Propagete: The result of analyze", learntClause, backtranckLevel)
			}

			if len(learntClause) == 1 {
				s.uncheckedEnqueue(learntClause[0], nil)
			} else {
				c := MkClause(learntClause, true) // learnt: ture
				s.learnts = append(s.learnts, c)
				s.AttachLearntClause(c)
				s.claBumpActivity(c)
				s.uncheckedEnqueue(learntClause[0], c)
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

			if nofConflicts >= 0 && conflictC >= nofConflicts {
				if debug {
					log.Println("search: Reached bound on number of conflicts")
				}
				return false, ErrReachedBound
			}

			if s.withinBudget() == false {
				if debug {
					log.Println("search: Out of budget")
				}
				return false, ErrOutOfBudget
			}

			//simplify the set of problem clauses
			if s.decisionLevel() == 0 && s.Simplify(options) == false {
				if debug {
					log.Println("search: Simplified problem has a conflict (UNSAT)")
				}
				return false, nil
			}

			if float64(len(s.learnts)-len(s.trail)) >= s.maxLearnts {
				if debug {
					log.Println("search: Reduce the set of learnt clauses", len(s.learnts), len(s.trail), s.maxLearnts)
				}
				s.reduceDB(options)
			}

			next := LitUndef
			for s.decisionLevel() < len(s.assumptions) {
				if debug {
					log.Println("search: Perform user provided assumption")
				}
				p := s.assumptions[s.decisionLevel()]
				switch {
				case s.assigns.IsTrue(p):
					// Dummy decision level
					s.newDecisionLevel()
				case s.assigns.IsFalse(p):
					//s.analyzeFinal(p.Not(), conflict)
					return false, nil
				default:
					next = p
					break
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
					return true, nil
				}
			}

			if debug {
				log.Println("search: Increase decision level and enqueue next", next)
			}
			s.newDecisionLevel()
			s.uncheckedEnqueue(next, nil)
		}
	}
}

func (s *Solver) pickBranchLit(options *SolverOptions) Lit {
	next := VarUndef

	// Random decision
	if drand(&options.RandomSeed) < options.RandomVarFreq && s.orderHeap.IsEmpty() == false {
		next = s.orderHeap.heap[irand(&options.RandomSeed, len(s.orderHeap.heap))]
		if s.assigns.IsVarUndef(next) && s.decision[next] == true {
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
	for next == VarUndef || !s.assigns.IsVarUndef(next) || s.decision[next] == false {
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
func (s *Solver) reduceDB(options *SolverOptions) {
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
			s.RemoveClause(c, options)
		} else {
			s.learnts[j] = s.learnts[i]
			j++
		}
	}
	s.learnts = s.learnts[:j]
	if debug {
		log.Println("reduceDB: The number of new learnts", j)
	}
	// checkGarbage()
}

func (s *Solver) withinBudget() bool {
	return !s.asynchInterrupt && (s.conflictBudget < 0 || s.Conflicts < uint(s.conflictBudget)) && (s.propagationBudget < 0 || s.Propagations < uint(s.propagationBudget))
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
	case options.CCMinMode == 2:
		j := 1
		for _, p := range outLearnt[1:] {
			if s.vardata[p.Var()].reason == nil || s.litRedundant(p, seen) == false {
				outLearnt[j] = p
				j++
			}
		}
		s.maxLiterals += uint(len(outLearnt))
		outLearnt = outLearnt[:j]
		s.totLiterals += uint(len(outLearnt))
	case options.CCMinMode == 1:
		j := 1
		for _, p := range outLearnt[1:] {
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
		s.maxLiterals += uint(len(outLearnt))
		outLearnt = outLearnt[:j]
		s.totLiterals += uint(len(outLearnt))
	default:
		s.maxLiterals += uint(len(outLearnt))
		s.totLiterals += uint(len(outLearnt))
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

	if debug && debugAssert {
		log.Println("litRedundant assertion (seen[var(p)] == seen_undef || seen[var(p)] == seen_source):", seen[p.Var()] == 0 || seen[p.Var()] == 1)
		log.Println("litRedundant assertion (reason(var(p)) != nil):", s.vardata[p.Var()].reason != nil)
	}

	stack := s.stack[:0] //make([]redundantStackElem, 0, 10)
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
