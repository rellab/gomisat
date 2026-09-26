package reliability

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/rellab/gomisat/pkg/gomisat"
)

// countCNF counts the models of a model's CNF with the solver's counter.
func countCNF(t *testing.T, m *Model) *big.Int {
	t.Helper()
	clauses, numVars, _ := m.CNF()
	s := gomisat.NewSolver()
	options := gomisat.DefaultSolverOptions()
	cnf := &gomisat.CNF{NumVars: numVars, NumClauses: len(clauses), Clauses: clauses}
	s.AddCNF(cnf, options)
	got, _ := s.CountModels(options, gomisat.DefaultCountOptions())
	return got
}

func binomial(n, k int) *big.Int {
	return new(big.Int).Binomial(int64(n), int64(k))
}

// TestSeriesAndParallel checks the two textbook structures. A series system works
// in exactly one way, a parallel system in every way but total failure.
func TestSeriesAndParallel(t *testing.T) {
	for n := 1; n <= 6; n++ {
		series := New()
		series.Assert(series.And(series.Events("x", n)...))
		if got := countCNF(t, series); got.Cmp(big.NewInt(1)) != 0 {
			t.Errorf("series of %d: count = %v, want 1", n, got)
		}

		parallel := New()
		parallel.Assert(parallel.Or(parallel.Events("x", n)...))
		want := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(n)), big.NewInt(1))
		if got := countCNF(t, parallel); got.Cmp(want) != 0 {
			t.Errorf("parallel of %d: count = %v, want %v", n, got, want)
		}
	}
}

// TestAtLeastCounts is the important one: the k-out-of-n encoding introduces a
// gate variable per (position, count) pair, and if any of them were not pinned
// down by the events the count would come out too high. The number of working
// states of a k-out-of-n system is the sum of binomials from k to n.
func TestAtLeastCounts(t *testing.T) {
	for n := 1; n <= 8; n++ {
		for k := 0; k <= n+1; k++ {
			m := New()
			m.Assert(m.AtLeast(k, m.Events("x", n)...))
			want := big.NewInt(0)
			for i := k; i <= n; i++ {
				if i < 0 {
					continue
				}
				want.Add(want, binomial(n, i))
			}
			if k > n {
				want = big.NewInt(0)
			}
			if got := countCNF(t, m); got.Cmp(want) != 0 {
				t.Errorf("%d-out-of-%d: count = %v, want %v", k, n, got, want)
			}
		}
	}
}

// TestMultiStateCounts checks the order encoding: a component with m states has
// exactly m consistent assignments, and a system of them has the product.
func TestMultiStateCounts(t *testing.T) {
	m := New()
	m.MultiState("a", 3)
	m.MultiState("b", 4)
	m.MultiState("c", 2)
	want := big.NewInt(3 * 4 * 2)
	if got := countCNF(t, m); got.Cmp(want) != 0 {
		t.Errorf("count = %v, want %v", got, want)
	}
}

// ---- random structure functions against direct evaluation ----

type treeKind int

const (
	kindEvent treeKind = iota
	kindAnd
	kindOr
	kindAtLeast
	kindNot
)

type tree struct {
	kind  treeKind
	event int
	k     int
	kids  []*tree
}

func randomTree(rng *rand.Rand, events, depth int) *tree {
	if depth == 0 || rng.Intn(3) == 0 {
		return &tree{kind: kindEvent, event: rng.Intn(events)}
	}
	n := 2 + rng.Intn(3)
	kids := make([]*tree, 0, n)
	for i := 0; i < n; i++ {
		kids = append(kids, randomTree(rng, events, depth-1))
	}
	switch rng.Intn(4) {
	case 0:
		return &tree{kind: kindAnd, kids: kids}
	case 1:
		return &tree{kind: kindOr, kids: kids}
	case 2:
		return &tree{kind: kindAtLeast, k: 1 + rng.Intn(len(kids)), kids: kids}
	default:
		return &tree{kind: kindNot, kids: kids[:1]}
	}
}

func (t *tree) encode(m *Model, events []Node) Node {
	switch t.kind {
	case kindEvent:
		return events[t.event]
	case kindNot:
		return t.kids[0].encode(m, events).Not()
	}
	kids := make([]Node, 0, len(t.kids))
	for _, kid := range t.kids {
		kids = append(kids, kid.encode(m, events))
	}
	switch t.kind {
	case kindAnd:
		return m.And(kids...)
	case kindOr:
		return m.Or(kids...)
	default:
		return m.AtLeast(t.k, kids...)
	}
}

func (t *tree) eval(assignment []bool) bool {
	switch t.kind {
	case kindEvent:
		return assignment[t.event]
	case kindNot:
		return t.kids[0].eval(assignment) == false
	case kindAnd:
		for _, kid := range t.kids {
			if kid.eval(assignment) == false {
				return false
			}
		}
		return true
	case kindOr:
		for _, kid := range t.kids {
			if kid.eval(assignment) {
				return true
			}
		}
		return false
	default:
		n := 0
		for _, kid := range t.kids {
			if kid.eval(assignment) {
				n++
			}
		}
		return n >= t.k
	}
}

// TestEncodingPreservesCount is the property the whole package rests on: the
// number of models of the CNF must equal the number of event assignments that
// satisfy the structure function, no matter how many auxiliary variables the
// encoding introduced.
func TestEncodingPreservesCount(t *testing.T) {
	rng := rand.New(rand.NewSource(20260926))
	for i := 0; i < 150; i++ {
		events := 3 + rng.Intn(6) // 3..8
		top := randomTree(rng, events, 1+rng.Intn(3))

		m := New()
		nodes := m.Events("x", events)
		m.Assert(top.encode(m, nodes))

		want := big.NewInt(0)
		assignment := make([]bool, events)
		for mask := 0; mask < 1<<events; mask++ {
			for v := 0; v < events; v++ {
				assignment[v] = mask&(1<<v) != 0
			}
			if top.eval(assignment) {
				want.Add(want, big.NewInt(1))
			}
		}

		got := countCNF(t, m)
		if got.Cmp(want) != 0 {
			_, numVars, numEvents := m.CNF()
			t.Fatalf("case %d (%d events, %d variables, %d auxiliaries): count = %v, want %v",
				i, numEvents, numVars, numVars-numEvents, got, want)
		}
	}
}

// ---- weighted counting against the closed forms of reliability theory ----

func approx(t *testing.T, got *big.Float, want float64, what string) {
	t.Helper()
	g, _ := got.Float64()
	if want == 0 {
		if g > 1e-12 {
			t.Errorf("%s = %g, want 0", what, g)
		}
		return
	}
	if rel := (g - want) / want; rel > 1e-9 || rel < -1e-9 {
		t.Errorf("%s = %.15g, want %.15g (relative error %g)", what, g, want, rel)
	}
}

// TestSeriesReliability: a series system works only if every component does, so
// its reliability is the product.
func TestSeriesReliability(t *testing.T) {
	ps := []float64{0.9, 0.95, 0.99, 0.8, 0.7}
	m := New()
	events := m.Events("x", len(ps))
	m.Assert(m.And(events...))
	want := 1.0
	for i, p := range ps {
		m.SetProbability(events[i], p)
		want *= p
	}
	approx(t, m.Reliability(), want, "series reliability")
}

// TestParallelReliability: a parallel system fails only if every component does.
func TestParallelReliability(t *testing.T) {
	ps := []float64{0.5, 0.6, 0.7, 0.2}
	m := New()
	events := m.Events("x", len(ps))
	m.Assert(m.Or(events...))
	fail := 1.0
	for i, p := range ps {
		m.SetProbability(events[i], p)
		fail *= 1 - p
	}
	approx(t, m.Reliability(), 1-fail, "parallel reliability")
}

// TestKOutOfNReliability checks the binomial formula for identical components,
// which is the standard textbook case and exercises the whole gate encoding under
// weights.
func TestKOutOfNReliability(t *testing.T) {
	const p = 0.85
	for n := 1; n <= 10; n++ {
		for k := 1; k <= n; k++ {
			m := New()
			events := m.Events("x", n)
			m.Assert(m.AtLeast(k, events...))
			for _, e := range events {
				m.SetProbability(e, p)
			}
			want := 0.0
			for i := k; i <= n; i++ {
				c, _ := new(big.Float).SetInt(new(big.Int).Binomial(int64(n), int64(i))).Float64()
				want += c * math.Pow(p, float64(i)) * math.Pow(1-p, float64(n-i))
			}
			approx(t, m.Reliability(), want, fmt.Sprintf("%d-out-of-%d reliability", k, n))
		}
	}
}

// TestRandomTreeReliability is the general check: for a random structure function
// with random component probabilities, the weighted count must equal the sum of
// the probabilities of the event assignments that satisfy it.
func TestRandomTreeReliability(t *testing.T) {
	rng := rand.New(rand.NewSource(20260928))
	for i := 0; i < 80; i++ {
		events := 3 + rng.Intn(6)
		top := randomTree(rng, events, 1+rng.Intn(3))

		m := New()
		nodes := m.Events("x", events)
		m.Assert(top.encode(m, nodes))

		ps := make([]float64, events)
		for v := 0; v < events; v++ {
			ps[v] = 0.05 + 0.9*rng.Float64()
			m.SetProbability(nodes[v], ps[v])
		}

		want := 0.0
		assignment := make([]bool, events)
		for mask := 0; mask < 1<<events; mask++ {
			w := 1.0
			for v := 0; v < events; v++ {
				assignment[v] = mask&(1<<v) != 0
				if assignment[v] {
					w *= ps[v]
				} else {
					w *= 1 - ps[v]
				}
			}
			if top.eval(assignment) {
				want += w
			}
		}
		approx(t, m.Reliability(), want, fmt.Sprintf("case %d", i))
	}
}

// TestMultiStateOneHotCounts checks that the one-hot encoding is one-to-one with
// the states, so that counting it counts states.
func TestMultiStateOneHotCounts(t *testing.T) {
	m := New()
	m.MultiStateOneHot("a", 3)
	m.MultiStateOneHot("b", 5)
	clauses, numVars, _ := m.CNF()
	s := gomisat.NewSolver()
	options := gomisat.DefaultSolverOptions()
	s.AddCNF(&gomisat.CNF{NumVars: numVars, NumClauses: len(clauses), Clauses: clauses}, options)
	got, _ := s.CountModels(options, gomisat.DefaultCountOptions())
	if want := big.NewInt(15); got.Cmp(want) != 0 {
		t.Errorf("count = %v, want %v", got, want)
	}
}

// TestMultiStateReliability is the case the project exists for: components with
// several states, a threshold that says which states count as working, and a
// k-out-of-n system over those. With identical components the answer is the
// binomial formula on p = P(state >= threshold).
func TestMultiStateReliability(t *testing.T) {
	// Four states, worst to best.
	q := []float64{0.05, 0.15, 0.3, 0.5}
	const n = 8

	for threshold := 1; threshold < len(q); threshold++ {
		p := 0.0
		for s := threshold; s < len(q); s++ {
			p += q[s]
		}
		for _, k := range []int{1, n / 2, n} {
			m := New()
			working := make([]Node, 0, n)
			for i := 0; i < n; i++ {
				indicators := m.MultiStateOneHot(fmt.Sprintf("c%d", i), len(q))
				m.SetStateProbabilities(indicators, q)
				working = append(working, m.AtLeastState(indicators, threshold))
			}
			m.Assert(m.AtLeast(k, working...))

			want := 0.0
			for i := k; i <= n; i++ {
				c, _ := new(big.Float).SetInt(new(big.Int).Binomial(int64(n), int64(i))).Float64()
				want += c * math.Pow(p, float64(i)) * math.Pow(1-p, float64(n-i))
			}
			approx(t, m.Reliability(), want,
				fmt.Sprintf("%d-out-of-%d at threshold %d (p=%.2f)", k, n, threshold, p))
		}
	}
}
