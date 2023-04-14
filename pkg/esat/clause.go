package esat

import (
	_ "fmt"
	_ "log"
	"strconv"
	"strings"
)

type Clause struct {
	lits []int
	eval uint
}

func (c *Clause) String() string {
	s := make([]string, 0, len(c.lits))
	for _, x := range c.lits {
		s = append(s, strconv.Itoa(x))
	}
	return "[" + strings.Join(s, ",") + "] (" + strconv.Itoa(int(c.eval)) + ")"
}

type ESat struct {
	vars    []int
	clauses []*Clause
	lits    map[int][]*Clause
}

func MkESat(cnf [][]int) *ESat {
	es := &ESat{
		vars:    nil,
		clauses: make([]*Clause, 0),
		lits:    make(map[int][]*Clause),
	}
	maxL := 1
	varsTmp := make(map[int]struct{})
	for _, c := range cnf {
		if len(c) > maxL {
			maxL = len(c)
		}
		cl := &Clause{
			lits: c,
			eval: 0,
		}
		es.clauses = append(es.clauses, cl)
		for _, p := range c {
			if v, ok := es.lits[p]; ok == false {
				es.lits[p] = []*Clause{cl}
			} else {
				v = append(v, cl)
			}
			if p > 0 {
				if _, ok := varsTmp[p]; ok == false {
					varsTmp[p] = struct{}{}
				}
			} else {
				if _, ok := varsTmp[-p]; ok == false {
					varsTmp[-p] = struct{}{}
				}
			}
		}
	}
	es.vars = make([]int, 0, len(varsTmp))
	for k, _ := range varsTmp {
		es.vars = append(es.vars, k)
	}
	// eval
	for _, c := range es.clauses {
		c.eval = 0x1 << (maxL - len(c.lits))
	}
	return es
}

func (es *ESat) MkAssign(vars []int) map[int]bool {
	ans := make(map[int]bool)
	for _, x := range vars {
		evalplus := uint(0)
		evalminus := uint(0)
		for _, c := range es.lits[x] {
			evalplus += c.eval
		}
		for _, c := range es.lits[-x] {
			evalminus += c.eval
		}
		if evalplus > evalminus {
			ans[x] = true
			for _, c := range es.lits[x] {
				c.eval = 0
			}
			for _, c := range es.lits[-x] {
				c.eval = c.eval << 1
			}
		} else {
			ans[x] = false
			for _, c := range es.lits[-x] {
				c.eval = 0
			}
			for _, c := range es.lits[x] {
				c.eval = c.eval << 1
			}
		}
	}
	return ans
}

func (es *ESat) MkAssign3(vars []int) map[int]bool {
	ans := make(map[int]bool)
	for _, x := range vars {
		t, g := es.Gain(x)
		switch {
		case g != 0 && t == true:
			ans[x] = true
			es.Update(x)
		case g != 0 && t == false:
			ans[x] = false
			es.Update(-x)
		case g == 0:
			ans[x] = true
			es.Update(x)
		}
	}
	return ans
}

func (es *ESat) Gain(x int) (bool, uint) {
	evalplus := uint(0)
	evalminus := uint(0)
	for _, c := range es.lits[x] {
		evalplus += c.eval
	}
	for _, c := range es.lits[-x] {
		evalminus += c.eval
	}
	if evalplus > evalminus {
		return true, evalplus - evalminus
	} else {
		return false, evalminus - evalplus
	}
}

func (es *ESat) Update(p int) {
	for _, c := range es.lits[p] {
		c.eval = 0
	}
	for _, c := range es.lits[-p] {
		c.eval = c.eval << 1
	}
}

func (es *ESat) MkAssign2(vars map[int]struct{}) map[int]bool {
	ans := make(map[int]bool)
	for len(vars) > 0 {
		maxx := 0
		maxt := false
		maxg := uint(0)
		first := true
		for x, _ := range vars {
			t, g := es.Gain(x)
			if g > maxg || first == true {
				maxx = x
				maxt = t
				maxg = g
				first = false
			}
		}
		// log.Println("determine", maxx, maxt)
		ans[maxx] = maxt
		if maxt {
			es.Update(maxx)
		} else {
			es.Update(-maxx)
		}
		delete(vars, maxx)
	}
	return ans
}
