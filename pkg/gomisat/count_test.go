package gomisat

import (
	"math/big"
	"math/rand"
	"testing"
)

// bruteForceCount counts the satisfying assignments of a formula over nvars
// variables by enumeration. It is the oracle the counter is checked against, and
// it is the counting counterpart of the oracle the solver already uses.
func bruteForceCount(clauses [][]int64, nvars int) *big.Int {
	if nvars > 20 {
		panic("bruteForceCount: too many variables")
	}
	total := big.NewInt(0)
	one := big.NewInt(1)
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
		if ok {
			total.Add(total, one)
		}
	}
	return total
}

// countWith builds a solver over nvars variables, counts, and returns the result.
// nvars is passed explicitly so that variables no clause mentions are still
// counted: they are free and double the result.
func countWith(t *testing.T, clauses [][]int64, nvars int, copt *CountOptions) (*big.Int, CountStats) {
	t.Helper()
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, c := range clauses {
		s.AddClauseFromCode(c, options)
	}
	if nvars > 0 {
		s.addVar(int64(nvars-1), options)
	}
	return s.CountModels(options, copt)
}

func TestCountSmallByHand(t *testing.T) {
	tests := []struct {
		name    string
		clauses [][]int64
		nvars   int
		want    int64
	}{
		{"empty formula, one variable", nil, 1, 2},
		{"empty formula, three variables", nil, 3, 8},
		{"single unit", [][]int64{{1}}, 1, 1},
		{"unit and a free variable", [][]int64{{1}}, 2, 2},
		{"contradiction", [][]int64{{1}, {-1}}, 1, 0},
		{"one binary clause", [][]int64{{1, 2}}, 2, 3},
		{"xor of two", [][]int64{{1, 2}, {-1, -2}}, 2, 2},
		// Two independent copies of the binary clause: the count must be the
		// product, 3 * 3, which is what decomposition is for.
		{"two independent clauses", [][]int64{{1, 2}, {3, 4}}, 4, 9},
		{"three independent clauses", [][]int64{{1, 2}, {3, 4}, {5, 6}}, 6, 27},
		// A chain, which does not decompose at the root.
		{"chain", [][]int64{{1, 2}, {-2, 3}, {-3, 4}}, 4, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := countWith(t, tt.clauses, tt.nvars, DefaultCountOptions())
			if got.Cmp(big.NewInt(tt.want)) != 0 {
				t.Errorf("count = %v, want %d", got, tt.want)
			}
			if oracle := bruteForceCount(tt.clauses, tt.nvars); oracle.Cmp(got) != 0 {
				t.Errorf("count = %v, brute force says %v", got, oracle)
			}
		})
	}
}

// TestCountAgainstBruteForce is the counterpart of the solver's oracle test: on
// small random instances the count must equal the enumeration, in every
// combination of decomposition and caching.
func TestCountAgainstBruteForce(t *testing.T) {
	configs := map[string]*CountOptions{
		"plain":                 {},
		"decompose":             {UseDecomposition: true},
		"cache":                 {UseCache: true},
		"decompose+cache":       {UseCache: true, UseDecomposition: true},
		"order":                 {Branching: BranchEliminationOrder},
		"order+decompose+cache": {UseCache: true, UseDecomposition: true, Branching: BranchEliminationOrder},
	}
	for name, copt := range configs {
		t.Run(name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(20260926))
			for i := 0; i < 200; i++ {
				nvars := 3 + rng.Intn(10) // 3..12
				nclauses := rng.Intn(int(float64(nvars) * 3.5))
				clauses := randomCNF(rng, nvars, nclauses, 1+rng.Intn(3))
				want := bruteForceCount(clauses, nvars)
				got, _ := countWith(t, clauses, nvars, copt)
				if got.Cmp(want) != 0 {
					t.Fatalf("case %d (nvars=%d): count = %v, want %v\nclauses=%v",
						i, nvars, got, want, clauses)
				}
			}
		})
	}
}

// TestCountDecompositionMatchesPlain checks the configurations against each other
// on instances too large to enumerate, which catches disagreements the oracle
// cannot reach.
func TestCountDecompositionMatchesPlain(t *testing.T) {
	rng := rand.New(rand.NewSource(20260927))
	for i := 0; i < 40; i++ {
		nvars := 14 + rng.Intn(12) // 14..25
		nclauses := int(float64(nvars) * (1.0 + 2.0*rng.Float64()))
		clauses := randomCNF(rng, nvars, nclauses, 3)

		plain, _ := countWith(t, clauses, nvars, &CountOptions{})
		dec, _ := countWith(t, clauses, nvars, &CountOptions{UseDecomposition: true})
		cached, stats := countWith(t, clauses, nvars, DefaultCountOptions())
		ordered, _ := countWith(t, clauses, nvars, &CountOptions{Branching: BranchEliminationOrder})
		if ordered.Cmp(plain) != 0 {
			t.Fatalf("case %d: elimination-order branching says %v, plain search says %v\nclauses=%v",
				i, ordered, plain, clauses)
		}
		if dec.Cmp(plain) != 0 {
			t.Fatalf("case %d: decomposition says %v, plain search says %v\nclauses=%v", i, dec, plain, clauses)
		}
		if cached.Cmp(plain) != 0 {
			t.Fatalf("case %d: cache says %v, plain search says %v (hits=%d)\nclauses=%v",
				i, cached, plain, stats.CacheHits, clauses)
		}
	}
}

// TestCountIndependentBlocks builds a formula out of k independent copies of the
// same hard little block. The count has to be the k-th power, and decomposition
// has to be what makes it affordable.
func TestCountIndependentBlocks(t *testing.T) {
	const blocks = 12
	const varsPerBlock = 4
	// One block: (a or b) and (not a or c) and (b or not c), 9 models of 16.
	block := [][]int64{{1, 2}, {-1, 3}, {2, -3}}
	blockCount := bruteForceCount(block, varsPerBlock)

	clauses := make([][]int64, 0, blocks*len(block))
	for b := 0; b < blocks; b++ {
		shift := int64(b * varsPerBlock)
		for _, c := range block {
			shifted := make([]int64, 0, len(c))
			for _, code := range c {
				if code > 0 {
					shifted = append(shifted, code+shift)
				} else {
					shifted = append(shifted, code-shift)
				}
			}
			clauses = append(clauses, shifted)
		}
	}
	want := new(big.Int).Exp(blockCount, big.NewInt(blocks), nil)

	got, stats := countWith(t, clauses, blocks*varsPerBlock, DefaultCountOptions())
	if got.Cmp(want) != 0 {
		t.Fatalf("count = %v, want %v (= %v^%d)", got, want, blockCount, blocks)
	}
	t.Logf("%v models, %d decisions, %d components, %d cache hits",
		got, stats.Decisions, stats.Components, stats.CacheHits)
	if stats.Components < blocks {
		t.Errorf("decomposition found %d components, want at least %d", stats.Components, blocks)
	}
}

// TestEliminationOrderIsAPermutation checks the order itself: every variable gets
// exactly one position, so a score can be compared without further care.
func TestEliminationOrderIsAPermutation(t *testing.T) {
	clauses := loadCNF(t, "../../testdata/satlib/sat-uniform-20-91/uf20-01.cnf")
	s, options := solverFor(clauses)
	scores := s.eliminationScores(s.clauses)
	if scores == nil {
		t.Fatal("no scores were produced")
	}
	_ = options
	seen := make(map[int32]bool, len(scores))
	for v, score := range scores {
		if score < 0 || int(score) >= len(scores) {
			t.Fatalf("variable %d has position %d, outside 0..%d", v, score, len(scores)-1)
		}
		if seen[score] {
			t.Fatalf("position %d is used twice", score)
		}
		seen[score] = true
	}
}

// TestBranchingOnCardinality is the case that motivated the elimination order: a
// k-out-of-n constraint has no separable structure under an occurrence-based order
// and the search is exponential, while the elimination order finds the structure.
// The test asserts the gap rather than a time, so it stays meaningful on any
// machine.
func TestBranchingOnCardinality(t *testing.T) {
	clauses := atLeastCNF(18, 9)
	occurrence, occStats := countWith(t, clauses, 18, &CountOptions{
		UseCache: true, UseDecomposition: true, Branching: BranchOccurrence})
	ordered, ordStats := countWith(t, clauses, 18, &CountOptions{
		UseCache: true, UseDecomposition: true, Branching: BranchEliminationOrder})

	if occurrence.Cmp(ordered) != 0 {
		t.Fatalf("the two orders disagree: %v against %v", occurrence, ordered)
	}
	// sum of C(18, i) for i from 9 to 18
	if want := big.NewInt(155382); ordered.Cmp(want) != 0 {
		t.Errorf("count = %v, want %v", ordered, want)
	}
	t.Logf("occurrence %d decisions, elimination order %d decisions",
		occStats.Decisions, ordStats.Decisions)
	if ordStats.Decisions*10 > occStats.Decisions {
		t.Errorf("the elimination order took %d decisions against %d: expected at least a factor of ten",
			ordStats.Decisions, occStats.Decisions)
	}
}

// atLeastCNF encodes "at least k of the first n variables are true" with the same
// recursive construction pkg/reliability uses, as plain DIMACS codes.
func atLeastCNF(n, k int) [][]int64 {
	next := int64(n)
	var clauses [][]int64
	newVar := func() int64 { next++; return next }
	and := func(a, b int64) int64 {
		g := newVar()
		clauses = append(clauses, []int64{-g, a}, []int64{-g, b}, []int64{g, -a, -b})
		return g
	}
	or := func(a, b int64) int64 {
		g := newVar()
		clauses = append(clauses, []int64{g, -a}, []int64{g, -b}, []int64{-g, a, b})
		return g
	}
	trueVar := int64(0)
	getTrue := func() int64 {
		if trueVar == 0 {
			trueVar = newVar()
			clauses = append(clauses, []int64{trueVar})
		}
		return trueVar
	}
	memo := map[[2]int]int64{}
	var rec func(i, j int) int64
	rec = func(i, j int) int64 {
		if j <= 0 {
			return getTrue()
		}
		if n-i < j {
			return -getTrue()
		}
		key := [2]int{i, j}
		if g, ok := memo[key]; ok {
			return g
		}
		g := or(and(int64(i+1), rec(i+1, j-1)), rec(i+1, j))
		memo[key] = g
		return g
	}
	clauses = append(clauses, []int64{rec(0, k)})
	return clauses
}
