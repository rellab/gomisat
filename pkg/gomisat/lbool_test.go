package gomisat

import (
	"fmt"
	"testing"
)

func TestLBool01(t *testing.T) {
	x := []LBool{LTrue, LFalse, LUndef}
	y := []LBool{LTrue, LFalse, LUndef}
	for _, i := range x {
		for _, j := range y {
			fmt.Println(i, " & ", j, " = ", LAnd(i, j))
		}
	}
}

func TestLBool02(t *testing.T) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			x := LBool(i)
			y := LBool(j)
			fmt.Println(i, " == ", j, " = ", LEq(x, y))
		}
	}
}
