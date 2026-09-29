package gomisat

// Restart policies.
//
// MiniSat restarts on a Luby sequence of conflict counts: a schedule fixed in
// advance, blind to how the search is going. Modern solvers decide from the
// quality of what they are learning. The rule here is Biere's: keep a fast and a
// slow exponential moving average of the literal block distance of the learnt
// clauses and restart when the fast one is enough above the slow one, meaning the
// clauses being learnt right now are worse than the run's own standard.
//
// The optional blocking rule is Glucose's: when the trail is much longer than
// usual the assignment looks promising, so a restart would throw away good work.
// Blocking is expressed here by pulling the fast average back to the slow one,
// which is the same thing Glucose achieves by clearing its recent-LBD queue.

const (
	RestartLuby      = "luby"      // MiniSat's Luby sequence
	RestartGeometric = "geometric" // fixed geometric growth
	RestartEMA       = "ema"       // Biere: fast vs slow EMA of the LBD
	RestartEMABlock  = "ema-block" // the same, with Glucose's trail blocking
)

// noteConflict feeds one conflict into the restart statistics. trailSize is the
// length of the trail at the time of the conflict, before backjumping.
func (s *Solver) noteConflict(trailSize int, lbd int, options *SolverOptions) {
	value := float64(lbd)
	if s.emaReady == false {
		s.emaFastLBD = value
		s.emaSlowLBD = value
		s.emaTrail = float64(trailSize)
		s.emaReady = true
		return
	}
	s.emaFastLBD += options.RestartEMAFast * (value - s.emaFastLBD)
	s.emaSlowLBD += options.RestartEMASlow * (value - s.emaSlowLBD)
	s.emaTrail += options.RestartEMASlow * (float64(trailSize) - s.emaTrail)

	// Glucose does not block before the trail average has had time to settle;
	// without that bound the rule fires from the first conflicts on and the
	// solver stops restarting almost entirely.
	if options.RestartPolicy == RestartEMABlock && s.Conflicts > options.RestartBlockAfter &&
		s.emaTrail > 0 && float64(trailSize) > options.RestartBlockMargin*s.emaTrail {
		// The assignment is unusually deep: hold the restart off by forgetting
		// the recent average.
		s.emaFastLBD = s.emaSlowLBD
		s.Blocked++
	}
}

// restartDue reports whether the current search segment should be abandoned.
// It is only consulted by the dynamic policies; the Luby and geometric schedules
// are enforced by the conflict budget the caller passes to search.
func (s *Solver) restartDue(options *SolverOptions) bool {
	switch options.RestartPolicy {
	case RestartEMA, RestartEMABlock:
	default:
		return false
	}
	if s.emaReady == false || s.Conflicts-s.restartBase < options.RestartMinInterval {
		return false
	}
	return s.emaFastLBD > options.RestartMargin*s.emaSlowLBD
}
