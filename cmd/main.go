// Command gomisat solves a single DIMACS CNF file.
//
// The exit status follows the SAT competition convention, so that the binary can
// be driven from a script: 10 for satisfiable, 20 for unsatisfiable, 0 when the
// answer is unknown and 1 on error.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rellab/gomisat/pkg/gomisat"
)

func main() {
	model := flag.Bool("model", false, "print the satisfying assignment as a DIMACS v-line")
	timeout := flag.Float64("timeout", 0, "wall clock limit in seconds (0 = no limit)")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: gomisat [flags] <file.cnf>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	buf, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "gomisat:", err)
		os.Exit(1)
	}
	clauses, err := gomisat.ParseDimacs(buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gomisat:", err)
		os.Exit(1)
	}

	s := gomisat.NewSolver()
	options := gomisat.DefaultSolverOptions()
	for _, c := range clauses {
		s.AddClauseFromCode(c, options)
	}
	if *timeout > 0 {
		timer := time.AfterFunc(time.Duration(*timeout*float64(time.Second)), s.Interrupt)
		defer timer.Stop()
	}

	start := time.Now()
	status := s.Solve(options)
	elapsed := time.Since(start)

	fmt.Printf("c vars %d clauses %d\n", s.NumVars(), len(clauses))
	fmt.Printf("c conflicts %d propagations %d decisions %d restarts %d\n",
		s.Conflicts, s.Propagations, s.Decisions, s.Starts)
	fmt.Printf("c time %.6f s\n", elapsed.Seconds())

	switch status {
	case gomisat.LTrue:
		fmt.Println("s SATISFIABLE")
		if *model {
			printModel(s)
		}
		os.Exit(10)
	case gomisat.LFalse:
		fmt.Println("s UNSATISFIABLE")
		os.Exit(20)
	default:
		fmt.Println("s UNKNOWN")
		os.Exit(0)
	}
}

func printModel(s *gomisat.Solver) {
	fmt.Print("v")
	for v := 0; v < s.NumVars(); v++ {
		code := v + 1
		if s.ModelValue(gomisat.Var(v)) == gomisat.LFalse {
			code = -code
		}
		fmt.Printf(" %d", code)
	}
	fmt.Println(" 0")
}
