package gomisat

import (
	"math"
	"math/big"
	"math/rand"
	"testing"
)

// bruteForceWeight sums the weight of every satisfying assignment, the weight of
// an assignment being the product of its literal weights.
func bruteForceWeight(clauses [][]int64, nvars int, pTrue []float64) float64 {
	total := 0.0
	for mask := 0; mask < 1<<nvars; mask++ {
		ok := true
		for _, c := range clauses {
			sat := false
			for _, code := range c {
				v := int(code)
				if v < 0 {
					v = -v
				}
				set := mask&(1<<(v-1)) != 0
				if (code > 0) == set {
					sat = true
					break
				}
			}
			if sat == false {
				ok = false
				break
			}
		}
		if ok == false {
			continue
		}
		w := 1.0
		for v := 0; v < nvars; v++ {
			if mask&(1<<v) != 0 {
				w *= pTrue[v]
			} else {
				w *= 1 - pTrue[v]
			}
		}
		total += w
	}
	return total
}

func weightedWith(t *testing.T, clauses [][]int64, nvars int, pTrue []float64, copt *CountOptions) *big.Float {
	t.Helper()
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, c := range clauses {
		s.AddClauseFromCode(c, options)
	}
	if nvars > 0 {
		s.addVar(int64(nvars-1), options)
	}
	var w *Weights
	if pTrue != nil {
		w = NewWeights(s.NumVars())
		for v, p := range pTrue {
			w.SetProbability(Var(v), p)
		}
	}
	got, _ := s.WeightedCount(options, copt, w)
	return got
}

func closeEnough(got *big.Float, want float64) bool {
	g, _ := got.Float64()
	if want == 0 {
		return math.Abs(g) < 1e-12
	}
	return math.Abs(g-want)/math.Abs(want) < 1e-9
}

// TestWeightedUnitWeightsMatchesCount cross-checks the two counters: with weight
// one on every literal the weighted count is the model count, and the two are
// separate implementations of the same recursion.
func TestWeightedUnitWeightsMatchesCount(t *testing.T) {
	rng := rand.New(rand.NewSource(20260926))
	for i := 0; i < 150; i++ {
		nvars := 3 + rng.Intn(10)
		nclauses := rng.Intn(int(float64(nvars) * 3.5))
		clauses := randomCNF(rng, nvars, nclauses, 1+rng.Intn(3))

		exact, _ := countWith(t, clauses, nvars, DefaultCountOptions())
		weighted := weightedWith(t, clauses, nvars, nil, DefaultCountOptions())

		want := new(big.Float).SetInt(exact)
		if weighted.Cmp(want) != 0 {
			t.Fatalf("case %d: weighted count with unit weights = %v, model count = %v\nclauses=%v",
				i, weighted, exact, clauses)
		}
	}
}

// TestWeightedAgainstBruteForce checks the weighted count against summing over
// every assignment, in each configuration.
func TestWeightedAgainstBruteForce(t *testing.T) {
	configs := map[string]*CountOptions{
		"plain":                 {},
		"decompose+cache":       {UseCache: true, UseDecomposition: true},
		"order+decompose+cache": {UseCache: true, UseDecomposition: true, Branching: BranchEliminationOrder},
	}
	for name, copt := range configs {
		t.Run(name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(20260927))
			for i := 0; i < 150; i++ {
				nvars := 3 + rng.Intn(9)
				nclauses := rng.Intn(int(float64(nvars) * 3.0))
				clauses := randomCNF(rng, nvars, nclauses, 1+rng.Intn(3))
				pTrue := make([]float64, nvars)
				for v := range pTrue {
					pTrue[v] = 0.05 + 0.9*rng.Float64()
				}
				want := bruteForceWeight(clauses, nvars, pTrue)
				got := weightedWith(t, clauses, nvars, pTrue, copt)
				if closeEnough(got, want) == false {
					t.Fatalf("case %d (nvars=%d): weighted count = %v, brute force = %g\nclauses=%v",
						i, nvars, got, want, clauses)
				}
			}
		})
	}
}

// TestWeightedSurvivesUnderflow is why the accumulator is not a float64: the
// probability of a long series of components is far below the smallest double.
func TestWeightedSurvivesUnderflow(t *testing.T) {
	const n = 1200
	const p = 0.5
	clauses := make([][]int64, 0, n)
	for v := 1; v <= n; v++ {
		clauses = append(clauses, []int64{int64(v)}) // every component must work
	}
	pTrue := make([]float64, n)
	for v := range pTrue {
		pTrue[v] = p
	}
	got := weightedWith(t, clauses, n, pTrue, DefaultCountOptions())

	// 0.5^1200, far below the smallest positive double.
	want := new(big.Float).SetPrec(256).SetInt64(1)
	want.Quo(want, new(big.Float).SetPrec(256).SetInt(new(big.Int).Lsh(big.NewInt(1), n)))
	if got.Cmp(want) != 0 {
		t.Errorf("weighted count = %v, want %v", got, want)
	}
	if f, _ := got.Float64(); f != 0 {
		t.Errorf("as a float64 this is %g, so the test is not exercising underflow any more", f)
	}
}

// TestHashedCacheAgreesWithExact checks the 128-bit component digest against the
// full component description. A collision would show up as a wrong count, and this
// is the only thing standing between the default cache and that risk.
func TestHashedCacheAgreesWithExact(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	for i := 0; i < 200; i++ {
		nvars := 4 + rng.Intn(12)
		nclauses := int(float64(nvars) * (1.0 + 2.0*rng.Float64()))
		clauses := randomCNF(rng, nvars, nclauses, 3)
		pTrue := make([]float64, nvars)
		for v := range pTrue {
			pTrue[v] = 0.05 + 0.9*rng.Float64()
		}

		hashed := weightedWith(t, clauses, nvars, pTrue, &CountOptions{
			UseCache: true, UseDecomposition: true, Branching: BranchEliminationOrder})
		exact := weightedWith(t, clauses, nvars, pTrue, &CountOptions{
			UseCache: true, UseDecomposition: true, Branching: BranchEliminationOrder, ExactCache: true})

		if hashed.Cmp(exact) != 0 {
			t.Fatalf("case %d: the hashed cache says %v, the exact cache says %v\nclauses=%v",
				i, hashed, exact, clauses)
		}
	}
}

// TestHashComponentDistinguishes checks that the digest reacts to the things that
// distinguish one component from another, including moving an element between the
// two lists.
func TestHashComponentDistinguishes(t *testing.T) {
	base := hashComponent([]Var{1, 2, 3}, []CRef{10, 11})
	cases := map[string]cacheKey{
		"a variable added":      hashComponent([]Var{1, 2, 3, 4}, []CRef{10, 11}),
		"a variable changed":    hashComponent([]Var{1, 2, 4}, []CRef{10, 11}),
		"variables reordered":   hashComponent([]Var{1, 3, 2}, []CRef{10, 11}),
		"a clause added":        hashComponent([]Var{1, 2, 3}, []CRef{10, 11, 12}),
		"a clause changed":      hashComponent([]Var{1, 2, 3}, []CRef{10, 12}),
		"one fewer of each":     hashComponent([]Var{1, 2}, []CRef{10}),
		"lists of equal values": hashComponent([]Var{1, 2}, []CRef{3, 10, 11}),
	}
	for name, other := range cases {
		if other == base {
			t.Errorf("%s: the digest did not change", name)
		}
	}
}
