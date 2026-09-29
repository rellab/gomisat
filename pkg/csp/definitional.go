package csp

import "log"

// Definitional conversion to CNF.
//
// The conversion the package started with is the polarity-optimised one: an
// auxiliary variable p standing for a subformula gets only the implication the
// satisfiability question needs, p -> subformula. That is correct for solving and
// it is smaller, but it is wrong for counting. If the subformula holds and nothing
// forces p, then p may be either value and every model is counted twice.
//
// It does not always go wrong, which is what makes it dangerous. With two branches
// of a disjunction the surrounding clauses happen to pin the auxiliaries down; with
// three, a formula with 37 solutions was encoded into 61 models. Measured, not
// argued -- see TestCountPreservation.
//
// So this file defines each auxiliary in both directions, p <-> subformula, which
// makes each solution of the original extend to exactly one model of the CNF. It
// costs more clauses and it is the default, because an encoding that is wrong for
// counting is a worse trade than one that is larger.

type definer struct {
	cnf []Clause
	aux []*BoolVar
}

// definitional converts one constraint into clauses whose auxiliary variables are
// all determined by the original variables.
func definitional(c Constraint, cnf []Clause, auxvars []*BoolVar) ([]Clause, []*BoolVar) {
	d := &definer{cnf: cnf, aux: auxvars}
	d.assert(c)
	d.splitNonSimple()
	return d.cnf, d.aux
}

// assert states that a constraint holds. A constraint being asserted needs no
// variable of its own: a conjunction is asserted by asserting its arguments, and a
// disjunction becomes one clause. Only the arguments that are not literals need
// defining, and those still need both directions.
func (d *definer) assert(c Constraint) {
	if x, ok := c.(*Operator); ok {
		switch x.op {
		case opAnd:
			for _, arg := range x.args {
				d.assert(arg)
			}
			return
		case opOr:
			clause := make(Clause, 0, len(x.args))
			for _, arg := range x.args {
				clause = append(clause, d.define(arg))
			}
			d.cnf = append(d.cnf, clause)
			return
		}
	}
	d.cnf = append(d.cnf, Clause{d.define(c)})
}

func (d *definer) fresh() *BoolVar {
	p := newAuxBoolVar(len(d.aux))
	d.aux = append(d.aux, p)
	return p
}

// define returns a literal that holds exactly when c does, adding the clauses that
// make that so.
func (d *definer) define(c Constraint) Literal {
	switch x := c.(type) {
	case *BoolVar:
		return x
	case *BoolNot:
		return x
	case *Comparator:
		return x
	case *Operator:
		lits := make([]Literal, 0, len(x.args))
		for _, arg := range x.args {
			lits = append(lits, d.define(arg))
		}
		p := d.fresh()
		switch x.op {
		case opAnd:
			// p -> every argument, and every argument -> p.
			for _, l := range lits {
				d.cnf = append(d.cnf, Clause{l, negate(p)})
			}
			all := make(Clause, 0, len(lits)+1)
			for _, l := range lits {
				all = append(all, negate(l))
			}
			d.cnf = append(d.cnf, append(all, p))
		case opOr:
			// p -> some argument, and every argument -> p.
			some := make(Clause, 0, len(lits)+1)
			some = append(some, negate(p))
			some = append(some, lits...)
			d.cnf = append(d.cnf, some)
			for _, l := range lits {
				d.cnf = append(d.cnf, Clause{negate(l), p})
			}
		default:
			log.Fatal("csp: operator does not exist")
		}
		return p
	}
	log.Fatalf("csp: %T cannot be a literal", c)
	return nil
}

// negate returns the literal that holds exactly when this one does not.
func negate(l Literal) Literal {
	switch x := l.(type) {
	case *BoolVar:
		return &BoolNot{x}
	case *BoolNot:
		return x.b
	case *Comparator:
		// A comparator's negation is another comparator, but it has to be brought
		// back to the "<= 0" form the encoder accepts.
		n, ok := x.Not().ToLeZero().(*Comparator)
		if ok == false {
			log.Fatal("csp: the negation of a comparator is not a comparator")
		}
		return n
	}
	log.Fatalf("csp: %T cannot be negated", l)
	return nil
}

// splitNonSimple makes sure no clause holds more than one literal that the encoder
// cannot put into a clause on its own -- a comparator over a sum of several
// variables. Each such literal is replaced by an auxiliary that is again defined in
// both directions, so the count is still preserved.
func (d *definer) splitNonSimple() {
	out := make([]Clause, 0, len(d.cnf))
	for _, clause := range d.cnf {
		if isSimple(clause) {
			out = append(out, clause)
			continue
		}
		replaced := make(Clause, 0, len(clause))
		for _, lit := range clause {
			if lit.isSimple() {
				replaced = append(replaced, lit)
				continue
			}
			p := d.fresh()
			out = append(out, Clause{lit, negate(p)}, Clause{negate(lit), p})
			replaced = append(replaced, p)
		}
		out = append(out, replaced)
	}
	d.cnf = out
}
