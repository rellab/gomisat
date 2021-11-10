package sat

import (
	"fmt"
	"testing"
)

func TestLBool02(t *testing.T) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			x := LBool(i)
			y := LBool(j)
			fmt.Println(i, " == ", j, " = ", x == y)
		}
	}
}

func TestVar01(t *testing.T) {
	x := VarUndef
	fmt.Println(x)
}

func TestLit01(t *testing.T) {
	x := MkLit(Var(1), false)
	fmt.Println(x)
	y := MkLit(Var(1), true)
	fmt.Println(y)
}

func TestLit02(t *testing.T) {
	x := MkLit(Var(1), false)
	fmt.Println(x)
	fmt.Println(x.Not())
}
