package csp

import (
	"math/big"

	"github.com/rellab/gomisat/pkg/gomisat"
)

// Solving and counting a model.
//
// The encoding is the order encoding: an integer variable with the domain
// d[0] < d[1] < ... < d[n-1] is represented by n-1 propositional variables, the
// k-th meaning "this variable is at most d[k]", with clauses making the chain
// monotone. A solution therefore reads off as the first of those that holds.

// Build encodes the model and returns its clauses in DIMACS numbering together
// with the number of propositional variables. It is idempotent: the encoding is
// done once and kept.
func (c *Model) Build() (clauses [][]int64, numVars int) {
	if c.built == false {
		c.CNF()
		c.genBase()
		c.Encode()
		c.built = true
	}
	if c.unsat {
		// The empty clause: unsatisfiable, and every consumer agrees on that.
		return [][]int64{{}}, c.numCodes
	}
	out := make([][]int64, 0, len(c.sat))
	for _, clause := range c.sat {
		lits := make([]int64, 0, len(clause))
		for _, lit := range clause {
			lits = append(lits, int64(lit))
		}
		out = append(out, lits)
	}
	return out, c.numCodes
}

func (c *Model) newSolver() (*gomisat.Solver, *gomisat.SolverOptions) {
	clauses, numVars := c.Build()
	s := gomisat.NewSolver()
	options := gomisat.DefaultSolverOptions()
	s.AddCNF(&gomisat.CNF{NumVars: numVars, NumClauses: len(clauses), Clauses: clauses}, options)
	return s, options
}

// A Solution is an assignment to the variables of a model.
type Solution struct {
	model *Model
	value map[gomisat.Var]gomisat.LBool
}

// Solve looks for a solution. The second result is false when the model has none.
func (c *Model) Solve() (*Solution, bool) {
	s, options := c.newSolver()
	if s.Solve(options) != gomisat.LTrue {
		return nil, false
	}
	return &Solution{model: c, value: s.Model()}, true
}

// holds reports whether the propositional variable with the given DIMACS code is
// true in the solution.
func (sol *Solution) holds(code int) bool {
	return sol.value[gomisat.Var(code-1)] == gomisat.LTrue
}

// Int returns the value of an integer variable.
func (sol *Solution) Int(v *IntVar) int {
	base := sol.model.baseCode[v.id]
	// The k-th code means "v <= domain[k]", and the chain makes the true codes a
	// suffix, so the value is the first one that holds.
	for k := 0; k < v.domain.size()-1; k++ {
		if sol.holds(base + k) {
			return v.domain[k]
		}
	}
	return v.domain[v.domain.size()-1]
}

// Bool returns the value of a Boolean variable.
func (sol *Solution) Bool(v *BoolVar) bool {
	return sol.holds(sol.model.baseCode[v.id])
}

// Count returns the number of solutions of the model.
//
// It requires the definitional conversion: with the polarity-optimised one an
// auxiliary variable can take either value and the same solution is counted more
// than once. Count refuses rather than returning a number that looks right.
func (c *Model) Count() (*big.Int, bool) {
	if c.Definitional == false {
		return nil, false
	}
	s, options := c.newSolver()
	count, _ := s.CountModels(options, gomisat.DefaultCountOptions())
	return count, true
}

// WeightedCount returns the sum over the solutions of the product of the weights
// of their literals, which is what turns a model of a system into its probability.
// weights is called for each propositional variable of an integer or Boolean
// variable and returns the weight of it being true and of it being false; return
// 1, 1 to leave a variable counted rather than weighted.
func (c *Model) WeightedCount(weight func(code int) (whenTrue, whenFalse float64)) (*big.Float, bool) {
	if c.Definitional == false {
		return nil, false
	}
	s, options := c.newSolver()
	w := gomisat.NewWeights(s.NumVars())
	for code := 1; code <= c.numCodes; code++ {
		t, f := weight(code)
		w.Set(gomisat.Var(code-1), big.NewFloat(t), big.NewFloat(f))
	}
	value, _ := s.WeightedCount(options, gomisat.DefaultCountOptions(), w)
	return value, true
}

// Code returns the DIMACS code of a Boolean variable, and the base code of an
// integer variable: its k-th code, for k from 0 to the domain size minus two,
// means "the variable is at most its k-th domain value". Weights and assumptions
// are expressed in terms of these.
func (c *Model) Code(id int) int {
	c.Build()
	return c.baseCode[id]
}

// BoolCode returns the DIMACS code of a Boolean variable.
func (c *Model) BoolCode(v *BoolVar) int { return c.Code(v.id) }

// IntCodes returns the DIMACS codes of an integer variable, in domain order.
func (c *Model) IntCodes(v *IntVar) []int {
	base := c.Code(v.id)
	out := make([]int, 0, v.domain.size()-1)
	for k := 0; k < v.domain.size()-1; k++ {
		out = append(out, base+k)
	}
	return out
}

// Domain returns the values an integer variable may take.
func (v *IntVar) Domain() []int { return append([]int(nil), v.domain...) }
