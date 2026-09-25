package gomisat

import "sort"

// This file holds the parts of the solver interface that a repeated-query
// (incremental) workload needs: the result of a solve has to be observable from
// the outside, and an unsatisfiable answer under assumptions has to report why.
//
// See DESIGN.md: every later phase (tiered clause management, decomposition,
// component caching, cross-query reuse) is validated against the invariants
// these accessors make testable.

// Model returns the satisfying assignment found by the last solve, as a map
// from variable to truth value. The returned map is a copy and may be modified
// by the caller. It is empty unless the last solve returned LTrue.
func (s *Solver) Model() map[Var]LBool {
	m := make(map[Var]LBool, len(s.model))
	for k, v := range s.model {
		m[k] = v
	}
	return m
}

// ModelValue returns the value assigned to v by the last solve, or LUndef if
// the last solve did not return LTrue or v is out of range.
func (s *Solver) ModelValue(v Var) LBool {
	if value, ok := s.model[v]; ok {
		return value
	}
	return LUndef
}

// Conflict returns the raw conflict set produced by the last unsatisfiable
// solve: the negations of the assumption literals that are responsible. It is
// the form used internally; UnsatCore is usually what a caller wants.
// The result is sorted, so it is stable across runs.
func (s *Solver) Conflict() []Lit {
	out := make([]Lit, 0, len(s.conflict))
	for p := range s.conflict {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// UnsatCore returns the subset of the assumptions of the last solve that is
// already sufficient to make the clause set unsatisfiable. It is empty when the
// clause set is unsatisfiable on its own, or when the last solve did not return
// LFalse. The result is sorted, so it is stable across runs.
func (s *Solver) UnsatCore() []Lit {
	out := make([]Lit, 0, len(s.conflict))
	for p := range s.conflict {
		out = append(out, p.Not())
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Assumptions returns the assumption literals used by the last solve.
func (s *Solver) Assumptions() []Lit {
	out := make([]Lit, len(s.assumptions))
	copy(out, s.assumptions)
	return out
}

// ClearAssumptions drops the assumptions of the last solve. Solve does this on
// every call, so it is only needed to release the memory explicitly.
func (s *Solver) ClearAssumptions() {
	s.assumptions = s.assumptions[:0]
}

// NumVars returns the number of variables known to the solver.
func (s *Solver) NumVars() int {
	return int(s.nextVar)
}

// NumClauses returns the number of problem clauses (learnt clauses excluded).
func (s *Solver) NumClauses() uint64 {
	return s.numClauses
}

// NumLearnts returns the current number of learnt clauses.
func (s *Solver) NumLearnts() int {
	return len(s.learnts)
}

// Ok reports whether the clause set is still possibly satisfiable. Once it is
// false the solver answers LFalse to every further solve.
func (s *Solver) Ok() bool {
	return s.ok
}

// SetConfBudget limits the next solve to at most n further conflicts; the solve
// then returns LUndef. A negative n removes the limit.
func (s *Solver) SetConfBudget(n int64) {
	if n < 0 {
		s.conflictBudget = -1
		return
	}
	s.conflictBudget = int64(s.Conflicts) + n
}

// SetPropBudget limits the next solve to at most n further propagations; the
// solve then returns LUndef. A negative n removes the limit.
func (s *Solver) SetPropBudget(n int64) {
	if n < 0 {
		s.propagationBudget = -1
		return
	}
	s.propagationBudget = int64(s.Propagations) + n
}

// Interrupt asks a running solve to stop and return LUndef. It is safe to call
// from another goroutine, which is how a wall-clock timeout is implemented; the
// flag is polled at every decision, so a solve stops promptly.
func (s *Solver) Interrupt() {
	s.asynchInterrupt.Store(true)
}

// ClearInterrupt clears a pending interrupt so the solver can be used again.
func (s *Solver) ClearInterrupt() {
	s.asynchInterrupt.Store(false)
}

// analyzeFinal computes the subset of the assumptions that explains why the
// literal p (the negation of a falsified assumption) cannot hold. The result
// contains p together with the negations of the assumption literals reached
// through the implication graph, i.e. the conflict set in MiniSat's sense.
//
// Precondition: it is called from the assumption placement loop, so every
// decision literal currently on the trail is an assumption. That is what makes
// the result a subset of the assumptions.
func (s *Solver) analyzeFinal(p Lit) map[Lit]struct{} {
	outConflict := map[Lit]struct{}{p: {}}
	if s.decisionLevel() == 0 {
		return outConflict
	}
	seen := map[Var]struct{}{p.Var(): {}}
	for i := len(s.trail) - 1; i >= s.trailLim[0]; i-- {
		x := s.trail[i].Var()
		if _, ok := seen[x]; ok == false {
			continue
		}
		if reason := s.vardata[x].reason; reason == nil {
			// A decision, hence an assumption: it belongs to the core.
			outConflict[s.trail[i].Not()] = struct{}{}
		} else {
			for j := 1; j < len(reason.lits); j++ {
				if s.vardata[reason.lits[j].Var()].level > 0 {
					seen[reason.lits[j].Var()] = struct{}{}
				}
			}
		}
		delete(seen, x)
	}
	return outConflict
}
