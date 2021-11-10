package dimacs

import (
	"fmt"
	"testing"
)

func TestDimacs01(t *testing.T) {
	cs, _ := ParseDimacs([]byte(`
	p cnf 5 6
	4 5 6 3 0
	-1 2 1 0
	`))
	for _, x := range cs {
		fmt.Println(x)
	}
}
