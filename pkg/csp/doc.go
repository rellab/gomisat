// Package csp states finite-domain constraint problems and answers them with the
// solver and the model counter of this module.
//
// A model holds Boolean variables and integer variables with finite domains, and
// constraints built from linear sums over the integers compared against zero,
// combined with and, or, implication and negation. The encoding to propositional
// logic is the order encoding: an integer variable with domain
// d[0] < d[1] < ... < d[n-1] becomes n-1 propositional variables, the k-th meaning
// "this variable is at most d[k]", with clauses keeping that chain monotone. Long
// linear sums are broken up with auxiliary integer variables before encoding, which
// is what keeps the encoding of a wide sum from exploding.
//
// # Solving
//
//	m := csp.New()
//	x := m.IntVarRange(0, 10)
//	y := m.IntVarRange(0, 10)
//	m.Add(csp.LeZero(csp.NewSum(map[*csp.IntVar]int{x: 2, y: 3}, -12))) // 2x + 3y <= 12
//	m.Add(csp.GeZero(csp.NewSum(map[*csp.IntVar]int{x: 1, y: -1}, 0)))  // x >= y
//
//	if sol, ok := m.Solve(); ok {
//		fmt.Println(sol.Int(x), sol.Int(y))
//	}
//
// A sum is a map from integer variables to coefficients plus a constant, and
// LeZero, GeZero, EqZero and NeZero compare it against zero. Boolean variables come
// from NewBool and can be mixed into the same constraints.
//
// # Counting
//
// Count returns the number of solutions and WeightedCount the sum over the
// solutions of the product of the weights of their literals, which is how a model
// of a system becomes its probability.
//
// Counting puts a requirement on the encoding that solving does not. An auxiliary
// variable introduced for a subformula has to be *determined* by the original
// variables, or the same solution is counted more than once. The conversion that
// satisfiability normally uses gives each auxiliary only the implication it needs,
// which is smaller and enough to decide satisfiability, and wrong for counting:
// measured here, a formula with 37 solutions came out as 61 models. So the default
// conversion defines every auxiliary in both directions, and Count refuses to
// answer at all when Definitional has been turned off, rather than returning a
// number that looks right.
//
// TestCountPreservation is the test that keeps this honest: it counts the same five
// models under both conversions and requires the definitional one to agree with
// enumeration and the other one to disagree somewhere.
//
// # Provenance
//
// The encoder is derived from github.com/okamumu/gocsp, which follows the order
// encoding of Sugar (Tamura and Banbara). What was added here: the definitional
// conversion above, decoding a solution back into the values of the model's own
// variables, counting and weighted counting, and repairs to two defects -- a
// constraint added without decomposition made the encoder panic, and a constraint
// with no solutions made it exit the process instead of reporting the model
// unsatisfiable.
package csp
