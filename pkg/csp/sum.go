package csp

import (
	"fmt"
)

type Sum struct {
	coef map[*IntVar]int
	b    int
}

func NewSum(coef map[*IntVar]int, b int) *Sum {
	return &Sum{
		coef: coef,
		b:    b,
	}
}

func (s *Sum) String() string {
	str := "["
	for k, v := range s.coef {
		str += fmt.Sprintf("+(%d)*%s", v, k.String())
	}
	str += fmt.Sprintf("+(%d)]", s.b)
	return str
}

func (s *Sum) size() int {
	return len(s.coef)
}

func (s *Sum) applyFunc(other *Sum, f func(int, int) int) *Sum {
	for k, v := range other.coef {
		if tmp := f(s.coef[k], v); tmp != 0 {
			s.coef[k] = tmp
		} else {
			delete(s.coef, k)
		}
	}
	s.b = f(s.b, other.b)
	return s
}

func (s *Sum) neg() *Sum {
	for k, v := range s.coef {
		s.coef[k] = -v
	}
	s.b = -s.b
	return s
}

func (s *Sum) addConst(b int) *Sum {
	s.b += b
	return s
}

func (s *Sum) add(other *Sum) *Sum {
	return s.applyFunc(other, func(x, y int) int {
		return x + y
	})
}

func (s *Sum) sub(other *Sum) *Sum {
	return s.applyFunc(other, func(x, y int) int {
		return x - y
	})
}

// func (s *Sum) MulConst(a int) *Sum {
// 	if a != 0 {
// 		for k, v := range s.coef {
// 			s.coef[k] = a * v
// 		}
// 		s.b *= a
// 	} else {
// 		s.coef = make(map[*IntVar]int)
// 		s.b = 0
// 	}
// 	return s
// }

func (s *Sum) copy() *Sum {
	coef := make(map[*IntVar]int)
	for k, v := range s.coef {
		coef[k] = v
	}
	return &Sum{
		coef: coef,
		b:    s.b,
	}
}
