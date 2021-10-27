package gomisat

import (
	"errors"
	_ "fmt"
	_ "log"
	"strconv"
)

var (
	VarUndef    Var   = Var(-1)
	LitUndef    Lit   = Lit(-1)
	ErrLitError error = errors.New("Literal error")
)

type Var int64

type Lit int64

func (p Lit) String() string {
	if p == LitUndef {
		return "Undef"
	}
	if p.Sign() == false {
		return strconv.FormatInt(int64(p.Var()), 10)
	} else {
		return strconv.FormatInt(-int64(p.Var()), 10)
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
	if p&1 == 1 {
		return true
	} else {
		return false
	}
}

func (p Lit) Var() Var {
	return Var(p >> 1)
}

func (p Lit) Not() Lit {
	return Lit(p ^ 1)
}

func LitFlip(p Lit, b bool) Lit {
	if b {
		return Lit(p ^ 1)
	} else {
		return p
	}
}
