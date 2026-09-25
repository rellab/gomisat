package gomisat

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTierOf(t *testing.T) {
	options := DefaultSolverOptions()
	tests := []struct {
		lbd  int
		want clauseTier
	}{
		{1, tierCore}, {2, tierCore}, {3, tierMid}, {6, tierMid}, {7, tierLocal}, {50, tierLocal},
	}
	for _, tt := range tests {
		if got := tierOf(tt.lbd, options); got != tt.want {
			t.Errorf("tierOf(%d) = %v, want %v", tt.lbd, got, tt.want)
		}
	}
}

// solveInstance runs one committed instance and hands back the solver so that the
// state of the learnt clause database can be inspected.
func solveInstance(t *testing.T, path string, options *SolverOptions) *Solver {
	t.Helper()
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	cnf, err := ParseDimacsCNF(buf)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	s := NewSolver()
	s.AddCNF(cnf, options)
	s.Solve(options)
	return s
}

// TestLBDIsWellFormed checks the recorded LBD against its definition: it counts
// distinct decision levels, so it is at least 1 and at most the clause length.
func TestLBDIsWellFormed(t *testing.T) {
	options := DefaultSolverOptions()
	path := filepath.Join(satlibDir, "unsat-dimacs-dubois", "dubois100.cnf")
	s := solveInstance(t, path, options)
	if len(s.learnts) == 0 {
		t.Fatal("no learnt clause survived, nothing to check")
	}
	for _, c := range s.learnts {
		if c.lbd < 1 || c.lbd > len(c.lits) {
			t.Fatalf("clause %v has lbd %d, which is outside 1..%d", c, c.lbd, len(c.lits))
		}
		if want := tierOf(c.lbd, options); c.tier != want && c.tier < want {
			// A clause may sit in a better tier than its current LBD suggests
			// only through promotion, never in a worse one.
			t.Fatalf("clause with lbd %d is in tier %v, want at least %v", c.lbd, c.tier, want)
		}
	}
}

// TestReduceDBKeepsProtectedClauses pins the deletion policy: binary, locked and
// core clauses survive a reduction, and the local tier loses half of itself.
func TestReduceDBKeepsProtectedClauses(t *testing.T) {
	options := DefaultSolverOptions()
	path := filepath.Join(satlibDir, "unsat-dimacs-dubois", "dubois100.cnf")
	s := solveInstance(t, path, options)

	protected := make(map[*Clause]bool)
	local := 0
	for _, c := range s.learnts {
		switch {
		case len(c.lits) <= 2 || c.tier != tierLocal || s.Locked(c):
			protected[c] = true
		default:
			local++
		}
	}
	before := len(s.learnts)
	removed := s.reduceDBTiered(options)
	if want := local / 2; removed != want {
		t.Errorf("removed %d clauses, want %d (half of the %d local ones)", removed, want, local)
	}
	if got := len(s.learnts); got != before-removed {
		t.Errorf("database holds %d clauses, want %d", got, before-removed)
	}
	survived := make(map[*Clause]bool, len(s.learnts))
	for _, c := range s.learnts {
		survived[c] = true
	}
	for c := range protected {
		if survived[c] == false {
			t.Fatalf("a protected clause was deleted: %v tier=%v lbd=%d", c, c.tier, c.lbd)
		}
	}
}

// TestSatlibStatusWithoutLBD keeps the activity-only path exercised, since it is
// the reference the LBD work is measured against.
func TestSatlibStatusWithoutLBD(t *testing.T) {
	options := DefaultSolverOptions()
	options.UseLBD = false
	for _, dir := range []string{"file-dimacs-aim", "unsat-dimacs-dubois", "sat-flat125-301"} {
		t.Run(dir, func(t *testing.T) {
			for _, path := range satlibInstances(t, dir) {
				name := filepath.Base(path)
				want := expectedStatus(dir, name)
				clauses := loadCNF(t, path)
				s, _ := solverFor(clauses)
				got := s.Solve(options)
				if got != want {
					t.Errorf("%s: got %v, want %v", name, got, want)
					continue
				}
				if got == LTrue {
					if i := unsatisfiedClause(clauses, s.Model()); i >= 0 {
						t.Errorf("%s: model does not satisfy clause %d", name, i)
					}
				}
			}
		})
	}
}
