package gomisat

import (
	"log"
	"sort"
)

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
//	mid    LBD <= LBDTier2       demoted to local once it goes unused
//	local  everything else       deletion candidate
//
// At every reduction, half of the candidates go, worst LBD first. Only the core
// tier, binary clauses and clauses that are currently a reason are safe. With
// ProtectTier2 the mid tier is spared as well while it is still in use, which is
// what this used to do unconditionally; on the benchmark corpus that protects so
// much of the database that the tiers stop paying for themselves.
//
// A mid clause that has not taken part in conflict analysis for Tier2MaxAge
// conflicts is demoted to local. A clause whose LBD improves while it is used is
// promoted.

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
func (a *clauseArena) LBD(c CRef) int { return int(a.meta[c].lbd) }

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
func (s *Solver) noteLearnt(c CRef, lbd int, options *SolverOptions) {
	m := &s.arena.meta[c]
	m.lbd = int32(lbd)
	m.touched = uint32(s.Conflicts)
	m.tier = tierOf(lbd, options)
}

// noteUsed is called when a learnt clause takes part in conflict analysis. It
// refreshes the clause's age and, if the clause now spans fewer levels than when
// it was learnt, records the better LBD and promotes it.
func (s *Solver) noteUsed(c CRef, options *SolverOptions) {
	m := &s.arena.meta[c]
	m.touched = uint32(s.Conflicts)
	if options.UseLBD == false {
		return
	}
	if lbd := s.computeLBD(s.arena.Lits(c)); int32(lbd) < m.lbd {
		m.lbd = int32(lbd)
		if tier := tierOf(lbd, options); tier > m.tier {
			m.tier = tier
		}
	}
}

// reduceDBTiered deletes half of the local tier, after demoting the mid-tier
// clauses that have gone unused. It returns the number of clauses deleted.
func (s *Solver) reduceDBTiered(options *SolverOptions) int {
	kept := make([]CRef, 0, len(s.learnts))
	candidates := make([]CRef, 0, len(s.learnts))

	for _, c := range s.learnts {
		m := &s.arena.meta[c]
		if m.tier == tierMid && uint64(uint32(s.Conflicts)-m.touched) > options.Tier2MaxAge {
			m.tier = tierLocal
		}
		protected := m.tier == tierCore
		if options.ProtectTier2 && m.tier == tierMid {
			protected = true
		}
		if int(m.size) <= 2 || protected || s.Locked(c) {
			kept = append(kept, c)
			continue
		}
		candidates = append(candidates, c)
	}

	// Worst first: many levels, and among equals the least recently useful.
	sort.Slice(candidates, func(i, j int) bool {
		mi, mj := &s.arena.meta[candidates[i]], &s.arena.meta[candidates[j]]
		if mi.lbd != mj.lbd {
			return mi.lbd > mj.lbd
		}
		return mi.activity < mj.activity
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
	s.collectGarbage()
	return drop
}

// reductionDue reports whether the learnt clause database should be reduced now,
// and advances the schedule when it says yes.
//
// With the LBD tiers the schedule is driven by the conflict count and the
// interval grows after every reduction, which is what Glucose does. MiniSat
// instead compares the size of the database against a budget that itself grows,
// which does not bound the database: a protected tier can fill the budget and
// then every further conflict asks for a reduction that cannot free anything.
func (s *Solver) reductionDue(options *SolverOptions) bool {
	if options.NoReduce {
		return false
	}
	if options.ReduceByConflicts == false {
		return float64(len(s.learnts)-len(s.trail)) >= s.maxLearnts
	}
	if s.reduceInterval == 0 {
		s.reduceInterval = options.ReduceFirst
		s.reduceAt = options.ReduceFirst
	}
	if s.Conflicts < s.reduceAt {
		return false
	}
	s.reduceInterval += options.ReduceInc
	s.reduceAt = s.Conflicts + s.reduceInterval
	return true
}

// garbageFraction is the share of the literal store that belongs to deleted
// clauses.
func (s *Solver) garbageFraction() float64 { return s.arena.wastedFraction() }

// collectGarbage sweeps the watch lists, compacts the literal store and makes the
// slots of deleted clauses available again, in that order: the sweep is what makes
// recycling safe, because afterwards nothing refers to a deleted clause.
//
// Clause references are indices into the metadata array, so compaction does not
// have to rewrite anything outside the arena. MiniSat, whose references are
// offsets into the literal store, relocates every watcher and every reason here.
func (s *Solver) collectGarbage() {
	if s.arena.wastedFraction() < 0.2 && len(s.arena.pending) == 0 {
		return
	}
	if debug {
		log.Println("collectGarbage: literals", len(s.arena.lits), "wasted", s.arena.wasted,
			"pending", len(s.arena.pending))
	}
	s.sweepWatches()
	if s.arena.wastedFraction() >= 0.2 {
		s.arena.compact()
	}
	s.arena.recycle()
}
