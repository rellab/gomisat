package gomisat

import "sort"

// Literal block distance and tiered management of learnt clauses.
//
// The LBD of a clause is the number of distinct decision levels among its
// literals (Audemard and Simon, Glucose). It predicts usefulness better than
// activity does: a clause with a low LBD connects few levels and tends to keep
// propagating long after the activity that produced it has decayed.
//
// Clauses are kept in three tiers, following Glucose 4 and its descendants:
//
//	core   LBD <= LBDCore        never deleted
//	mid    LBD <= LBDTier2       kept while it is still being used
//	local  everything else       half of it is deleted at every reduction
//
// A mid clause that has not taken part in conflict analysis for Tier2MaxAge
// conflicts is demoted to local. A clause whose LBD improves while it is used is
// promoted. Binary and locked clauses are never deleted, as before.

type clauseTier uint8

const (
	tierLocal clauseTier = iota
	tierMid
	tierCore
)

func (t clauseTier) String() string {
	switch t {
	case tierCore:
		return "core"
	case tierMid:
		return "mid"
	default:
		return "local"
	}
}

// LBD returns the literal block distance recorded for a learnt clause, or 0 for
// a problem clause.
func (c *Clause) LBD() int { return c.lbd }

// computeLBD counts the distinct decision levels of the given literals. It has
// to be called before backjumping, while the levels still describe the trail
// that produced the clause.
func (s *Solver) computeLBD(lits []Lit) int {
	s.lbdGeneration++
	n := 0
	for _, p := range lits {
		level := s.vardata[p.Var()].level
		for level >= len(s.lbdStamp) {
			s.lbdStamp = append(s.lbdStamp, 0)
		}
		if s.lbdStamp[level] != s.lbdGeneration {
			s.lbdStamp[level] = s.lbdGeneration
			n++
		}
	}
	return n
}

func tierOf(lbd int, options *SolverOptions) clauseTier {
	switch {
	case lbd <= options.LBDCore:
		return tierCore
	case lbd <= options.LBDTier2:
		return tierMid
	default:
		return tierLocal
	}
}

// noteLearnt records the LBD of a freshly learnt clause and files it in a tier.
func (s *Solver) noteLearnt(c *Clause, lbd int, options *SolverOptions) {
	c.lbd = lbd
	c.touched = s.Conflicts
	c.tier = tierOf(lbd, options)
}

// noteUsed is called when a learnt clause takes part in conflict analysis. It
// refreshes the clause's age and, if the clause now spans fewer levels than when
// it was learnt, records the better LBD and promotes it.
func (s *Solver) noteUsed(c *Clause, options *SolverOptions) {
	c.touched = s.Conflicts
	if options.UseLBD == false {
		return
	}
	if lbd := s.computeLBD(c.lits); lbd < c.lbd {
		c.lbd = lbd
		if tier := tierOf(lbd, options); tier > c.tier {
			c.tier = tier
		}
	}
}

// reduceDBTiered deletes half of the local tier, after demoting the mid-tier
// clauses that have gone unused. It returns the number of clauses deleted.
func (s *Solver) reduceDBTiered(options *SolverOptions) int {
	kept := make([]*Clause, 0, len(s.learnts))
	candidates := make([]*Clause, 0, len(s.learnts))

	for _, c := range s.learnts {
		if c.tier == tierMid && s.Conflicts-c.touched > options.Tier2MaxAge {
			c.tier = tierLocal
		}
		if len(c.lits) <= 2 || c.tier != tierLocal || s.Locked(c) {
			kept = append(kept, c)
			continue
		}
		candidates = append(candidates, c)
	}

	// Worst first: many levels, and among equals the least recently useful.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].lbd != candidates[j].lbd {
			return candidates[i].lbd > candidates[j].lbd
		}
		return candidates[i].activity < candidates[j].activity
	})

	drop := len(candidates) / 2
	for i, c := range candidates {
		if i < drop {
			s.RemoveClause(c)
			continue
		}
		kept = append(kept, c)
	}
	s.learnts = kept
	return drop
}
