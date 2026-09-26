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
	"runtime/pprof"
	"time"

	"github.com/rellab/gomisat/pkg/gomisat"
)

func main() {
	model := flag.Bool("model", false, "print the satisfying assignment as a DIMACS v-line")
	noLBD := flag.Bool("no-lbd", false, "manage learnt clauses by activity only, as MiniSat does")
	cpuProfile := flag.String("cpuprofile", "", "write a CPU profile to this file")
	restart := flag.String("restart", "", "restart policy: luby, geometric, ema or ema-block")
	count := flag.Bool("count", false, "count the models instead of looking for one")
	noDecompose := flag.Bool("no-decompose", false, "count without component decomposition")
	noCache := flag.Bool("no-cache", false, "count without the component cache")
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
	cnf, err := gomisat.ParseDimacsCNF(buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gomisat:", err)
		os.Exit(1)
	}

	s := gomisat.NewSolver()
	options := gomisat.DefaultSolverOptions()
	options.UseLBD = *noLBD == false
	if *restart != "" {
		options.RestartPolicy = *restart
	}
	s.AddCNF(cnf, options)
	if *timeout > 0 {
		timer := time.AfterFunc(time.Duration(*timeout*float64(time.Second)), s.Interrupt)
		defer timer.Stop()
	}

	// The profile has to be stopped explicitly: this command exits through
	// os.Exit to report the answer in its status, which skips deferred calls.
	stopProfile := func() {}
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "gomisat:", err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, "gomisat:", err)
			os.Exit(1)
		}
		stopProfile = func() {
			pprof.StopCPUProfile()
			f.Close()
		}
	}

	if *count {
		copt := &gomisat.CountOptions{
			UseCache:         *noCache == false,
			UseDecomposition: *noDecompose == false,
		}
		start := time.Now()
		models, stats := s.CountModels(options, copt)
		elapsed := time.Since(start)
		stopProfile()
		fmt.Printf("c vars %d clauses %d\n", s.NumVars(), len(cnf.Clauses))
		fmt.Printf("c decisions %d conflicts %d components %d\n",
			stats.Decisions, stats.Conflicts, stats.Components)
		fmt.Printf("c cache hits %d misses %d entries %d\n",
			stats.CacheHits, stats.CacheMiss, stats.CacheSize)
		fmt.Printf("c time %.6f s\n", elapsed.Seconds())
		fmt.Printf("s mc %s\n", models.String())
		if models.Sign() == 0 {
			os.Exit(20)
		}
		os.Exit(10)
	}

	start := time.Now()
	status := s.Solve(options)
	elapsed := time.Since(start)
	stopProfile()

	fmt.Printf("c vars %d clauses %d\n", s.NumVars(), len(cnf.Clauses))
	fmt.Printf("c conflicts %d propagations %d decisions %d restarts %d\n",
		s.Conflicts, s.Propagations, s.Decisions, s.Starts)
	fmt.Printf("c learnts %d blocked %d\n", s.NumLearnts(), s.Blocked)
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
