package gomisat

// chunkArena hands out slices with a stack discipline, for data whose lifetime
// follows a depth-first recursion.
//
// The counter builds a component at every node of its search: two slices, for the
// clauses that are left and for their unassigned variables. Allocating those from
// the heap made component construction three quarters of everything the counter
// allocated. They cannot come from one growable buffer either, because appending
// to it would move the backing array and invalidate the slices already handed out.
//
// So the storage is a list of fixed chunks that are never reallocated. A slice
// handed out stays valid for as long as the arena lives, and a mark taken on entry
// to a subtree releases everything that subtree allocated when it is restored.
type chunkArena[T any] struct {
	chunks [][]T
	chunk  int // index of the chunk being filled
	used   int // elements used in that chunk
	size   int // capacity of a regular chunk
}

type arenaMark struct {
	chunk, used int
}

func newChunkArena[T any](size int) *chunkArena[T] {
	if size < 64 {
		size = 64
	}
	return &chunkArena[T]{
		chunks: [][]T{make([]T, size)},
		size:   size,
	}
}

// alloc returns a slice of n elements. Its contents are whatever the previous user
// of that space left behind.
func (a *chunkArena[T]) alloc(n int) []T {
	if n == 0 {
		return nil
	}
	if n > a.size {
		// Larger than a chunk. Rare enough not to be worth complicating the mark
		// discipline for, so the heap takes it.
		return make([]T, n)
	}
	if a.used+n > len(a.chunks[a.chunk]) {
		a.chunk++
		a.used = 0
		if a.chunk == len(a.chunks) {
			a.chunks = append(a.chunks, make([]T, a.size))
		}
	}
	out := a.chunks[a.chunk][a.used : a.used+n : a.used+n]
	a.used += n
	return out
}

func (a *chunkArena[T]) mark() arenaMark { return arenaMark{a.chunk, a.used} }

func (a *chunkArena[T]) release(m arenaMark) {
	a.chunk, a.used = m.chunk, m.used
}
