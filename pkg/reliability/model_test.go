package reliability

import (
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
