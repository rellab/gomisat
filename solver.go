package gomisat

type SolverOptions struct {
	Verbosity                  int
	VarDecay                   float64
	ClauseDecay                float64
	RandomVarFreq              float64
	RandomSeed                 float64
	LubyRestart                bool
	CcminMode                  int     // Controls conflict clause minimization (0=none, 1=basic, 2=deep).
	PhaseSaving                int     // Controls the level of phase saving (0=none, 1=limited, 2=full).
	RndPol                     bool    // Use random polarities for branching heuristics.
	RndInitAct                 bool    // Initialize variable activities with a small random value.
	GarbageFrac                float64 // The fraction of wasted memory allowed before a garbage collection is triggered.
	MinLearntsLim              int     // Minimum number to set the learnts limit to.
	RestartFirst               int     // The initial restart limit. (default 100)
	RestartInc                 float64 // The factor with which the restart limit is multiplied in each restart. (default 1.5)
	LearntsizeFactor           float64 // The intitial limit for learnt clauses is a factor of the original clauses. (default 1 / 3)
	LeantsizeInt               float64 // The limit for learnt clauses is multiplied with this factor each restart. (default 1.1)
	LearntsizeAdjustStartConfl int
	LearntsizeAdjustInc        float64
}

type SolverStatistics struct {
	Solves          uint64
	Starts          uint64
	Decisions       uint64
	Propagations    uint64
	Conflicts       uint64
	NumClauses      uint64
	ClausesLiterals uint64
	LearntsLiterals uint64
}

func NewSolverStatistics() SolverStatistics {
	return SolverStatistics{
		Solves:          0,
		Starts:          0,
		Decisions:       0,
		Propagations:    0,
		Conflicts:       0,
		NumClauses:      0,
		ClausesLiterals: 0,
		LearntsLiterals: 0,
	}
}

type Solver struct {
	clauses     []*Clause // List of problem clauses.
	learnts     []*Clause // List of learnt clauses.
	trail       []Lit     // Assignment stack; stores all assigments made in the order they were made.
	trailLim    []int     // Separator indices for different decision levels in 'trail'.
	assumptions []Lit     // Current set of assumptions provided to solve by the user.

	activity map[Var]float64 // A heuristic measurement of the activity of a variable.
	assigns  map[Var]LBool   // The current assignments.
	polarity map[Var]byte    // The preferred polarity of each variable.
	userPol  map[Var]LBool   // The users preferred polarity of each variable.
	decision map[Var]byte    // Declares if a variable is eligible for selection in the decision heuristic.
	// 	vardata  map[Var]VarData // Stores reason and level for each variable.
}
