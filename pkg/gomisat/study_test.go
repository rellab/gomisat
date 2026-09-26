package gomisat

import (
	"math/big"
	"math/rand"
	"testing"
)

// TestStudyMatchesFreshCounts is the reuse invariant for counting, the counterpart
// of the solver's reused-against-fresh test: a query answered with a cache carried
// over from earlier queries must give exactly the answer a fresh counter gives.
func TestStudyMatchesFreshCounts(t *testing.T) {
	rng := rand.New(rand.NewSource(20260929))
	for i := 0; i < 60; i++ {
		nvars := 6 + rng.Intn(8)
		nclauses := int(float64(nvars) * (1.0 + 2.0*rng.Float64()))
		clauses := randomCNF(rng, nvars, nclauses, 3)

		// One solver for the whole sequence, with a cache that survives.
		s := NewSolver()
		options := DefaultSolverOptions()
		for _, c := range clauses {
			s.AddClauseFromCode(c, options)
		}
		weights := NewWeights(s.NumVars())
		for v := 0; v < s.NumVars(); v++ {
			weights.SetProbability(Var(v), 0.1+0.8*rng.Float64())
		}
		study := NewStudy(s, options, DefaultCountOptions(), weights)

		for q := 0; q < 8; q++ {
			assumptions := codesToLits(randomAssumptions(rng, nvars, 1+rng.Intn(3)))

			reused, _ := study.Count(assumptions...)

			// The same query on a solver built from scratch, with the assumptions
			// as unit clauses.
			fresh := NewSolver()
			freshOptions := DefaultSolverOptions()
			for _, c := range clauses {
				fresh.AddClauseFromCode(c, freshOptions)
			}
			for _, p := range assumptions {
				code := int64(p.Var()) + 1
				if p.Sign() {
					code = -code
				}
				fresh.AddClauseFromCode([]int64{code}, freshOptions)
			}
			freshWeights := NewWeights(fresh.NumVars())
			for v := 0; v < fresh.NumVars(); v++ {
				freshWeights.Set(Var(v), weights.Of(MkLit(Var(v), false)), weights.Of(MkLit(Var(v), true)))
			}
			want, _ := fresh.WeightedCount(freshOptions, DefaultCountOptions(), freshWeights)

			if nearlyEqual(reused, want, 200) == false {
				t.Fatalf("instance %d query %d: the study says %v, a fresh counter says %v\nclauses=%v assumptions=%v",
					i, q, reused, want, clauses, assumptions)
			}
		}
	}
}

// TestStudyCacheIsReused checks that carrying the cache over actually saves work,
// on a sequence of queries over the same formula.
func TestStudyCacheIsReused(t *testing.T) {
	clauses := atLeastCNF(16, 8)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, c := range clauses {
		s.AddClauseFromCode(c, options)
	}
	study := NewStudy(s, options, DefaultCountOptions(), nil)

	// Condition on each component in turn having failed, then on each working.
	var first, later CountStats
	for q := 0; q < 16; q++ {
		_, stats := study.Count(MkLit(Var(q), true))
		if q == 0 {
			first = stats
		} else {
			later.Decisions += stats.Decisions
			later.CacheHits += stats.CacheHits
			later.CacheMiss += stats.CacheMiss
		}
	}
	t.Logf("first query: %d decisions, %d misses; the next fifteen: %d decisions, %d hits, %d misses",
		first.Decisions, first.CacheMiss, later.Decisions, later.CacheHits, later.CacheMiss)
	if later.CacheHits == 0 {
		t.Error("no cache entry was reused across queries")
	}
	// The later queries together should not cost fifteen times the first one.
	if later.Decisions > 15*first.Decisions {
		t.Errorf("fifteen further queries took %d decisions against %d for the first: no reuse",
			later.Decisions, first.Decisions)
	}
}

// TestStudyContradictoryAssumptions checks the degenerate query.
func TestStudyContradictoryAssumptions(t *testing.T) {
	s := NewSolver()
	options := DefaultSolverOptions()
	s.AddClauseFromCode([]int64{1, 2}, options)
	study := NewStudy(s, options, DefaultCountOptions(), nil)

	if got, _ := study.Count(litOf(-1), litOf(-2)); got.Sign() != 0 {
		t.Errorf("count under contradictory assumptions = %v, want 0", got)
	}
	// The study must still work afterwards: 3 models of (x1 or x2).
	if got, _ := study.Count(); got.Cmp(big.NewFloat(3)) != 0 {
		t.Errorf("count without assumptions = %v, want 3", got)
	}
}

// nearlyEqual compares two weighted counts to within relBits of relative
// precision.
//
// Bit-identical equality is the wrong requirement here. A weighted count is a sum
// of products of floating-point weights, and the cache changes the order in which
// those products are formed, so the last bits of the mantissa differ between a
// cached and an uncached evaluation. The values agree to about 77 decimal digits at
// the default precision; anything a reliability figure is used for needs far fewer.
// Exact reproducibility would mean exact rational arithmetic, which is noted in
// DESIGN.md as an option and is not free.
func nearlyEqual(a, b *big.Float, relBits uint) bool {
	if a.Sign() == 0 || b.Sign() == 0 {
		return a.Cmp(b) == 0
	}
	diff := new(big.Float).Sub(a, b)
	diff.Abs(diff)
	scale := new(big.Float).Abs(a)
	tolerance := new(big.Float).SetMantExp(scale, scale.MantExp(nil)-int(relBits))
	return diff.Cmp(tolerance) <= 0
}
