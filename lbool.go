package gomisat

import (
	_ "fmt"
	"log"
)

type LBool uint8

var (
	LTrue  LBool
	LFalse LBool
	LUndef LBool
)

func init() {
	LTrue = LBool(0)
	LFalse = LBool(1)
	LUndef = LBool(2)
}

func (b LBool) String() string {
	switch {
	case b == LTrue:
		return "T"
	case b == LFalse:
		return "F"
	case b == LUndef:
		return "U"
	case b == 3:
		return "X"
	default:
		log.Fatal("LBool takes an invalid value:", uint8(b))
		return "-"
	}
}

func Btolb(x bool) LBool {
	if x {
		return LTrue
	} else {
		return LFalse
	}
}

func LEq(x, y LBool) bool {
	if (x&2)&(y&2) == 0 {
		return x == y
	} else {
		return true
	}
}

func LNeq(x, y LBool) bool {
	return !LEq(x, y)
}

func LFlip(x LBool, y bool) LBool {
	if y {
		return LBool(x ^ 1)
	} else {
		return x
	}
}

func LAnd(x, y LBool) LBool {
	sel := (x << 1) | (y << 3)
	v := 0xf7f755f4 >> sel & 0x3
	return LBool(v)
}

func LOr(x, y LBool) LBool {
	sel := (x << 1) | (y << 3)
	v := 0xfcfcf400 >> sel & 0x3
	return LBool(v)
}
