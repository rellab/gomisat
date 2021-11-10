package sat

func (s *Solver) AddClauseFromCode(codes []int, options *SolverOptions) {
	lits := make([]Lit, 0, len(codes))
	for _, v := range codes {
		switch {
		case v > 0:
			s.addVar(v-1, options) // v starts with 0
			lits = append(lits, MkLit(Var(v-1), false))
		case v < 0:
			s.addVar(-(v + 1), options) // v starts with 0
			lits = append(lits, MkLit(Var(-(v+1)), true))
		default:
		}
	}
	s.AddClause(lits...)
}

// add a variable from a general int
// This function is called from AddClauseFromCode only
func (s *Solver) addVar(v int, options *SolverOptions) {
	for v >= int(s.nextVar) {
		s.NewVar(true, options)
	}
}
