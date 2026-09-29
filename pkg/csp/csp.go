package csp

import _ "fmt"

type Model struct {
	varId           int
	boolVars        []*BoolVar
	auxBoolVars     []*BoolVar
	intVars         []*IntVar
	auxIntVars      []*IntVar
	constraints     []Constraint
	cnf             []Clause
	cnfDone         []bool   // indicator whether the simplify is done or note
	cnfStart        []int    // the index to start Clause for the corresponding Model constraint
	cnfStartAuxBool []int    // the index to start auxbool for the corresponding Model constraint
	tmpCNF          []Clause // this is tempolary used in simplify

	baseCode map[int]int // SAT code base
	sat      [][]int     // SAT code

	// Definitional selects the CNF conversion. With it, every auxiliary variable
	// is defined in both directions, so each solution of the model corresponds to
	// exactly one model of the CNF and the two can be counted against each other.
	// Without it the conversion is the polarity-optimised one, which is smaller and
	// correct for solving only. See definitional.go.
	Definitional bool

	built    bool
	numCodes int
	unsat    bool
}

func New() *Model {
	return &Model{
		varId:           0,
		boolVars:        make([]*BoolVar, 0),
		auxBoolVars:     make([]*BoolVar, 0),
		intVars:         make([]*IntVar, 0),
		auxIntVars:      make([]*IntVar, 0),
		constraints:     make([]Constraint, 0),
		cnf:             make([]Clause, 0),
		cnfDone:         make([]bool, 0),
		cnfStart:        make([]int, 0),
		cnfStartAuxBool: make([]int, 0),
		tmpCNF:          make([]Clause, 0),

		baseCode: make(map[int]int),
		sat:      make([][]int, 0),

		Definitional: true,
	}
}

func (c *Model) NewBool() *BoolVar {
	v := newBoolVar(c.varId)
	c.boolVars = append(c.boolVars, v)
	c.varId++
	return v
}

func (c *Model) IntVarRange(lb, ub int) *IntVar {
	d := make([]int, ub-lb+1)
	for i, _ := range d {
		d[i] = lb + i
	}
	v := newIntVar(c.varId, d)
	c.intVars = append(c.intVars, v)
	c.varId++
	return v
}

// AddConstraint
// The method to add a Model constraint; Comparators, Operators and Bool
// The argument `decomp` indicates whether the Model constraint is decomposed to up to three terms or not
// for all the linear functions in the Model constraint.
func (c *Model) AddConstraint(x Constraint, decomp bool) {
	if decomp {
		var cs Constraint
		start := len(c.auxIntVars)
		cs, c.auxIntVars = x.Decomp(c.auxIntVars)
		// rewrite id
		for k := start; k < len(c.auxIntVars); k++ {
			c.auxIntVars[k].id = c.varId
			c.varId++
		}
		c.constraints = append(c.constraints, cs.ToLeZero())
		c.cnfDone = append(c.cnfDone, false)
		c.cnfStart = append(c.cnfStart, 0)
		c.cnfStartAuxBool = append(c.cnfStartAuxBool, 0)
	} else {
		c.constraints = append(c.constraints, x.ToLeZero())
		// The bookkeeping has to grow here too. It did not, so CNF panicked on any
		// constraint added without decomposition; every test in the package this
		// came from passed decomp = true, which is why it went unnoticed.
		c.cnfDone = append(c.cnfDone, false)
		c.cnfStart = append(c.cnfStart, 0)
		c.cnfStartAuxBool = append(c.cnfStartAuxBool, 0)
	}
}

// Add adds a constraint, decomposing long linear sums into auxiliary variables.
// It is AddConstraint with the usual choice made.
func (c *Model) Add(x Constraint) {
	c.AddConstraint(x, true)
}

// To save the current states (constraints and codes)
func (c *Model) Save() int {
	// TODO: should be implemented
	return 0
}

// To load the status
func (c *Model) Load(num int) {
	// TODO: should be implemented
}

func (c *Model) CNF() {
	for i, cs := range c.constraints {
		if c.cnfDone[i] == false {
			c.cnfStart[i] = len(c.cnf)
			c.cnfStartAuxBool[i] = len(c.auxBoolVars)
			if c.Definitional {
				c.cnf, c.auxBoolVars = definitional(cs, c.cnf, c.auxBoolVars)
			} else {
				c.cnf, c.auxBoolVars = simplify(cs, c.cnf, c.auxBoolVars, c.tmpCNF)
			}
			// rewrite id
			for k := c.cnfStartAuxBool[i]; k < len(c.auxBoolVars); k++ {
				c.auxBoolVars[k].id = c.varId
				c.varId++
			}
			c.cnfDone[i] = true
		}
	}
}

func (c *Model) genBase() {
	code := 1
	for _, v := range c.intVars {
		c.baseCode[v.id] = code
		// for k := 0; k < v.domain.size()-1; k++ {
		// 	log.Println("int", v, "<=", v.domain.x[k], "code", code+k)
		// }
		for k := 0; k < v.domain.size()-2; k++ {
			c.sat = append(c.sat, []int{-(code + k), code + k + 1})
		}
		code += v.domain.size() - 1
	}
	for _, v := range c.auxIntVars {
		c.baseCode[v.id] = code
		// for k := 0; k < v.domain.size()-1; k++ {
		// 	log.Println("aux int", v, "<=", v.domain.x[k], "code", code+k)
		// }
		for k := 0; k < v.domain.size()-2; k++ {
			c.sat = append(c.sat, []int{-(code + k), code + k + 1})
		}
		code += v.domain.size() - 1
	}
	for _, v := range c.boolVars {
		// log.Println("bool", v, "code", code)
		c.baseCode[v.id] = code
		code += 1
	}
	for _, v := range c.auxBoolVars {
		// log.Println("aux bool", v, "code", code)
		c.baseCode[v.id] = code
		code += 1
	}
	c.numCodes = code - 1
}

// Encode turns the clauses into propositional ones. A clause that cannot be
// satisfied at all makes the whole model unsatisfiable, which is recorded rather
// than fatal: a constraint a caller wrote may simply have no solutions, and a
// library that exits the process over it is not usable.
func (c *Model) Encode() {
	for _, x := range c.cnf {
		tmp, ok := Encode(x, c.baseCode)
		if ok == false {
			c.unsat = true
			return
		}
		c.sat = append(c.sat, tmp...)
	}
}

// Unsatisfiable reports whether the encoding found the model to have no solutions.
// It is only meaningful after Build.
func (c *Model) Unsatisfiable() bool {
	c.Build()
	return c.unsat
}

// func (c *Model) GetValue(x *IntVar) int {
// 	// bin search
// 	base := c.baseCode[x.id]
// 	l := x.domain[0]
// 	u := x.domain[x.domain.size()-2]
// 	if value, ok := c.assigns[u]; ok == true && {

// 	}
// 	m := (l + u) / 2
// 	for c.baseCode[assigns
// 	for i := 0; i < v.domain.size(); i++ {
// 		if i == v.domain.size()-1 {
// 			fmt.Println(v, v.domain.x[i])
// 			break
// 		} else if a := assigns[c.baseCode[v.id]+i]; a == true {
// 			fmt.Println(v, v.domain.x[i])
// 			break
// 		}
// 	}

// }
