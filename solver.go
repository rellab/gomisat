package gomisat

import (
	_ "log"
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
	MinLearntsLim              int     // Minimum number to set the learnts limit to.
	RestartFirst               int     // The initial restart limit. (default 100)
	RestartInc                 float64 // The factor with which the restart limit is multiplied in each restart. (default 1.5)
	LearntsizeFactor           float64 // The intitial limit for learnt clauses is a factor of the original clauses. (default 1 / 3)
	LeantsizeInt               float64 // The limit for learnt clauses is multiplied with this factor each restart. (default 1.1)
	LearntsizeAdjustStartConfl int
	LearntsizeAdjustInc        float64
}

type SolverStatistics struct {
	Solves       uint64
	Starts       uint64
	Decisions    uint64
	Propagations uint64
	Conflicts    uint64
}

func NewSolverStatistics() SolverStatistics {
	return SolverStatistics{}
}

type VarData struct {
	reason *Clause
	level  int
}

type SolverResourceContraints struct {
	conflictBudget    int64
	propagationBudget int64
	asynchInterrupt   bool
}

type Watcher struct {
	clause  *Clause
	blocker Lit
}

func (w *Watcher) Eq(other *Watcher) bool {
	return w.clause == other.clause
}

func (w *Watcher) Neq(other *Watcher) bool {
	return w.clause != other.clause
}

type Solver struct {
	clauses     []*Clause // List of problem clauses.
	learnts     []*Clause // List of learnt clauses.
	trail       []Lit     // Assignment stack; stores all assigments made in the order they were made.
	trailLim    []int     // Separator indices for different decision levels in 'trail'.
	assumptions []Lit     // Current set of assumptions provided to solve by the user.

	activity map[Var]float64   // A heuristic measurement of the activity of a variable.
	assigns  map[Var]LBool     // The current assignments.
	polarity map[Var]byte      // The preferred polarity of each variable.
	userPol  map[Var]LBool     // The users preferred polarity of each variable.
	decision map[Var]byte      // Declares if a variable is eligible for selection in the decision heuristic.
	vardata  map[Var]VarData   // Stores reason and level for each variable.
	watches  map[Lit][]Watcher // watched[lit] is a list of constraints watching 'lit' (will go there if literal becomes true).

	maxLearnts            float64
	learntsizeAdjustConfl float64
	learntsizeAdjustCnt   int

	stats SolverStatistics

	decVars         uint64
	numClauses      uint64
	numLearntes     uint64
	clausesLiterals uint64
	learntsLiterals uint64
	maxLiterals     uint64
	totLiterals     uint64

	// solver state
	ok          bool  // If false, the constraints are already unsatisfiable. No part of the solver state may be used
	qhead       int   // Head of queue (as index into the trail; no more explicit propagation queue in MiniSat)
	simpDBProps int64 // Remaining number of propatations that must be made before next execution of 'simplify()'
	nextVar     Var

	releasedVars []Var
	freeVars     []Var

	// Temporarie
	seen           map[Var]byte
	analyzeStack   []int
	analyzeToClear []Lit
	addTmp         []Lit
}

func NewSolver() *Solver {
	return &Solver{
		activity:     make(map[Var]float64),
		assigns:      make(map[Var]LBool),
		polarity:     make(map[Var]byte),
		userPol:      make(map[Var]LBool),
		decision:     make(map[Var]byte),
		vardata:      make(map[Var]VarData),
		watches:      make(map[Lit][]Watcher),
		nextVar:      0,
		releasedVars: make([]Var, 0),
		freeVars:     make([]Var, 0),
		ok:           true,
		qhead:        0,
	}
}

func (s *Solver) setDecisionVar(v Var, b bool) {
	if b && s.decision[v] == 0 {
		s.decVars++
		s.decision[v] = 1
	} else if !b && s.decision[v] != 0 {
		s.decVars--
		s.decision[v] = 0
	}
	// insertVarOrder(v)
}

// Add a new variable with parameters specifying variable mode.
//   upol: Assinged value for a variable. The default is LUndef
//   dvar: Indicator whether a variable is to be determined. The default is true.
func (s *Solver) newVar(upol LBool, dvar bool) Var {
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
	s.activity[v] = 0
	s.polarity[v] = 1
	s.userPol[v] = upol
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
	p := Lit(-1)
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
		return false
	} else if len(ps) == 1 {
		s.UncheckedEnqueue(ps[0], nil)
		if confl := s.Propagate(); confl == nil {
			s.ok = true
			return true
		} else {
			s.ok = false
			return false
		}
	} else {
		// set clause
		c := MkClause(ps, true, false)
		s.clauses = append(s.clauses, c)
		s.AttachClause(c)
		return true
	}
}

func (s *Solver) LitValue(p Lit) LBool {
	return s.assigns[p.Var()].Flip(p.Sign())
}

func (s *Solver) AttachClause(c *Clause) {
	s.watches[c.lits[0].Not()] = append(s.watches[c.lits[0].Not()], Watcher{c, c.lits[1]})
	s.watches[c.lits[1].Not()] = append(s.watches[c.lits[0].Not()], Watcher{c, c.lits[0]})
	if c.header.learnt {
		s.numLearntes++
		s.learntsLiterals += uint64(c.Size())
	} else {
		s.numClauses++
		s.clausesLiterals += uint64(c.Size())
	}
}

func (s *Solver) decisionLevel() int {
	return len(s.trailLim)
}

func (s *Solver) UncheckedEnqueue(p Lit, from *Clause) {
	s.assigns[p.Var()] = NewLBool(!p.Sign())
	s.vardata[p.Var()] = VarData{reason: from, level: s.decisionLevel()}
	s.trail = append(s.trail, p)
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

		i := 0
		j := 0
		for i < len(ws) {
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

			// If 0th watch is true, then clause is already satisfied.
			first := c.lits[0]
			w := Watcher{c, first}
			if first != blocker && s.LitValue(first) == LTrue {
				ws[j] = w
				j++
				continue
			}
			i++

			// Look for new watch
			for k := 2; k < c.Size(); k++ {
				if s.LitValue(c.lits[k]) != LFalse {
					c.lits[1], c.lits[k] = c.lits[k], falseLit
					s.watches[c.lits[1].Not()] = append(s.watches[c.lits[1].Not()], w)
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
	s.stats.Propagations += uint64(numProps)
	s.simpDBProps -= int64(numProps)
	return confl
}

func (s *Solver) AddClauseFromCode(codes []int64) {
	lits := make([]Lit, 0, len(codes))
	for _, v := range codes {
		switch {
		case v > 0:
			s.addVar(v - 1) // v starts with 0
			lits = append(lits, MkLit(Var(v-1), false))
		case v < 0:
			s.addVar(-(v + 1)) // v starts with 0
			lits = append(lits, MkLit(Var(-(v+1)), true))
		default:
		}
	}
	s.AddClause(lits...)
}

// add a variable from a general int64
// This function is called from AddClauseFromCode only
func (s *Solver) addVar(v int64) {
	for v >= int64(s.nextVar) {
		s.newVar(LUndef, true)
	}

}
