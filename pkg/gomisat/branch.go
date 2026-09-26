package gomisat

// Branching order for counting.
//
// A counting search lives or dies by how soon it breaks a component apart, and
// that is a property of the variable order, not of the search. The occurrence
// count, which is what a solver-flavoured heuristic reaches for first, says
// nothing about separation: on a cardinality constraint every variable occurs
// about equally often and the search stays exponential.
//
// What is wanted is a variable that cuts the component in two. The cheap and
// well-understood way to find such variables is an elimination order of the
// primal graph -- the graph whose vertices are the variables and whose edges join
// variables sharing a clause. Eliminating a vertex means removing it and joining
// its remaining neighbours, the "elimination game"; a min-degree order keeps the
// cliques it creates small. Vertices eliminated *last* are the ones that hold the
// graph together, so those are the ones to branch on first. That is a poor
// relation of what sharpSAT-TD does properly with a tree decomposition, but it is
// the same idea and it costs one pass over the formula.

// Branching strategies for CountOptions.Branching.
const (
	// BranchOccurrence picks the variable occurring in the most residual clauses.
	BranchOccurrence = "occurrence"
	// BranchEliminationOrder picks the variable eliminated latest by a min-degree
	// elimination order of the primal graph, breaking ties by occurrence.
	BranchEliminationOrder = "order"
)

// maxEliminationVars is the size beyond which the elimination order is not worth
// computing; the counter falls back to occurrence counting. Counting is hopeless
// on such formulas anyway, but the fallback keeps the cost bounded.
const maxEliminationVars = 20000

// eliminationScores returns, for each variable, its position in a min-degree
// elimination order of the primal graph of the given clauses. A larger score
// means the variable survived longer, so it is closer to the root of the implied
// tree decomposition and is a better variable to branch on.
//
// It returns nil when the formula is too large for this to be worthwhile.
func (s *Solver) eliminationScores(clauses []CRef) []int32 {
	n := s.NumVars()
	if n == 0 || n > maxEliminationVars {
		return nil
	}

	adj := make([]map[Var]struct{}, n)
	present := make([]bool, n)
	for _, ref := range clauses {
		if s.arena.Dead(ref) {
			continue
		}
		lits := s.arena.Lits(ref)
		for i, p := range lits {
			u := p.Var()
			present[u] = true
			if adj[u] == nil {
				adj[u] = make(map[Var]struct{}, len(lits))
			}
			for j, q := range lits {
				if i == j {
					continue
				}
				adj[u][q.Var()] = struct{}{}
			}
		}
	}

	scores := make([]int32, n)
	remaining := make([]bool, n)
	left := 0
	for v := 0; v < n; v++ {
		if present[v] {
			remaining[v] = true
			left++
		}
	}

	// Variables no clause mentions come first: they constrain nothing, so they
	// are the last thing worth branching on.
	next := int32(0)
	for v := 0; v < n; v++ {
		if present[v] == false {
			scores[v] = next
			next++
		}
	}

	degree := func(v Var) int {
		d := 0
		for u := range adj[v] {
			if remaining[u] {
				d++
			}
		}
		return d
	}

	for left > 0 {
		best, bestDegree := VarUndef, 1<<30
		for v := 0; v < n; v++ {
			if remaining[v] == false {
				continue
			}
			if d := degree(Var(v)); d < bestDegree {
				best, bestDegree = Var(v), d
			}
		}

		// Eliminate best: its remaining neighbours become a clique.
		nbrs := make([]Var, 0, bestDegree)
		for u := range adj[best] {
			if remaining[u] {
				nbrs = append(nbrs, u)
			}
		}
		for i, a := range nbrs {
			for _, b := range nbrs[i+1:] {
				adj[a][b] = struct{}{}
				adj[b][a] = struct{}{}
			}
		}
		remaining[best] = false
		left--
		scores[best] = next
		next++
	}
	return scores
}
