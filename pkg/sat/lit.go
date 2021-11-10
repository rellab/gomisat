package sat

import (
	"errors"
	_ "fmt"
	"log"
	"strconv"
)

// LBool

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

func NewLBool(x bool) LBool {
	if x {
		return LTrue
	} else {
		return LFalse
	}
}

func (x LBool) Not() LBool {
	return LBool(x ^ 1)
}

// Variables for SAT

var (
	VarUndef    Var   = Var(-1)
	LitUndef    Lit   = Lit(-1)
	ErrLitError error = errors.New("Literal error")
)

type Var int

func (v Var) String() string {
	if v == VarUndef {
		return "Undef"
	} else {
		return strconv.Itoa(int(v) + 1)
	}
}

// Literal for SAT

type Lit int

func (p Lit) String() string {
	if p == LitUndef {
		return "Undef"
	}
	if p.Sign() == false {
		return strconv.Itoa(int(p.Var()) + 1)
	} else {
		return "~" + strconv.Itoa(int(p.Var())+1)
	}
}

func MkLit(v Var, sign bool) Lit {
	if sign == false {
		return Lit(v + v)
	} else {
		return Lit(v + v + 1)
	}
}

func (p Lit) Sign() bool {
	return p&1 == 1
}

func (p Lit) Var() Var {
	return Var(p >> 1)
}

func (p Lit) Not() Lit {
	return p ^ 1
}
