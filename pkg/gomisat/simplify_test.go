package gomisat

import "testing"

// TestSimplifyKeepsUnsatisfiedClauses pins down a defect that was invisible from
// the outside: RemoveSatisfied truncated the clause list to the number of deleted
// clauses instead of to the survivors, so a simplification with nothing to delete
// emptied the list. The solver still answered correctly, because the watch lists
// hold the clauses independently of these lists, but reduceDB could no longer see
// any learnt clause and the database grew without bound.
func TestSimplifyKeepsUnsatisfiedClauses(t *testing.T) {
	// No unit clauses, so nothing is assigned and no clause is satisfied at the
	// root level: simplification has nothing to remove.
	cs := [][]int64{{1, 2, 3}, {-1, 2}, {-2, 3}, {-3, 1}, {1, -2, -3}}
	s, options := solverFor(cs)
	_ = options

	if got := len(s.clauses); got != len(cs) {
		t.Fatalf("before: %d clauses, want %d", got, len(cs))
	}
	if s.Simplify() == false {
		t.Fatal("Simplify reported a conflict")
	}
	if got := len(s.clauses); got != len(cs) {
		t.Errorf("after: %d clauses, want %d", got, len(cs))
	}
}

// TestSimplifyRemovesSatisfiedAndKeepsTrail checks the other half: clauses that
// are satisfied at the root do go away, and the root assignments stay on the
// trail. The trail filter used to keep exactly the released variables instead of
// dropping them, which emptied the trail on every simplification.
func TestSimplifyRemovesSatisfiedAndKeepsTrail(t *testing.T) {
	// The unit clause fixes x1, which satisfies the second clause.
	cs := [][]int64{{1}, {1, 2}, {-2, 3, 4}}
	s, _ := solverFor(cs)

	if s.Simplify() == false {
		t.Fatal("Simplify reported a conflict")
	}
	if len(s.trail) == 0 {
		t.Error("the root level trail was emptied")
	}
	for _, c := range s.clauses {
		if s.Satisfied(c) {
			t.Errorf("a satisfied clause survived: %v", s.arena.String(c))
		}
	}
	if got := len(s.clauses); got != 1 {
		t.Errorf("%d clauses survived, want 1", got)
	}
}

// TestLearntsStayVisibleAfterSimplify is the consequence that mattered: the
// learnt clauses the solver still watches must remain in the list that the
// reduction walks, otherwise the database cannot be bounded.
func TestLearntsStayVisibleAfterSimplify(t *testing.T) {
	options := DefaultSolverOptions()
	s := solveInstance(t, "../../testdata/satlib/unsat-dimacs-dubois/dubois50.cnf", options)
	if len(s.learnts) == 0 {
		t.Skip("no learnt clause survived this instance")
	}
	watched := make(map[CRef]bool)
	for _, ws := range s.watches {
		for _, w := range ws {
			if s.arena.Learnt(w.cref) {
				watched[w.cref] = true
			}
		}
	}
	listed := make(map[CRef]bool, len(s.learnts))
	for _, c := range s.learnts {
		listed[c] = true
	}
	for c := range watched {
		if listed[c] == false {
			t.Fatalf("learnt clause %v is watched but not in s.learnts, so it can never be deleted",
				s.arena.String(c))
		}
	}
	t.Logf("%d learnt clauses, all reachable from the list", len(s.learnts))
}

// TestArenaGarbageIsCollected checks that deleting clauses eventually gives the
// literal store its space back.
func TestArenaGarbageIsCollected(t *testing.T) {
	options := DefaultSolverOptions()
	s := solveInstance(t, "../../testdata/satlib/unsat-dimacs-dubois/dubois100.cnf", options)
	before := len(s.arena.lits)
	for _, c := range append([]CRef(nil), s.learnts...) {
		s.RemoveClause(c)
	}
	s.learnts = s.learnts[:0]
	if s.garbageFraction() <= 0.2 {
		t.Skipf("not enough waste to trigger a collection (%.3f)", s.garbageFraction())
	}
	s.collectGarbage()
	if n := len(s.arena.pending); n != 0 {
		t.Errorf("%d deleted clauses are still pending after a collection", n)
	}
	if got := s.garbageFraction(); got != 0 {
		t.Errorf("garbageFraction = %v after collection, want 0", got)
	}
	if len(s.arena.lits) >= before {
		t.Errorf("literal store did not shrink: %d -> %d", before, len(s.arena.lits))
	}
}

// checkWatchInvariants verifies the structural invariants that the arena has to
// keep. A deleted clause hands its metadata slot back to a free list, so a
// watcher, reason or list entry that outlived its clause would silently start
// referring to a different one; these checks catch that.
func checkWatchInvariants(t *testing.T, s *Solver) {
	t.Helper()

	// Deletion is lazy, so a watch list may still hold entries of deleted clauses
	// until they are swept. Sweep first, then require that none are left: what
	// must never happen is a slot being recycled while a watcher still points at
	// it, and the sweep is what rules that out.
	s.sweepWatches()

	for p, ws := range s.watches {
		for _, w := range ws {
			if w.cref == CRefUndef {
				t.Fatalf("watch list of literal %d holds an undefined clause", p)
			}
			if s.arena.Dead(w.cref) {
				t.Fatalf("watch list of literal %d still holds the deleted clause %v after a sweep",
					p, s.arena.String(w.cref))
			}
			// A clause is watched exactly on the negations of its first two
			// literals.
			lits := s.arena.Lits(w.cref)
			if len(lits) < 2 {
				t.Fatalf("watched clause %v has fewer than two literals", s.arena.String(w.cref))
			}
			if Lit(p) != lits[0].Not() && Lit(p) != lits[1].Not() {
				t.Fatalf("clause %v is watched by literal %v, which is neither %v nor %v",
					s.arena.String(w.cref), Lit(p), lits[0].Not(), lits[1].Not())
			}
		}
	}

	for _, list := range [][]CRef{s.clauses, s.learnts} {
		for _, c := range list {
			if s.arena.Dead(c) {
				t.Fatalf("a clause list holds the deleted clause %v", s.arena.String(c))
			}
		}
	}

	for v := Var(0); v < s.nextVar; v++ {
		reason := s.vardata[v].reason
		if reason == CRefUndef {
			continue
		}
		// A stale reason of an unassigned variable is harmless, and MiniSat keeps
		// them too; only a reason that is actually in use has to be live.
		if s.assigns[v] != LUndef && s.arena.Dead(reason) {
			t.Fatalf("variable %v is assigned with the deleted clause %v as its reason",
				v, s.arena.String(reason))
		}
	}
}

func TestWatchInvariantsAfterSolve(t *testing.T) {
	options := DefaultSolverOptions()
	for _, name := range []string{
		"unsat-dimacs-dubois/dubois100.cnf",
		"sat-flat125-301/flat125-1.cnf",
		"file-dimacs-aim/aim-200-2_0-yes1-1.cnf",
	} {
		t.Run(name, func(t *testing.T) {
			s := solveInstance(t, "../../testdata/satlib/"+name, options)
			checkWatchInvariants(t, s)
		})
	}
}

// TestWatchInvariantsAcrossReuse exercises the same invariants over a sequence of
// queries, where clause deletion and metadata reuse actually interleave.
func TestWatchInvariantsAcrossReuse(t *testing.T) {
	options := DefaultSolverOptions()
	clauses := loadCNF(t, "../../testdata/satlib/sat-flat125-301/flat125-2.cnf")
	s, _ := solverFor(clauses)
	if s.Solve(options) != LTrue {
		t.Fatal("expected the instance to be satisfiable")
	}
	checkWatchInvariants(t, s)

	model := s.Model()
	for i := 0; i < 20; i++ {
		// Assume the negation of what the model says about one variable, which
		// forces real work on every query.
		v := Var(i % s.NumVars())
		p := MkLit(v, model[v] == LTrue)
		s.SolveWithAssumptions([]Lit{p}, options)
		checkWatchInvariants(t, s)
	}
}
