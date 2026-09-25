package gomisat

// Regression suite for the invariants that must keep holding while the solver
// is extended (tiered learnt-clause management, restart policies, vivification,
// component decomposition and caching, cross-query reuse). The point is that a
// later phase which breaks one of them fails here rather than silently
// returning a wrong count.
//
// Four invariants are checked:
//
//  1. status    - the answer matches the known status of every SATLIB instance
//  2. model     - an LTrue answer comes with an assignment satisfying every clause
//  3. oracle    - status and model agree with exhaustive enumeration on small
//                 random instances, with and without assumptions
//  4. reuse     - a reused solver answers exactly as a freshly built one, and
//                 an UNSAT core is a subset of the assumptions sufficient on its own
//
// By default a sample of each SATLIB directory is used so that the suite stays
// fast. Set GOMISAT_FULL=1 to sweep all 2186 instances.

import (
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const satlibDir = "../../testdata/satlib"

// sampleLimit is how many instances per directory are used unless GOMISAT_FULL
// is set.
const sampleLimit = 40

func fullSweep() bool { return os.Getenv("GOMISAT_FULL") == "1" }

// ---------- helpers ----------

// litOf converts a DIMACS literal code to a solver literal, matching the
// numbering used by AddClauseFromCode.
func litOf(code int64) Lit {
	if code > 0 {
		return MkLit(Var(code-1), false)
	}
	return MkLit(Var(-code-1), true)
}

func numVarsOf(clauses [][]int64) int {
	n := 0
	for _, c := range clauses {
		for _, code := range c {
			v := int(code)
			if v < 0 {
				v = -v
			}
			if v > n {
				n = v
			}
		}
	}
	return n
}

func solverFor(clauses [][]int64) (*Solver, *SolverOptions) {
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, c := range clauses {
		s.AddClauseFromCode(c, options)
	}
	return s, options
}

// unsatisfiedClause returns the index of a clause not satisfied by the model,
// or -1 when the model satisfies all of them.
func unsatisfiedClause(clauses [][]int64, model map[Var]LBool) int {
	for i, c := range clauses {
		sat := false
		for _, code := range c {
			p := litOf(code)
			value, ok := model[p.Var()]
			if ok == false {
				continue
			}
			if p.Sign() {
				value = value.Not()
			}
			if value == LTrue {
				sat = true
				break
			}
		}
		if sat == false {
			return i
		}
	}
	return -1
}

// bruteForceSat decides satisfiability by enumerating all assignments of nvars
// variables that agree with the given assumption codes. Only for small nvars.
func bruteForceSat(clauses [][]int64, nvars int, assumps []int64) bool {
	if nvars > 20 {
		panic("bruteForceSat: too many variables")
	}
	for mask := 0; mask < 1<<nvars; mask++ {
		value := func(code int64) bool {
			v := int(code)
			if v < 0 {
				v = -v
			}
			set := mask&(1<<(v-1)) != 0
			if code < 0 {
				return set == false
			}
			return set
		}
		ok := true
		for _, a := range assumps {
			if value(a) == false {
				ok = false
				break
			}
		}
		if ok == false {
			continue
		}
		for _, c := range clauses {
			sat := false
			for _, code := range c {
				if value(code) {
					sat = true
					break
				}
			}
			if sat == false {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// randomCNF builds a random k-CNF with no duplicated or complementary literals
// inside a clause.
func randomCNF(rng *rand.Rand, nvars, nclauses, width int) [][]int64 {
	clauses := make([][]int64, 0, nclauses)
	for i := 0; i < nclauses; i++ {
		used := make(map[int]bool, width)
		c := make([]int64, 0, width)
		for len(c) < width {
			v := rng.Intn(nvars) + 1
			if used[v] {
				continue
			}
			used[v] = true
			if rng.Intn(2) == 0 {
				c = append(c, int64(v))
			} else {
				c = append(c, int64(-v))
			}
		}
		clauses = append(clauses, c)
	}
	return clauses
}

// randomAssumptions picks k distinct variables with random polarity.
func randomAssumptions(rng *rand.Rand, nvars, k int) []int64 {
	if k > nvars {
		k = nvars
	}
	perm := rng.Perm(nvars)[:k]
	out := make([]int64, 0, k)
	for _, v := range perm {
		code := int64(v + 1)
		if rng.Intn(2) == 0 {
			code = -code
		}
		out = append(out, code)
	}
	return out
}

func codesToLits(codes []int64) []Lit {
	out := make([]Lit, 0, len(codes))
	for _, code := range codes {
		out = append(out, litOf(code))
	}
	return out
}

// expectedStatus derives the known answer of a SATLIB instance from the
// directory and file name.
func expectedStatus(dir, name string) LBool {
	switch {
	case strings.HasPrefix(dir, "unsat-"):
		return LFalse
	case strings.HasPrefix(dir, "sat-"):
		return LTrue
	case strings.Contains(name, "-yes"):
		return LTrue
	case strings.Contains(name, "-no"):
		return LFalse
	}
	return LUndef
}

func satlibInstances(t *testing.T, dir string) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(satlibDir, dir, "*.cnf"))
	if err != nil {
		t.Fatalf("glob %s: %v", dir, err)
	}
	if len(paths) == 0 {
		t.Fatalf("no instances found in %s", dir)
	}
	sort.Strings(paths)
	if fullSweep() == false && len(paths) > sampleLimit {
		// Deterministic spread over the directory rather than the first N,
		// which would only ever exercise one family of file names.
		step := len(paths) / sampleLimit
		sampled := make([]string, 0, sampleLimit)
		for i := 0; i < len(paths) && len(sampled) < sampleLimit; i += step {
			sampled = append(sampled, paths[i])
		}
		paths = sampled
	}
	return paths
}

func loadCNF(t *testing.T, path string) [][]int64 {
	t.Helper()
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	clauses, err := ParseDimacs(buf)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return clauses
}

// ---------- 1. status, 2. model ----------

func TestSatlibStatusAndModel(t *testing.T) {
	dirs := []string{
		"file-dimacs-aim",
		"sat-flat125-301",
		"sat-uniform-20-91",
		"unsat-dimacs-dubois",
		"unsat-uniform-50-218",
	}
	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			solved := 0
			paths := satlibInstances(t, dir)
			defer func() { t.Logf("%d/%d instances checked", solved, len(paths)) }()
			for _, path := range paths {
				name := filepath.Base(path)
				want := expectedStatus(dir, name)
				if want == LUndef {
					t.Fatalf("%s: cannot derive the expected status from the name", name)
				}
				clauses := loadCNF(t, path)
				solved++
				s, options := solverFor(clauses)
				got := s.Solve(options)
				if got != want {
					t.Errorf("%s: got %v, want %v", name, got, want)
					continue
				}
				if got == LTrue {
					if i := unsatisfiedClause(clauses, s.Model()); i >= 0 {
						t.Errorf("%s: model does not satisfy clause %d %v", name, i, clauses[i])
					}
				}
			}
		})
	}
}

// ---------- 3. oracle ----------

func TestRandomAgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(20260925))
	const cases = 400
	for i := 0; i < cases; i++ {
		nvars := 4 + rng.Intn(9) // 4..12
		// Straddle the phase transition so that both answers show up.
		nclauses := int(float64(nvars) * (2.0 + 3.0*rng.Float64()))
		clauses := randomCNF(rng, nvars, nclauses, 3)

		s, options := solverFor(clauses)
		got := s.Solve(options)
		want := LFalse
		if bruteForceSat(clauses, nvars, nil) {
			want = LTrue
		}
		if got != want {
			t.Fatalf("case %d (nvars=%d nclauses=%d): got %v, want %v\nclauses=%v",
				i, nvars, nclauses, got, want, clauses)
		}
		if got == LTrue {
			if j := unsatisfiedClause(clauses, s.Model()); j >= 0 {
				t.Fatalf("case %d: model does not satisfy clause %d %v", i, j, clauses[j])
			}
		}
	}
}

func TestRandomWithAssumptionsAgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(20260926))
	const cases = 400
	for i := 0; i < cases; i++ {
		nvars := 4 + rng.Intn(9)
		nclauses := int(float64(nvars) * (1.5 + 2.5*rng.Float64()))
		clauses := randomCNF(rng, nvars, nclauses, 3)
		assumps := randomAssumptions(rng, nvars, 1+rng.Intn(3))

		s, options := solverFor(clauses)
		got := s.SolveWithAssumptions(codesToLits(assumps), options)
		want := LFalse
		if bruteForceSat(clauses, nvars, assumps) {
			want = LTrue
		}
		if got != want {
			t.Fatalf("case %d (nvars=%d): got %v, want %v\nclauses=%v\nassumptions=%v",
				i, nvars, got, want, clauses, assumps)
		}
		if got == LTrue {
			model := s.Model()
			if j := unsatisfiedClause(clauses, model); j >= 0 {
				t.Fatalf("case %d: model does not satisfy clause %d %v", i, j, clauses[j])
			}
			// The model must also respect the assumptions.
			for _, a := range assumps {
				p := litOf(a)
				value := model[p.Var()]
				if p.Sign() {
					value = value.Not()
				}
				if value != LTrue {
					t.Fatalf("case %d: model violates assumption %d", i, a)
				}
			}
		}
	}
}

// ---------- 4. reuse ----------

// TestReusedSolverMatchesFresh runs a sequence of assumption queries against one
// solver and against a solver rebuilt for every query. Any state left behind by
// a previous solve shows up as a disagreement.
func TestReusedSolverMatchesFresh(t *testing.T) {
	rng := rand.New(rand.NewSource(20260927))
	const instances = 60
	const queriesPerInstance = 8
	for i := 0; i < instances; i++ {
		nvars := 6 + rng.Intn(7)
		nclauses := int(float64(nvars) * (1.5 + 2.5*rng.Float64()))
		clauses := randomCNF(rng, nvars, nclauses, 3)

		reused, reusedOptions := solverFor(clauses)
		for q := 0; q < queriesPerInstance; q++ {
			assumps := randomAssumptions(rng, nvars, 1+rng.Intn(3))
			lits := codesToLits(assumps)

			gotReused := reused.SolveWithAssumptions(lits, reusedOptions)

			fresh, freshOptions := solverFor(clauses)
			gotFresh := fresh.SolveWithAssumptions(lits, freshOptions)

			if gotReused != gotFresh {
				t.Fatalf("instance %d query %d: reused solver says %v, fresh solver says %v\nclauses=%v\nassumptions=%v",
					i, q, gotReused, gotFresh, clauses, assumps)
			}
			if gotReused == LTrue {
				if j := unsatisfiedClause(clauses, reused.Model()); j >= 0 {
					t.Fatalf("instance %d query %d: model of the reused solver does not satisfy clause %d %v",
						i, q, j, clauses[j])
				}
			}
			if reused.Ok() == false {
				// The clause set itself turned out to be unsatisfiable; every
				// further query is LFalse for both solvers, so stop here.
				break
			}
		}
	}
}

// TestUnsatCoreIsSubsetAndSufficient checks the two properties a core has to
// have to be usable for repeated queries: it only names assumptions, and it
// already explains the unsatisfiability on its own.
func TestUnsatCoreIsSubsetAndSufficient(t *testing.T) {
	rng := rand.New(rand.NewSource(20260928))
	const cases = 300
	checked := 0
	for i := 0; i < cases; i++ {
		nvars := 5 + rng.Intn(8)
		nclauses := int(float64(nvars) * (2.5 + 2.0*rng.Float64()))
		clauses := randomCNF(rng, nvars, nclauses, 3)
		assumps := randomAssumptions(rng, nvars, 2+rng.Intn(3))
		lits := codesToLits(assumps)

		s, options := solverFor(clauses)
		if s.SolveWithAssumptions(lits, options) != LFalse {
			continue
		}
		core := s.UnsatCore()
		if len(core) == 0 {
			// The clause set is unsatisfiable without any assumption.
			if s.Ok() {
				t.Fatalf("case %d: empty core but the solver is still ok", i)
			}
			continue
		}
		inAssumptions := make(map[Lit]bool, len(lits))
		for _, p := range lits {
			inAssumptions[p] = true
		}
		for _, p := range core {
			if inAssumptions[p] == false {
				t.Fatalf("case %d: core literal %v is not an assumption (assumptions=%v)", i, p, assumps)
			}
		}
		// Solving under the core alone must still be unsatisfiable.
		s2, options2 := solverFor(clauses)
		if got := s2.SolveWithAssumptions(core, options2); got != LFalse {
			t.Fatalf("case %d: core %v is not sufficient, got %v\nclauses=%v", i, core, got, clauses)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no unsatisfiable-under-assumptions case was produced")
	}
	t.Logf("checked %d cores", checked)
}

// TestSolveResetsAssumptions makes sure assumptions do not leak from one query
// into the next.
func TestSolveResetsAssumptions(t *testing.T) {
	// (x1 or x2) and (not x1 or x3): satisfiable, but not with x2 and x3 false.
	clauses := [][]int64{{1, 2}, {-1, 3}}
	s, options := solverFor(clauses)

	if got := s.SolveWithAssumptions(codesToLits([]int64{-2, -3}), options); got != LFalse {
		t.Fatalf("under assumptions: got %v, want F", got)
	}
	core := s.UnsatCore()
	if len(core) != 2 {
		t.Errorf("core = %v, want both assumptions", core)
	}
	if got := s.Solve(options); got != LTrue {
		t.Fatalf("after clearing the assumptions: got %v, want T", got)
	}
	if len(s.Assumptions()) != 0 {
		t.Errorf("assumptions = %v, want none", s.Assumptions())
	}
	if i := unsatisfiedClause(clauses, s.Model()); i >= 0 {
		t.Errorf("model does not satisfy clause %d", i)
	}
}

// TestSubsumes covers the clause helper used by the subsumption and
// vivification work of the next phase.
func TestSubsumes(t *testing.T) {
	mk := func(codes ...int64) *Clause {
		return MkClause(codesToLits(codes), true, false)
	}
	tests := []struct {
		name    string
		c, d    *Clause
		want    Lit
		wantErr bool
	}{
		{"subsumes", mk(1, 2), mk(1, 2, 3), LitUndef, false},
		{"identical", mk(1, 2), mk(1, 2), LitUndef, false},
		{"self subsuming", mk(1, 2), mk(-1, 2, 3), litOf(1), false},
		{"longer cannot subsume", mk(1, 2, 3), mk(1, 2), LitUndef, true},
		{"literal missing", mk(1, 4), mk(1, 2, 3), LitUndef, true},
		{"two negated literals", mk(1, 2), mk(-1, -2, 3), LitUndef, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Subsumes(tt.d)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
