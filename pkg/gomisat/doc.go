// Package gomisat is a CDCL satisfiability solver and an exact model counter.
//
// The solver decides whether a formula in conjunctive normal form is satisfiable,
// reports a satisfying assignment, answers a sequence of queries under assumptions
// and reports why an answer was negative. The counter says how many assignments
// satisfy a formula, or, with weights on the literals, what they are worth --
// which is how a model of a system becomes its probability.
//
// It began as a port of MiniSat 2.2 and has since been taken through the pieces
// that a counting engine needs. What each of those pieces is worth was measured
// rather than assumed, and DESIGN.md in the repository records every measurement,
// including the several predictions that turned out wrong.
//
// # Solving
//
//	s := gomisat.NewSolver()
//	options := gomisat.DefaultSolverOptions()
//	cnf, err := gomisat.ParseDimacsCNF(bytes)
//	if err != nil {
//		return err
//	}
//	s.AddCNF(cnf, options)
//
//	switch s.Solve(options) {
//	case gomisat.LTrue:
//		model := s.Model() // map[Var]LBool
//	case gomisat.LFalse:
//		// unsatisfiable
//	}
//
// Variables are numbered from zero and literals are made with [MkLit]; a DIMACS
// code v becomes MkLit(Var(v-1), false) and -v becomes MkLit(Var(v-1), true).
// [Solver.AddClauseFromCode] takes DIMACS codes directly.
//
// # Repeated queries
//
// A solver may be reused. [Solver.SolveWithAssumptions] fixes literals for one
// query and replaces them on the next, and [Solver.UnsatCore] reports which of them
// were responsible for a negative answer:
//
//	for _, assumptions := range queries {
//		if s.SolveWithAssumptions(assumptions, options) == gomisat.LFalse {
//			core := s.UnsatCore() // a subset of assumptions, sufficient on its own
//		}
//	}
//
// That a reused solver answers exactly as a fresh one is a tested invariant, not an
// aspiration: it was not true when this work started.
//
// # Counting
//
// [Solver.CountModels] counts the satisfying assignments and
// [Solver.WeightedCount] sums the products of their literal weights. Both branch,
// propagate, split what is left of the formula into independent components, count
// each and multiply, and memoise components so that a sub-problem met twice is
// solved once.
//
//	count, stats := s.CountModels(options, gomisat.DefaultCountOptions())
//
//	weights := gomisat.NewWeights(s.NumVars())
//	weights.SetProbability(gomisat.Var(0), 0.9)
//	value, _ := s.WeightedCount(options, gomisat.DefaultCountOptions(), weights)
//
// Two things to know before counting anything.
//
// The counter propagates with the problem clauses alone and must not be used on a
// solver that has learnt any. A learnt clause may span two components, and then
// assigning inside one of them propagates into the other, which is what a product
// of independent counts cannot survive. Count on a fresh solver, or solve on a copy.
//
// The auxiliary variables of an encoding have to be *determined* by the variables
// being counted, or the same solution is counted more than once. The two model
// builders in this module, [github.com/rellab/gomisat/pkg/reliability] and
// [github.com/rellab/gomisat/pkg/csp], both take care of it; an encoding written by
// hand may not, and the failure is silent. Checking a small case against enumeration
// costs little and catches it.
//
// # A sequence of counting queries
//
// [NewStudy] answers a sequence of weighted counting queries over one formula and
// keeps the component cache between them, which is worth a steady factor of about
// four on the sequences measured. Queries are expressed as assumptions rather than
// as edits to the formula, since that is what lets a cached component stay valid.
//
//	study := gomisat.NewStudy(s, options, gomisat.DefaultCountOptions(), weights)
//	for _, lit := range conditions {
//		value, _ := study.Count(lit)
//		_ = value
//	}
//
// The weights must not change while the cache lives, because a cached value is a
// weighted count; use [Study.ClearCache] if they do.
//
// # Options
//
// [DefaultSolverOptions] and [DefaultCountOptions] are the settings that measured
// best, and the settings they beat are still selectable so that the measurements
// can be repeated. In the solver: learnt clauses are managed by literal block
// distance in tiers, the database is reduced on a conflict-driven schedule, and
// restarts follow a fast and a slow average of clause quality with Glucose's
// blocking rule. In the counter: components are decomposed and memoised, and the
// branching variable comes from an elimination order of the primal graph, which on
// cardinality structure was worth three orders of magnitude over an occurrence
// count.
package gomisat
