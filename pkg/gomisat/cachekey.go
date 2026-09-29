package gomisat

// Cache keys.
//
// A component is identified by the clauses it still has and the variables of
// theirs that are still unassigned. Written out, that is hundreds of bytes for a
// large component, and a counter stores hundreds of thousands of them: the keys
// themselves became the memory and the copying they caused became the time.
//
// So the default key is a 128-bit hash of that description rather than the
// description itself. Two distinct components can then collide and be given each
// other's count. With k entries the chance of any collision is about
// k^2 / 2^129, which for a million entries is around 10^-27 -- far below the
// probability of the machine getting an arithmetic answer wrong. This is the same
// trade every search-based counter makes; it is what the "probabilistic" in
// GANAK's description of itself refers to.
//
// Since it is a trade and not a certainty, the exact key stays available:
// CountOptions.ExactCache keys the cache by the full description instead, and the
// regression tests run both and require the same answers, which is also what
// checks the hash.

// cacheKey is a 128-bit digest of a component.
type cacheKey struct {
	a, b uint64
}

const (
	fnvOffset1 uint64 = 14695981039346656037
	fnvPrime1  uint64 = 1099511628211
	fnvOffset2 uint64 = 1099511628211
	fnvPrime2  uint64 = 1111111111111111111
)

// hashComponent digests the variables and clauses of a component into two
// independent 64-bit values.
func hashComponent(vars []Var, clauses []CRef) cacheKey {
	a, b := fnvOffset1, fnvOffset2
	mix := func(x uint64) {
		a = (a ^ x) * fnvPrime1
		b = (b + x) * fnvPrime2
		b ^= b >> 31
	}
	mix(uint64(len(vars)))
	for _, v := range vars {
		mix(uint64(v))
	}
	// A separator, so that moving an element between the two lists changes the
	// digest.
	mix(^uint64(len(clauses)))
	for _, ref := range clauses {
		mix(uint64(ref))
	}
	return cacheKey{a: splitMix(a), b: splitMix(b)}
}

// splitMix is the finaliser of splitmix64: it spreads the bits of an accumulator
// that has only been multiplied and xored.
func splitMix(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}
