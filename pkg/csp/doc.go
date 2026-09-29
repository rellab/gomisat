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
// # What the definitional conversion costs
//
// It is not free, and where it costs anything depends on what the model is made of.
// Measured on graph colouring, where every edge is an inequality and so a
// disjunction, against purely linear models:
//
//	model                    clauses  literals  solving
//	colouring, 20..60 nodes    1.5x      1.7x    1.3x to 1.6x
//	one linear constraint      1.0x      1.0x    1.0x
//
// The integer machinery pays nothing: a comparator is already a literal, so no
// auxiliary is introduced and there is no second direction to state. The cost is
// entirely in the Boolean structure, and it is about half again as many clauses and
// half again as long to solve.
//
// So a caller that only wants solutions and never a count should set Definitional to
// false and take the smaller encoding. The default is the other way round because a
// count from the wrong encoding is wrong quietly, while a solve from either is
// right.
//
// # Where the encoding is large
//
// One linear constraint costs O(n^2 d^2) clauses, n being the number of variables
// and d the size of their domains: measured, 645 clauses for five variables over
// 0..9 and 57 685 for forty of them, and 490 for ten variables over 0..3 against
// 53 608 over 0..39. The decomposition keeps each encoded constraint down to three
// variables, but the auxiliary standing for a partial sum has a domain as wide as
// that sum can range.
//
// The bound the sum is compared against narrows those domains: a partial sum that
// already exceeds the bound cannot take part in a solution whatever the rest of the
// sum does. That is worth up to twenty times fewer clauses when the bound is tight
// relative to a variable's range -- ten variables over 0..9 summing to at most one
// went from 3368 clauses to 151 -- and nothing at all when it is not, which
// includes the k-out-of-n shape over 0/1 variables that reliability models use. It
// is never a loss: where it cannot narrow, the encoding is identical.
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
