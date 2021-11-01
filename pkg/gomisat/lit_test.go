package gomisat

import (
	"fmt"
	"testing"
)

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
	fmt.Println(LitFlip(x, true))
	fmt.Println(LitFlip(x, false))
}
