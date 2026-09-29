package csp

import (
	"fmt"
	"testing"
)

// TestEncodingIsDeterministic requires the same model to encode the same way every
// time.
//
// It did not. The variables of a linear sum come out of a map, whose iteration
// order Go randomises, and the ordering used to decompose the sum left many ties
// unbroken -- equal domains, coefficients from a small set. Ten encodings of the
// same twenty-variable model produced nine different formulas, ranging from 20 190
// to 35 060 clauses. Nothing about the answers was wrong, but no measurement over
// this package would have meant anything.
func TestEncodingIsDeterministic(t *testing.T) {
	cases := map[string]func(m *Model){
		"twenty variables over 0..9, one sum": func(m *Model) {
			n := 20
			x := make([]*IntVar, n)
			sum := make(map[*IntVar]int, n)
			for i := range x {
				x[i] = m.IntVarRange(0, 9)
				sum[x[i]] = 1 + i%3
			}
			m.Add(LeZero(NewSum(sum, -(n * 3))))
		},
		"mixed domains and coefficients": func(m *Model) {
			sum := make(map[*IntVar]int, 8)
			for i := 0; i < 8; i++ {
				v := m.IntVarRange(0, 2+i%4)
				sum[v] = 1 + (i * 7 % 5)
			}
			m.Add(GeZero(NewSum(sum, -6)))
		},
		"booleans and integers together": func(m *Model) {
			a, b := m.NewBool(), m.NewBool()
			sum := make(map[*IntVar]int, 4)
			for i := 0; i < 4; i++ {
				sum[m.IntVarRange(0, 5)] = i + 1
			}
			m.Add(Or(And(a, b), LeZero(NewSum(sum, -7))))
		},
	}

	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			encode := func() string {
				m := New()
				build(m)
				clauses, numVars := m.Build()
				return fmt.Sprint(numVars, clauses)
			}
			first := encode()
			for i := 1; i < 12; i++ {
				if got := encode(); got != first {
					t.Fatalf("encoding %d differs from the first (%d bytes against %d)",
						i, len(got), len(first))
				}
			}
		})
	}
}
