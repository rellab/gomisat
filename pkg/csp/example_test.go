package csp_test

import (
	"fmt"

	"github.com/rellab/gomisat/pkg/csp"
)

// A small integer problem, solved and counted.
func Example() {
	m := csp.New()
	x := m.IntVarRange(0, 10)
	y := m.IntVarRange(0, 10)

	// 2x + 3y <= 12 and x >= y
	m.Add(csp.LeZero(csp.NewSum(map[*csp.IntVar]int{x: 2, y: 3}, -12)))
	m.Add(csp.GeZero(csp.NewSum(map[*csp.IntVar]int{x: 1, y: -1}, 0)))

	sol, ok := m.Solve()
	if ok == false {
		fmt.Println("no solution")
		return
	}
	// Which solution comes back depends on the search, so the example checks it
	// rather than printing it.
	xv, yv := sol.Int(x), sol.Int(y)
	fmt.Printf("the solution satisfies both constraints: %v\n", 2*xv+3*yv <= 12 && xv >= yv)

	count, _ := m.Count()
	fmt.Printf("solutions: %v\n", count)

	// Output:
	// the solution satisfies both constraints: true
	// solutions: 13
}

// Boolean variables and a count.
func Example_boolean() {
	m := csp.New()
	a, b, c := m.NewBool(), m.NewBool(), m.NewBool()

	// At least one of a, b; and c whenever a.
	m.Add(csp.Or(a, b))
	m.Add(csp.Imp(a, c))

	count, _ := m.Count()
	fmt.Printf("solutions: %v\n", count)

	// Output:
	// solutions: 4
}
