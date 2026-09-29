package csp

import (
	// "log"
	"strconv"
)

// BoolVar

type BoolVar struct {
	id  int
	aux bool
}

func newBoolVar(id int) *BoolVar {
	return &BoolVar{
		id:  id,
		aux: false,
	}
}

func newAuxBoolVar(id int) *BoolVar {
	return &BoolVar{
		id:  id,
		aux: true,
	}
}

func (x BoolVar) String() string {
	var s string
	if x.aux == false {
		s = "b" + strconv.Itoa(int(x.id))
	} else {
		s = "ab" + strconv.Itoa(int(x.id))
	}
	return s
}

// IntVar

type IntVar struct {
	id     int
	domain DomainSet
	aux    bool
}

func newIntVar(id int, domain DomainSet) *IntVar {
	return &IntVar{
		id:     id,
		domain: domain,
		aux:    false,
	}
}

func newAuxIntVar(id int, domain DomainSet) *IntVar {
	return &IntVar{
		id:     id,
		domain: domain,
		aux:    true,
	}
}

func (x IntVar) String() string {
	var s string
	if x.aux == false {
		s = "x" + strconv.Itoa(x.id)
	} else {
		s = "ax" + strconv.Itoa(x.id)
	}
	return s
}
