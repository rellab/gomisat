package csp

import (
	// 	"log"
	"sort"
)

var (
	ErrDomainLower int = -1
	ErrDomainUpper int = -2
)

type DomainSet []int

// func (d DomainSet) copy() DomainSet {
// 	x := make([]int, len(d))
// 	copy(x, d)
// 	return DomainSet(x)
// }

func (d DomainSet) size() int {
	return len(d)
}

// func (d DomainSet) Lb() int {
// 	return d.x[0]
// }

// func (d DomainSet) Ub() int {
// 	return d.x[len(d.x)-1]
// }

func (d DomainSet) contains(value int) (int, bool) {
	return contains(d, value)
}

// func (d DomainSet) CrossApplyFunc(other DomainSet, f func(int, int) int) {
// 	crossApply(d, other, f)
// }

// The function to give us the position of a given value with a binary search
// The second return value indicates whehter a value cotaints or not.
// If the second value is false, the first value indicates the following conditions
// for a given value x:
//
//	x < lb  -> -1
//	lb <= x <= ub -> 0
//	ub < x -> 1
func contains(s []int, x int) (int, bool) {
	l := 0
	u := len(s) - 1
	switch {
	case x < s[l]:
		return ErrDomainLower, false // x is less than the lower bound of domain
	case s[u] <= x:
		return ErrDomainUpper, false // x is greather than or equal to the upper bound of domain
	case x == s[l]:
		return l, true
		// case s[u] == x:
		// 	return u, true
	}
	for u-l > 1 {
		m := (l + u) / 2
		if x == s[m] {
			return m, true
		} else if x < s[m] {
			u = m
		} else {
			l = m
		}
	}
	return l, false // x is in the range beween lower and upper bounds, but the domain does not have the same value
}

func removerRedundant(x []int) []int {
	if len(x) == 0 {
		return x
	}
	sort.Slice(x, func(i, j int) bool {
		return x[i] < x[j]
	})
	j := 1
	prev := x[0]
	for _, v := range x[1:] {
		if prev < v {
			x[j] = v
			j++
			prev = v
		}
	}
	return x[:j]
}

func crossApply(x, y []int, f func(int, int) int) []int {
	tmp := make([]int, 0, len(x)*len(y))
	for _, v1 := range x {
		for _, v2 := range y {
			tmp = append(tmp, f(v1, v2))
		}
	}
	return removerRedundant(tmp)
}

// func (d *DomainSet) Bound(lb, ub int) {
// 	d.x = bound(d.x, lb, ub)
// }

// func (d *DomainSet) Cap(other DomainSet) {
// 	d.x = cap(d.x, other.x)
// }

// func (d *DomainSet) Cup(other DomainSet) {
// 	d.x = cup(d.x, other.x)
// }

// func (d *DomainSet) Neg() {
// 	d.x = neg(d.x)
// }

// func (d *DomainSet) Add(other DomainSet) {
// 	d.x = crossApply(d.x, other.x, func(x, y int) int {
// 		return x + y
// 	})
// }

// func (d *DomainSet) Sub(other DomainSet) {
// 	d.x = crossApply(d.x, other.x, func(x, y int) int {
// 		return x - y
// 	})
// }

// func (d *DomainSet) Mul(other DomainSet) {
// 	d.x = crossApply(d.x, other.x, func(x, y int) int {
// 		return x * y
// 	})
// }

// func (d *DomainSet) FloorDiv(a int) {
// 	d.x = mapApply(d.x, func(x int) int {
// 		return floorDiv(x, a)
// 	})
// }

// func (d *DomainSet) CeilDiv(a int) {
// 	d.x = mapApply(d.x, func(x int) int {
// 		return ceilDiv(x, a)
// 	})
// }

// func filter(x []int, f func(int) bool) []int {
// 	j := 0
// 	for _, y := range x {
// 		if f(y) {
// 			x[j] = y
// 			j++
// 		}
// 	}
// 	return x[:j]
// }

// func bound(x []int, lb, ub int) []int {
// 	for i, v := range x {
// 		if v < lb {
// 			continue
// 		}
// 		x = x[i:]
// 		break
// 	}
// 	for i, v := range x {
// 		if v < ub {
// 			continue
// 		}
// 		x = x[:i]
// 		break
// 	}
// 	return x
// }

// // usage:
// //   x = cap(x, y) // update x
// func cap(x, y []int) []int {
// 	i := 0
// 	j := 0
// 	k := 0
// 	for i != len(x) && j != len(y) {
// 		if x[i] == y[j] {
// 			x[k] = x[i]
// 			i++
// 			j++
// 			k++
// 		} else if x[i] < y[j] {
// 			i++
// 		} else {
// 			j++
// 		}
// 	}
// 	return x[:k]
// }

// // usage:
// //   x = cup(x, y) // update x
// func cup(x, y []int) []int {
// 	z := make([]int, 0, len(x)+len(y))
// 	j := 0
// 	for _, v := range x {
// 		for v > y[j] {
// 			z = append(z, y[j])
// 			j++
// 		}
// 		z = append(z, v)
// 		if v == y[j] {
// 			j++
// 		}
// 	}
// 	return append(z, y[j:]...)
// }

// func neg(x []int) []int {
// 	for i := 0; i < len(x)/2; i++ {
// 		x[i], x[len(x)-1-i] = -x[len(x)-1-i], -x[i]
// 	}
// 	if len(x)%2 == 1 {
// 		x[len(x)/2] = -x[len(x)/2]
// 	}
// 	return x
// }

// func mapApply(x []int, f func(int) int) []int {
// 	for i, v := range x {
// 		x[i] = f(v)
// 	}
// 	return removerRedundant(x)
// }

func floorDiv(b, a int) int {
	if a > 0 {
		if b >= 0 {
			return b / a
		} else {
			return (b - a + 1) / a
		}
	} else if b >= 0 {
		return (b - a - 1) / a
	} else {
		return b / a
	}
}

func ceilDiv(b, a int) int {
	if a > 0 {
		if b >= 0 {
			return (b + a - 1) / a
		} else {
			return b / a
		}
	} else if b >= 0 {
		return b / a
	} else {
		return (b + a + 1) / a
	}
}

// type DomainInterval struct {
// 	lower int
// 	upper int
// }

// func (d *DomainInterval) Size() int {
// 	return d.upper - d.lower + 1
// }

// func (d *DomainInterval) Get(i int) int {
// 	return d.lower + i
// }

// func (d *DomainInterval) Lb() int {
// 	return d.lower
// }

// func (d *DomainInterval) Ub() int {
// 	return d.upper
// }

// func (d *DomainInterval) Contains(v int) bool {
// 	return d.lower <= v && v <= d.upper
// }

// func (d *DomainInterval) Bound(lb, ub int) {
// 	if d.lower < lb {
// 		d.lower = lb
// 	}
// 	if ub < d.upper {
// 		d.upper = ub
// 	}
// }
