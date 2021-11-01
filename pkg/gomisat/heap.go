package gomisat

import (
	"fmt"
)

type VarHeap struct {
	heap     []Var
	indicies map[Var]int
	lt       func(x, y Var) bool
}

func NewVarHeap(lt func(x, y Var) bool) *VarHeap {
	return &VarHeap{
		heap:     make([]Var, 0),
		indicies: make(map[Var]int),
		lt:       lt,
	}
}

func (h *VarHeap) String() string {
	return fmt.Sprint("Heap ", h.heap, " Indicies ", h.indicies)
}

func left(i int) int {
	return i*2 + 1
}

func right(i int) int {
	return (i + 1) * 2
}

func parent(i int) int {
	return (i - 1) >> 1
}

func (h *VarHeap) percolateUp(i int) {
	x := h.heap[i]
	p := parent(i)
	for i > 0 && h.lt(x, h.heap[p]) {
		h.heap[i] = h.heap[p]
		h.indicies[h.heap[p]] = i
		i, p = p, parent(p)
	}
	h.heap[i] = x
	h.indicies[x] = i
}

func (h *VarHeap) percolateDown(i int) {
	x := h.heap[i]
	for left(i) < len(h.heap) {
		var child int
		if right(i) < len(h.heap) && h.lt(h.heap[right(i)], h.heap[left(i)]) {
			child = right(i)
		} else {
			child = left(i)
		}
		if !h.lt(h.heap[child], x) {
			break
		}
		h.heap[i] = h.heap[child]
		h.indicies[h.heap[i]] = i
		i = child
	}
	h.heap[i] = x
	h.indicies[x] = i
}

func (h *VarHeap) IsEmpty() bool {
	return len(h.heap) == 0
}

func (h *VarHeap) InHeap(v Var) bool {
	_, ok := h.indicies[v]
	return ok
}

func (h *VarHeap) Decrease(v Var) {
	h.percolateUp(h.indicies[v])
}

func (h *VarHeap) Increase(v Var) {
	h.percolateDown(h.indicies[v])
}

func (h *VarHeap) Insert(v Var) {
	if _, ok := h.indicies[v]; ok == false {
		i := len(h.heap)
		h.heap = append(h.heap, v)
		h.indicies[v] = i
		h.percolateUp(i)
	}
}

func (h *VarHeap) RemoveMin() Var {
	x := h.heap[0]
	h.heap[0] = h.heap[len(h.heap)-1]
	h.indicies[h.heap[0]] = 0
	h.heap = h.heap[:len(h.heap)-1]
	delete(h.indicies, x)
	if len(h.heap) > 1 {
		h.percolateDown(0)
	}
	return x
}
