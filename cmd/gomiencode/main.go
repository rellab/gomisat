// Command gomiencode compares encodings of the same reliability model by what it
// costs to count them.
//
// This is the question DESIGN.md phase 5 ended on. The same k-out-of-n system can
// be stated as a chain of gates, as a balanced tree of unary counters, or as one
// linear constraint over order-encoded variables. All three describe one function,
// which the program checks by requiring the same answer from all of them and from
// the closed form; what differs is the size of the formula and the cost of counting
// it.
package main

import (
	"flag"
	"fmt"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/rellab/gomisat/pkg/csp"
	"github.com/rellab/gomisat/pkg/gomisat"
	"github.com/rellab/gomisat/pkg/reliability"
)

// An encoding of one system, ready to be counted.
type encoded struct {
	name    string
	clauses [][]int64
	numVars int
	weights *gomisat.Weights
	// failed[i] is the literal that conditions on component i having failed.
	failed []gomisat.Lit
}

func (e *encoded) solver() (*gomisat.Solver, *gomisat.SolverOptions) {
	s := gomisat.NewSolver()
	options := gomisat.DefaultSolverOptions()
	s.AddCNF(&gomisat.CNF{NumVars: e.numVars, NumClauses: len(e.clauses), Clauses: e.clauses}, options)
	return s, options
}

func (e *encoded) literals() int {
	n := 0
	for _, c := range e.clauses {
		n += len(c)
	}
	return n
}

// gateEncoding builds k-out-of-n from pkg/reliability, either the sequential chain
// or the balanced totalizer.
func gateEncoding(n, k int, p float64, balanced bool) *encoded {
	m := reliability.New()
	events := m.Events("c", n)
	if balanced {
		m.Assert(m.AtLeast(k, events...))
	} else {
		m.Assert(m.AtLeastSequential(k, events...))
	}
	for _, e := range events {
		m.SetProbability(e, p)
	}
	clauses, numVars, _ := m.CNF()
	failed := make([]gomisat.Lit, 0, n)
	for _, e := range events {
		// The event is true when the component works, so failing is its negation.
		failed = append(failed, gomisat.MkLit(gomisat.Var(e-1), true))
	}
	name := "gates, sequential"
	if balanced {
		name = "gates, balanced"
	}
	return &encoded{name: name, clauses: clauses, numVars: numVars, weights: m.Weights(), failed: failed}
}

// linearEncoding builds the same system as one linear constraint over variables of
// domain 0..1, order encoded.
func linearEncoding(n, k int, p float64) *encoded {
	m := csp.New()
	vars := make([]*csp.IntVar, n)
	sum := make(map[*csp.IntVar]int, n)
	for i := range vars {
		vars[i] = m.IntVarRange(0, 1)
		sum[vars[i]] = 1
	}
	m.Add(csp.GeZero(csp.NewSum(sum, -k))) // at least k of them are 1
	clauses, numVars := m.Build()

	// A variable of domain 0..1 has one code, meaning "at most 0", so that code is
	// true exactly when the component has failed.
	weights := gomisat.NewWeights(numVars)
	failed := make([]gomisat.Lit, 0, n)
	for _, v := range vars {
		code := m.IntCodes(v)[0]
		weights.Set(gomisat.Var(code-1), big.NewFloat(1-p), big.NewFloat(p))
		failed = append(failed, gomisat.MkLit(gomisat.Var(code-1), false))
	}
	return &encoded{name: "linear, order encoded", clauses: clauses, numVars: numVars,
		weights: weights, failed: failed}
}

// multiStateGates builds the reliability-relevant shape with pkg/reliability: each
// component has several states, one-hot encoded, and the system needs k of them in
// state threshold or better.
func multiStateGates(n, k, threshold, states int, balanced bool) *encoded {
	m := reliability.New()
	working := make([]reliability.Node, 0, n)
	all := make([][]reliability.Node, 0, n)
	for i := 0; i < n; i++ {
		ind := m.MultiStateOneHot("c", states)
		all = append(all, ind)
		working = append(working, m.AtLeastState(ind, threshold))
	}
	if balanced {
		m.Assert(m.AtLeast(k, working...))
	} else {
		m.Assert(m.AtLeastSequential(k, working...))
	}
	clauses, numVars, _ := m.CNF()
	// Conditioning on component i being in its worst state.
	failed := make([]gomisat.Lit, 0, n)
	for _, ind := range all {
		failed = append(failed, gomisat.MkLit(gomisat.Var(ind[0]-1), false))
	}
	name := "one-hot + gates, balanced"
	if balanced == false {
		name = "one-hot + gates, sequential"
	}
	// Unweighted, for the reason given in multiStateLinear.
	return &encoded{name: name, clauses: clauses, numVars: numVars, failed: failed}
}

// multiStateLinear builds the same system in the constraint language: the state of a
// component is an integer variable, an indicator of domain 0..1 is tied to "state at
// least threshold", and the system is one linear constraint over the indicators.
func multiStateLinear(n, k, threshold, states int) *encoded {
	m := csp.New()
	stateVars := make([]*csp.IntVar, n)
	indicators := make([]*csp.IntVar, n)
	sum := make(map[*csp.IntVar]int, n)
	for i := 0; i < n; i++ {
		stateVars[i] = m.IntVarRange(0, states-1)
		indicators[i] = m.IntVarRange(0, 1)
		m.Add(csp.Iff(
			csp.GeZero(csp.NewSum(map[*csp.IntVar]int{indicators[i]: 1}, -1)),
			csp.GeZero(csp.NewSum(map[*csp.IntVar]int{stateVars[i]: 1}, -threshold)),
		))
		sum[indicators[i]] = 1
	}
	m.Add(csp.GeZero(csp.NewSum(sum, -k)))
	clauses, numVars := m.Build()

	// No weights. The order encoding of a state variable cannot carry a state
	// distribution on independent literal weights: a component in state s leaves
	// every level above s false, and those literals would bring their own weights
	// into the product. That is the finding recorded under phase 3, and it is why
	// this family is compared on unweighted counts -- the number of state
	// configurations the system survives -- which is well defined for all three
	// encodings and has a closed form to check against. Weighting it would mean
	// stating the states one-hot here too, which is a different comparison.
	failed := make([]gomisat.Lit, 0, n)
	for i := 0; i < n; i++ {
		// The worst state is "state <= 0".
		failed = append(failed, gomisat.MkLit(gomisat.Var(m.IntCodes(stateVars[i])[0]-1), false))
	}
	return &encoded{name: "linear + order encoded states", clauses: clauses, numVars: numVars,
		failed: failed}
}

// closedForm is the reliability of a k-out-of-n system of identical components.
func closedForm(n, k int, p float64) float64 {
	total := 0.0
	for i := k; i <= n; i++ {
		c, _ := new(big.Float).SetInt(new(big.Int).Binomial(int64(n), int64(i))).Float64()
		total += c * math.Pow(p, float64(i)) * math.Pow(1-p, float64(n-i))
	}
	return total
}

func main() {
	sizes := flag.String("n", "10,20,30,40", "system sizes to measure")
	family := flag.String("family", "binary", "which family: binary (k-out-of-n over working/failed components) or multistate")
	states := flag.Int("states", 4, "states per component, for the multistate family")
	threshold := flag.Int("threshold", 2, "the state at which a component counts as working, for the multistate family")
	prob := flag.Float64("p", 0.9, "component reliability")
	skipSequential := flag.Bool("skip-sequential", false, "leave out the sequential gate encoding, which is hopeless above about sixty components")
	sample := flag.Int("queries", 8, "how many components to condition on when measuring a query sequence")
	budget := flag.Float64("budget", 60, "seconds to allow one encoding at one size")
	flag.Parse()

	var ns []int
	for _, f := range splitInts(*sizes) {
		ns = append(ns, f)
	}

	if *family == "multistate" {
		multistate(ns, *states, *threshold, *sample, *budget, *skipSequential)
		return
	}
	fmt.Printf("k-out-of-n with k = n/2, component reliability %.2f\n\n", *prob)
	fmt.Printf("%4s  %-22s %8s %9s %7s   %9s %9s   %11s %9s %7s\n",
		"n", "encoding", "clauses", "literals", "vars",
		"dec", "count s", "dec/query", "reused", "ratio")
	fmt.Printf("(the query sequence conditions on %d components; a dash means the time budget of %.0fs ran out)\n",
		*sample, *budget)

	for _, n := range ns {
		k := n / 2
		want := closedForm(n, k, *prob)
		encodings := []*encoded{gateEncoding(n, k, *prob, true), linearEncoding(n, k, *prob)}
		if *skipSequential == false {
			encodings = append([]*encoded{gateEncoding(n, k, *prob, false)}, encodings...)
		}
		for _, e := range encodings {
			// One count, checked against the closed form.
			s, options := e.solver()
			start := time.Now()
			value, stats := s.WeightedCount(options, gomisat.DefaultCountOptions(), e.weights)
			single := time.Since(start).Seconds()
			got, _ := value.Float64()
			if math.Abs(got-want)/want > 1e-9 {
				fmt.Fprintf(os.Stderr, "n=%d %s: reliability %.15g, closed form %.15g\n",
					n, e.name, got, want)
				os.Exit(1)
			}

			// The query sequence: condition on a sample of components having
			// failed. A sample rather than all of them, because the arm with the
			// cache cleared pays for a full count every time.
			queries := *sample
			if queries > n {
				queries = n
			}
			run := func(keep bool) (uint64, bool) {
				s, options := e.solver()
				study := gomisat.NewStudy(s, options, gomisat.DefaultCountOptions(), e.weights)
				start := time.Now()
				for _, lit := range e.failed[:queries] {
					study.Count(lit)
					if keep == false {
						study.ClearCache()
					}
					if time.Since(start).Seconds() > *budget {
						return 0, false
					}
				}
				return study.Totals().Decisions, true
			}
			fresh, freshOK := run(false)
			reused, reusedOK := run(true)

			fmt.Printf("%4d  %-22s %8d %9d %7d   %9d %9.3f",
				n, e.name, len(e.clauses), e.literals(), e.numVars, stats.Decisions, single)
			if freshOK && reusedOK {
				fmt.Printf("   %11d %9d %7.3f\n",
					fresh/uint64(queries), reused/uint64(queries), float64(reused)/float64(fresh))
			} else {
				fmt.Printf("   %11s %9s %7s\n", "-", "-", "-")
			}
		}
		fmt.Println()
	}
}

func splitInts(s string) []int {
	var out []int
	cur, have := 0, false
	for i := 0; i <= len(s); i++ {
		if i < len(s) && s[i] >= '0' && s[i] <= '9' {
			cur = cur*10 + int(s[i]-'0')
			have = true
			continue
		}
		if have {
			out = append(out, cur)
		}
		cur, have = 0, false
	}
	return out
}

// multistate runs the family the reliability work is actually about: components with
// several states, a threshold saying which states count as working, and a
// k-out-of-n system over those.
//
// The comparison is of unweighted counts -- how many state configurations the system
// survives -- because the order encoding cannot carry a state distribution on
// independent literal weights. All three encodings must agree with each other and
// with the closed form, which is the sum over i from k to n of C(n,i) a^i b^(n-i)
// with a the number of working states and b the number of failed ones.
func multistate(ns []int, states, threshold, sample int, budget float64, skipSequential bool) {
	working := states - threshold
	failedStates := threshold
	fmt.Printf("components of %d states, working at state %d or better, k = n/2\n", states, threshold)
	fmt.Printf("counting state configurations, not probability; see the comment in the source\n\n")
	fmt.Printf("%4s  %-30s %8s %9s %7s   %9s %9s   %11s %9s %7s\n",
		"n", "encoding", "clauses", "literals", "vars", "dec", "count s", "dec/query", "reused", "ratio")

	for _, n := range ns {
		k := n / 2
		want := big.NewInt(0)
		for i := k; i <= n; i++ {
			term := new(big.Int).Binomial(int64(n), int64(i))
			term.Mul(term, new(big.Int).Exp(big.NewInt(int64(working)), big.NewInt(int64(i)), nil))
			term.Mul(term, new(big.Int).Exp(big.NewInt(int64(failedStates)), big.NewInt(int64(n-i)), nil))
			want.Add(want, term)
		}

		encodings := []*encoded{
			multiStateGates(n, k, threshold, states, true),
			multiStateLinear(n, k, threshold, states),
		}
		if skipSequential == false {
			encodings = append([]*encoded{multiStateGates(n, k, threshold, states, false)}, encodings...)
		}
		for _, e := range encodings {
			s, options := e.solver()
			start := time.Now()
			got, stats := s.CountModels(options, gomisat.DefaultCountOptions())
			single := time.Since(start).Seconds()
			if got.Cmp(want) != 0 {
				fmt.Fprintf(os.Stderr, "n=%d %s: %v configurations, closed form %v\n", n, e.name, got, want)
				os.Exit(1)
			}

			queries := sample
			if queries > n {
				queries = n
			}
			run := func(keep bool) (uint64, bool) {
				s, options := e.solver()
				study := gomisat.NewStudy(s, options, gomisat.DefaultCountOptions(), nil)
				start := time.Now()
				for _, lit := range e.failed[:queries] {
					study.Count(lit)
					if keep == false {
						study.ClearCache()
					}
					if time.Since(start).Seconds() > budget {
						return 0, false
					}
				}
				return study.Totals().Decisions, true
			}
			fresh, freshOK := run(false)
			reused, reusedOK := run(true)

			fmt.Printf("%4d  %-30s %8d %9d %7d   %9d %9.3f",
				n, e.name, len(e.clauses), e.literals(), e.numVars, stats.Decisions, single)
			if freshOK && reusedOK {
				fmt.Printf("   %11d %9d %7.3f\n",
					fresh/uint64(queries), reused/uint64(queries), float64(reused)/float64(fresh))
			} else {
				fmt.Printf("   %11s %9s %7s\n", "-", "-", "-")
			}
		}
		fmt.Println()
	}
}
