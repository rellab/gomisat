package gomisat

import (
	"math/rand"
	"sort"
	"testing"
)

// newTestHeap builds a heap over n variables ordered by a given key, with the
// index array sized as the solver sizes it.
func newTestHeap(n int, key []float64) *VarHeap {
	h := NewVarHeap(func(x, y Var) bool { return key[x] > key[y] })
	h.indicies = make([]int, n)
	for i := range h.indicies {
		h.indicies[i] = UndefIndex
	}
	return h
}

// TestHeapReinsertAfterFullDrain pins down the case that made the solver report
// models with unassigned variables: removing the last element must leave the
// variable marked as absent, otherwise Insert ignores it forever.
func TestHeapReinsertAfterFullDrain(t *testing.T) {
	key := []float64{0.3, 0.1, 0.2}
	h := newTestHeap(len(key), key)
	for v := range key {
		h.Insert(Var(v))
	}
	for h.IsEmpty() == false {
		v := h.RemoveMin()
		if h.InHeap(v) {
			t.Fatalf("%v is still reported as being in the heap right after removal", v)
		}
	}
	for v := range key {
		if h.InHeap(Var(v)) {
			t.Errorf("%v is reported as being in the drained heap", Var(v))
		}
		h.Insert(Var(v))
		if h.InHeap(Var(v)) == false {
			t.Errorf("%v was not inserted back", Var(v))
		}
	}
	if got := len(h.heap); got != len(key) {
		t.Fatalf("heap holds %d variables after reinsertion, want %d", got, len(key))
	}
}

// TestHeapOrder compares the removal order against sorting, over random keys and
// interleaved insertions and removals.
func TestHeapOrder(t *testing.T) {
	rng := rand.New(rand.NewSource(20260929))
	const n = 64
	for round := 0; round < 200; round++ {
		key := make([]float64, n)
		for i := range key {
			key[i] = rng.Float64()
		}
		h := newTestHeap(n, key)

		present := make(map[Var]bool, n)
		for i := 0; i < n; i++ {
			if rng.Intn(2) == 0 {
				h.Insert(Var(i))
				present[Var(i)] = true
			}
		}
		// Drain, checking that the keys come out in decreasing order and that
		// every inserted variable comes out exactly once.
		want := make([]Var, 0, len(present))
		for v := range present {
			want = append(want, v)
		}
		sort.Slice(want, func(i, j int) bool {
			if key[want[i]] == key[want[j]] {
				return want[i] < want[j]
			}
			return key[want[i]] > key[want[j]]
		})
		got := make([]Var, 0, len(present))
		for h.IsEmpty() == false {
			got = append(got, h.RemoveMin())
		}
		if len(got) != len(want) {
			t.Fatalf("round %d: removed %d variables, inserted %d", round, len(got), len(want))
		}
		for i := range got {
			if key[got[i]] != key[want[i]] {
				t.Fatalf("round %d: position %d has key %v, want %v", round, i, key[got[i]], key[want[i]])
			}
		}
	}
}
