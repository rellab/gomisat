package csp

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/rellab/gomisat/pkg/gomisat"
)

// ---- oracles ----

// enumerate walks every assignment of the given integer domains and counts the
// ones the predicate accepts. It is the oracle both the solver and the counter are
// checked against.
func enumerate(domains [][]int, ok func(assignment []int) bool) int {
	n := 0
	assignment := make([]int, len(domains))
	var walk func(i int)
	walk = func(i int) {
		if i == len(domains) {
			if ok(assignment) {
				n++
			}
			return
		}
		for _, v := range domains[i] {
			assignment[i] = v
			walk(i + 1)
		}
	}
	walk(0)
	return n
}

func mustCount(t *testing.T, m *Model) *big.Int {
	t.Helper()
	got, ok := m.Count()
	if ok == false {
		t.Fatal("Count refused: the model is not using the definitional conversion")
	}
	return got
}

// ---- counting ----

// TestCountPreservation is the property that decides whether this package can be
// used for counting at all: the CNF must have exactly as many models as the model
// has solutions, however many auxiliary variables the encoding introduced.
//
// The polarity-optimised conversion fails it, and not on the first case one tries:
// two overlapping branches come out right and three do not. Both conversions are
// run here so that the difference stays visible.
func TestCountPreservation(t *testing.T) {
	// Every expected value is enumerated rather than worked out by hand. The first
	// version of this test had one of them wrong and reported an encoding defect as
	// correct because of it.
	bools := func(n int, ok func(b []bool) bool) int {
		count := 0
		b := make([]bool, n)
		for mask := 0; mask < 1<<n; mask++ {
			for i := range b {
				b[i] = mask&(1<<i) != 0
			}
			if ok(b) {
				count++
			}
		}
		return count
	}

	cases := []struct {
		name  string
		build func(m *Model)
		want  int
	}{
		{
			name: "a <-> b, as exclusive branches",
			build: func(m *Model) {
				a, b := m.NewBool(), m.NewBool()
				m.Add(Or(And(a, b), And(a.Not(), b.Not())))
			},
			want: bools(2, func(b []bool) bool { return b[0] == b[1] }),
		},
		{
			name: "(a&b) or (c&d), branches overlap",
			build: func(m *Model) {
				p, q, r, s := m.NewBool(), m.NewBool(), m.NewBool(), m.NewBool()
				m.Add(Or(And(p, q), And(r, s)))
			},
			want: bools(4, func(b []bool) bool { return (b[0] && b[1]) || (b[2] && b[3]) }),
		},
		{
			name: "three overlapping ands",
			build: func(m *Model) {
				v := make([]*BoolVar, 6)
				for i := range v {
					v[i] = m.NewBool()
				}
				m.Add(Or(And(v[0], v[1]), And(v[2], v[3]), And(v[4], v[5])))
			},
			want: bools(6, func(b []bool) bool {
				return (b[0] && b[1]) || (b[2] && b[3]) || (b[4] && b[5])
			}),
		},
		{
			name: "x in 0..5, x <= 5",
			build: func(m *Model) {
				x := m.IntVarRange(0, 5)
				m.Add(LeZero(NewSum(map[*IntVar]int{x: 1}, -5)))
			},
			want: enumerate([][]int{{0, 1, 2, 3, 4, 5}}, func(a []int) bool { return a[0] <= 5 }),
		},
		{
			name: "y + z <= 3 over 0..3",
			build: func(m *Model) {
				y, z := m.IntVarRange(0, 3), m.IntVarRange(0, 3)
				m.Add(LeZero(NewSum(map[*IntVar]int{y: 1, z: 1}, -3)))
			},
			want: enumerate([][]int{{0, 1, 2, 3}, {0, 1, 2, 3}},
				func(a []int) bool { return a[0]+a[1] <= 3 }),
		},
	}

	inflated := 0
	for _, tc := range cases {
		definitional := New()
		tc.build(definitional)
		if got := mustCount(t, definitional); got.Cmp(big.NewInt(int64(tc.want))) != 0 {
			t.Errorf("definitional, %s: %v models, %d solutions", tc.name, got, tc.want)
		}

		// The polarity-optimised conversion is counted directly, since Count
		// refuses it on purpose.
		optimised := New()
		optimised.Definitional = false
		tc.build(optimised)
		got := countBuilt(t, optimised)
		if got.Cmp(big.NewInt(int64(tc.want))) != 0 {
			t.Logf("polarity-optimised, %-34s %v models against %d solutions", tc.name, got, tc.want)
			inflated++
		}
	}
	if inflated == 0 {
		t.Error("the polarity-optimised conversion counted everything correctly, " +
			"so this test no longer shows why the definitional one is the default")
	}
}

// TestCountAgainstEnumeration checks the count of random linear models against
// walking their domains.
func TestCountAgainstEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(20260929))
	for i := 0; i < 40; i++ {
		nvars := 2 + rng.Intn(3)
		ub := 2 + rng.Intn(3)
		m := New()
		vars := make([]*IntVar, nvars)
		domains := make([][]int, nvars)
		for j := range vars {
			vars[j] = m.IntVarRange(0, ub)
			domains[j] = vars[j].Domain()
		}
		coef := make([]int, nvars)
		sum := make(map[*IntVar]int, nvars)
		for j := range coef {
			coef[j] = 1 + rng.Intn(3)
			sum[vars[j]] = coef[j]
		}
		rhs := rng.Intn(nvars * ub * 2)
		m.Add(LeZero(NewSum(sum, -rhs)))

		want := enumerate(domains, func(a []int) bool {
			total := 0
			for j, v := range a {
				total += coef[j] * v
			}
			return total <= rhs
		})
		if got := mustCount(t, m); got.Cmp(big.NewInt(int64(want))) != 0 {
			t.Fatalf("case %d: %v models, %d solutions (coef=%v rhs=%d ub=%d)",
				i, got, want, coef, rhs, ub)
		}
	}
}

// ---- solving ----

// TestSolveDecodesValues checks the other half: a solution has to come back as
// values of the model's own variables, and those values have to satisfy the
// constraints. Decoding was not implemented in the package this came from.
func TestSolveDecodesValues(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	solved, unsolved := 0, 0
	for i := 0; i < 60; i++ {
		nvars := 2 + rng.Intn(3)
		ub := 2 + rng.Intn(4)
		m := New()
		vars := make([]*IntVar, nvars)
		domains := make([][]int, nvars)
		for j := range vars {
			vars[j] = m.IntVarRange(0, ub)
			domains[j] = vars[j].Domain()
		}
		coef := make([]int, nvars)
		sum := make(map[*IntVar]int, nvars)
		for j := range coef {
			coef[j] = 1 + rng.Intn(3)
			sum[vars[j]] = coef[j]
		}
		// A negative bound now and then, so that both outcomes are covered.
		rhs := rng.Intn(nvars*ub) - 2
		m.Add(LeZero(NewSum(sum, -rhs)))

		want := enumerate(domains, func(a []int) bool {
			total := 0
			for j, v := range a {
				total += coef[j] * v
			}
			return total <= rhs
		})

		sol, ok := m.Solve()
		if ok == false {
			unsolved++
			if want != 0 {
				t.Fatalf("case %d: reported no solution, enumeration found %d", i, want)
			}
			continue
		}
		solved++
		if want == 0 {
			t.Fatalf("case %d: reported a solution, enumeration found none", i)
		}
		total := 0
		for j, v := range vars {
			value := sol.Int(v)
			if value < 0 || value > ub {
				t.Fatalf("case %d: variable %d decoded to %d, outside 0..%d", i, j, value, ub)
			}
			total += coef[j] * value
		}
		if total > rhs {
			t.Fatalf("case %d: the decoded solution has sum %d, which breaks <= %d", i, total, rhs)
		}
	}
	t.Logf("%d solved, %d without a solution", solved, unsolved)
	if solved == 0 || unsolved == 0 {
		t.Error("the cases did not cover both outcomes")
	}
}

// TestSolveBooleans covers Boolean variables and their decoding.
func TestSolveBooleans(t *testing.T) {
	m := New()
	a, b, c := m.NewBool(), m.NewBool(), m.NewBool()
	m.Add(And(Or(a, b), Or(a.Not(), c), Or(b.Not(), c.Not())))

	sol, ok := m.Solve()
	if ok == false {
		t.Fatal("no solution, but (a|b) & (~a|c) & (~b|~c) has several")
	}
	av, bv, cv := sol.Bool(a), sol.Bool(b), sol.Bool(c)
	if !((av || bv) && (!av || cv) && (!bv || !cv)) {
		t.Errorf("the decoded assignment a=%v b=%v c=%v does not satisfy the constraints", av, bv, cv)
	}
	want := 0
	for mask := 0; mask < 8; mask++ {
		x, y, z := mask&1 != 0, mask&2 != 0, mask&4 != 0
		if (x || y) && (!x || z) && (!y || !z) {
			want++
		}
	}
	if got := mustCount(t, m); got.Cmp(big.NewInt(int64(want))) != 0 {
		t.Errorf("count = %v, want %d", got, want)
	}
}

// TestUnsatisfiable checks the empty case end to end.
func TestUnsatisfiable(t *testing.T) {
	m := New()
	x := m.IntVarRange(0, 3)
	m.Add(LeZero(NewSum(map[*IntVar]int{x: 1}, -1))) // x <= 1
	m.Add(LeZero(NewSum(map[*IntVar]int{x: -1}, 2))) // x >= 2
	if _, ok := m.Solve(); ok {
		t.Error("reported a solution for x <= 1 and x >= 2")
	}
	if got := mustCount(t, m); got.Sign() != 0 {
		t.Errorf("count = %v, want 0", got)
	}
}

// countBuilt counts the models of a model's CNF whatever conversion it used. The
// exported Count refuses the polarity-optimised one; this is how the test shows
// what that refusal is protecting against.
func countBuilt(t *testing.T, m *Model) *big.Int {
	t.Helper()
	clauses, numVars := m.Build()
	s := gomisat.NewSolver()
	options := gomisat.DefaultSolverOptions()
	s.AddCNF(&gomisat.CNF{NumVars: numVars, NumClauses: len(clauses), Clauses: clauses}, options)
	count, _ := s.CountModels(options, gomisat.DefaultCountOptions())
	return count
}

// TestTightBoundsPreserveCountsAndSolutions covers the domain narrowing of
// decompSum. A tight bound lets most of an auxiliary domain be thrown away, and
// what must not change is the set of solutions or their number.
func TestTightBoundsPreserveCountsAndSolutions(t *testing.T) {
	for _, n := range []int{3, 4, 5} {
		for _, ub := range []int{1, 2, 3} {
			for rhs := -1; rhs <= n*ub+1; rhs++ {
				m := New()
				vars := make([]*IntVar, n)
				domains := make([][]int, n)
				sum := make(map[*IntVar]int, n)
				for i := range vars {
					vars[i] = m.IntVarRange(0, ub)
					domains[i] = vars[i].Domain()
					sum[vars[i]] = 1
				}
				m.Add(LeZero(NewSum(sum, -rhs)))

				want := enumerate(domains, func(a []int) bool {
					total := 0
					for _, v := range a {
						total += v
					}
					return total <= rhs
				})
				if got := mustCount(t, m); got.Cmp(big.NewInt(int64(want))) != 0 {
					t.Fatalf("n=%d ub=%d rhs=%d: %v models, %d solutions", n, ub, rhs, got, want)
				}

				sol, ok := m.Solve()
				if ok != (want > 0) {
					t.Fatalf("n=%d ub=%d rhs=%d: Solve says %v, enumeration found %d solutions",
						n, ub, rhs, ok, want)
				}
				if ok {
					total := 0
					for _, v := range vars {
						total += sol.Int(v)
					}
					if total > rhs {
						t.Fatalf("n=%d ub=%d rhs=%d: the decoded solution sums to %d", n, ub, rhs, total)
					}
				}
			}
		}
	}
}

// TestNarrowingWithMixedSigns covers the same for bounds in the other direction and
// for coefficients that are not all positive, where the reasoning about which end of
// a domain can be dropped is easy to get backwards.
func TestNarrowingWithMixedSigns(t *testing.T) {
	for _, coef := range [][]int{{1, 1, 1, 1}, {1, -1, 2, -2}, {-1, -1, -1, -1}, {3, -2, 1, -1}} {
		for rhs := -6; rhs <= 6; rhs += 2 {
			for _, ge := range []bool{false, true} {
				m := New()
				n := len(coef)
				vars := make([]*IntVar, n)
				domains := make([][]int, n)
				sum := make(map[*IntVar]int, n)
				for i := range vars {
					vars[i] = m.IntVarRange(0, 2)
					domains[i] = vars[i].Domain()
					sum[vars[i]] = coef[i]
				}
				if ge {
					m.Add(GeZero(NewSum(sum, -rhs)))
				} else {
					m.Add(LeZero(NewSum(sum, -rhs)))
				}

				want := enumerate(domains, func(a []int) bool {
					total := 0
					for i, v := range a {
						total += coef[i] * v
					}
					if ge {
						return total >= rhs
					}
					return total <= rhs
				})
				if got := mustCount(t, m); got.Cmp(big.NewInt(int64(want))) != 0 {
					t.Fatalf("coef=%v rhs=%d ge=%v: %v models, %d solutions", coef, rhs, ge, got, want)
				}
			}
		}
	}
}
