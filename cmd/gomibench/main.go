// Command gomibench runs the solver over a set of DIMACS instances and reports
// one CSV row per instance, so that a change to the search can be judged against
// a stored baseline instead of a single hand-timed instance.
//
//	gomibench -check testdata/satlib                      # correctness sweep
//	gomibench -o bench/baseline.csv testdata/satlib        # record a baseline
//	gomibench -baseline bench/baseline.csv testdata/satlib # compare against it
//
// With -check the expected answer is derived from the SATLIB directory and file
// names ("unsat-*" / "sat-*" directories, "-yes" / "-no" in the aim file names).
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rellab/gomisat/pkg/gomisat"
)

type result struct {
	name         string
	vars         int
	clauses      int
	status       string
	conflicts    uint64
	propagations uint64
	decisions    uint64
	restarts     uint64
	seconds      float64
}

var csvHeader = []string{
	"instance", "vars", "clauses", "status",
	"conflicts", "propagations", "decisions", "restarts", "seconds",
}

func (r result) row() []string {
	return []string{
		r.name,
		strconv.Itoa(r.vars),
		strconv.Itoa(r.clauses),
		r.status,
		strconv.FormatUint(r.conflicts, 10),
		strconv.FormatUint(r.propagations, 10),
		strconv.FormatUint(r.decisions, 10),
		strconv.FormatUint(r.restarts, 10),
		strconv.FormatFloat(r.seconds, 'f', 6, 64),
	}
}

func main() {
	timeout := flag.Float64("timeout", 0, "per-instance wall clock limit in seconds (0 = no limit)")
	confBudget := flag.Int64("conflicts", -1, "per-instance conflict limit (negative = no limit)")
	out := flag.String("o", "", "write the CSV to this file instead of stdout")
	baseline := flag.String("baseline", "", "compare the run against this CSV")
	check := flag.Bool("check", false, "verify the answers against the known status of each instance")
	expected := flag.String("expected", "", "file of known answers, as produced by scripts/expected-status.py")
	minTime := flag.Float64("min-time", 0.005, "ignore instances faster than this (seconds) when comparing timings")
	noLBD := flag.Bool("no-lbd", false, "manage learnt clauses by activity only, as MiniSat does")
	reduce := flag.String("reduce", "conflicts", "reduction trigger: conflicts (Glucose), size (MiniSat) or none")
	protectTier2 := flag.Bool("protect-tier2", false, "keep mid-tier (LBD <= 6) clauses out of the deletion candidates")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: gomibench [flags] <cnf file or directory>...")
		flag.PrintDefaults()
		os.Exit(2)
	}

	paths, err := collect(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "gomibench:", err)
		os.Exit(1)
	}
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "gomibench: no .cnf files found")
		os.Exit(1)
	}

	// The CSV is written as the run proceeds: a sweep over a large corpus takes
	// long enough that losing everything to an interruption matters, and the
	// partial file is useful on its own.
	w, closeCSV, err := openCSV(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gomibench:", err)
		os.Exit(1)
	}
	defer closeCSV()

	var truth map[string]string
	if *expected != "" {
		truth, err = readExpected(*expected)
		if err != nil {
			fmt.Fprintln(os.Stderr, "gomibench:", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%d known answers from %s\n", len(truth), *expected)
	}

	results := make([]result, 0, len(paths))
	mismatches := 0
	start := time.Now()
	lastReport := start
	for i, path := range paths {
		r, err := run(path, *timeout, *confBudget, *noLBD, *reduce, *protectTier2)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gomibench: %s: %v\n", path, err)
			os.Exit(1)
		}
		results = append(results, r)
		if err := w.Write(r.row()); err != nil {
			fmt.Fprintln(os.Stderr, "gomibench:", err)
			os.Exit(1)
		}
		w.Flush()
		if time.Since(lastReport) > 15*time.Second {
			fmt.Fprintf(os.Stderr, "  %d/%d  %.0fs elapsed  last %s %s %.2fs\n",
				i+1, len(paths), time.Since(start).Seconds(), filepath.Base(r.name), r.status, r.seconds)
			lastReport = time.Now()
		}
		if *check {
			// UNKNOWN means the instance was not solved within the limits, which
			// is not a wrong answer. Only a decided answer can contradict.
			want := knownStatus(truth, path)
			if want != "" && want != "UNKNOWN" && r.status != "UNKNOWN" && want != r.status {
				fmt.Fprintf(os.Stderr, "MISMATCH %s: got %s, want %s\n", r.name, r.status, want)
				mismatches++
			}
		}
	}
	elapsed := time.Since(start)
	closeCSV()

	solved := 0
	for _, r := range results {
		if r.status == "SAT" || r.status == "UNSAT" {
			solved++
		}
	}
	fmt.Fprintf(os.Stderr, "%d instances, %d solved, %.3fs total\n", len(results), solved, elapsed.Seconds())

	if *baseline != "" {
		if err := compare(*baseline, results, *minTime); err != nil {
			fmt.Fprintln(os.Stderr, "gomibench:", err)
			os.Exit(1)
		}
	}
	if mismatches > 0 {
		fmt.Fprintf(os.Stderr, "%d mismatches\n", mismatches)
		os.Exit(1)
	}
}

func collect(args []string) ([]string, error) {
	paths := make([]string, 0, 64)
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}
		if info.IsDir() == false {
			paths = append(paths, arg)
			continue
		}
		err = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() == false && strings.HasSuffix(path, ".cnf") {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(paths)
	return paths, nil
}

func run(path string, timeout float64, confBudget int64, noLBD bool, reduce string, protectTier2 bool) (result, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return result{}, err
	}
	cnf, err := gomisat.ParseDimacsCNF(buf)
	if err != nil {
		return result{}, err
	}

	s := gomisat.NewSolver()
	options := gomisat.DefaultSolverOptions()
	options.UseLBD = noLBD == false
	options.ReduceByConflicts = reduce == "conflicts"
	options.NoReduce = reduce == "none"
	options.ProtectTier2 = protectTier2
	s.AddCNF(cnf, options)
	if confBudget >= 0 {
		s.SetConfBudget(confBudget)
	}
	var timer *time.Timer
	if timeout > 0 {
		timer = time.AfterFunc(time.Duration(timeout*float64(time.Second)), s.Interrupt)
	}

	begin := time.Now()
	status := s.Solve(options)
	seconds := time.Since(begin).Seconds()
	if timer != nil {
		timer.Stop()
	}

	name := ""
	switch status {
	case gomisat.LTrue:
		name = "SAT"
	case gomisat.LFalse:
		name = "UNSAT"
	default:
		name = "UNKNOWN"
	}
	return result{
		name:         path,
		vars:         s.NumVars(),
		clauses:      len(cnf.Clauses),
		status:       name,
		conflicts:    s.Conflicts,
		propagations: s.Propagations,
		decisions:    s.Decisions,
		restarts:     s.Starts,
		seconds:      seconds,
	}, nil
}

// instanceKey identifies an instance by its family directory and file name, so
// that a table of known answers does not depend on where the corpus lives.
func instanceKey(path string) string {
	return filepath.Join(filepath.Base(filepath.Dir(path)), filepath.Base(path))
}

// readExpected loads a table of known answers: one "<family>/<instance>.cnf",
// status per line, '#' starting a comment.
func readExpected(path string) (map[string]string, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	truth := make(map[string]string, 2048)
	for _, line := range strings.Split(string(buf), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}
		truth[instanceKey(fields[0])] = strings.TrimSpace(fields[1])
	}
	if len(truth) == 0 {
		return nil, fmt.Errorf("%s: no answers found", path)
	}
	return truth, nil
}

// knownStatus prefers the table of answers decided by an independent solver and
// falls back to the claim encoded in the directory name. The fallback is only a
// claim: the Beijing family is named after what SATLIB says about it and is in
// fact mixed, which is why the table exists.
func knownStatus(truth map[string]string, path string) string {
	if status, ok := truth[instanceKey(path)]; ok {
		return status
	}
	return expectedStatus(path)
}

// expectedStatus is the fallback rule, derived from the corpus layout rather than
// from the solver, and only as good as the family names.
func expectedStatus(path string) string {
	dir := filepath.Base(filepath.Dir(path))
	name := filepath.Base(path)
	switch {
	case strings.HasPrefix(dir, "unsat-"):
		return "UNSAT"
	case strings.HasPrefix(dir, "sat-"):
		return "SAT"
	case strings.Contains(name, "-yes"):
		return "SAT"
	case strings.Contains(name, "-no"):
		return "UNSAT"
	}
	return ""
}

// openCSV returns a writer for the results plus a function that flushes and
// closes it; calling the latter twice is harmless.
func openCSV(path string) (*csv.Writer, func(), error) {
	f := os.Stdout
	closeFile := func() {}
	if path != "" {
		if dir := filepath.Dir(path); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, nil, err
			}
		}
		created, err := os.Create(path)
		if err != nil {
			return nil, nil, err
		}
		f = created
		closeFile = func() { created.Close() }
	}
	w := csv.NewWriter(f)
	if err := w.Write(csvHeader); err != nil {
		return nil, nil, err
	}
	w.Flush()
	done := false
	return w, func() {
		if done {
			return
		}
		done = true
		w.Flush()
		closeFile()
	}, nil
}

// compare reports how the run moved relative to a baseline: the median time
// ratio, the instances that got materially slower, and any changed answer.
func compare(path string, results []result, minTime float64) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return fmt.Errorf("%s: empty baseline", path)
	}
	base := make(map[string]result, len(rows)-1)
	for _, row := range rows[1:] {
		if len(row) < len(csvHeader) {
			continue
		}
		seconds, _ := strconv.ParseFloat(row[8], 64)
		conflicts, _ := strconv.ParseUint(row[4], 10, 64)
		base[row[0]] = result{name: row[0], status: row[3], conflicts: conflicts, seconds: seconds}
	}

	type delta struct {
		name  string
		ratio float64
	}
	ratios := make([]delta, 0, len(results))
	changed := 0
	for _, r := range results {
		b, ok := base[r.name]
		if ok == false {
			continue
		}
		if b.status != r.status {
			fmt.Fprintf(os.Stderr, "ANSWER CHANGED %s: %s -> %s\n", r.name, b.status, r.status)
			changed++
		}
		// Ignore instances that are too fast to time meaningfully.
		if b.seconds < minTime {
			continue
		}
		ratios = append(ratios, delta{r.name, r.seconds / b.seconds})
	}
	if len(ratios) == 0 {
		fmt.Fprintln(os.Stderr, "no comparable instances in the baseline")
		return nil
	}
	sort.Slice(ratios, func(i, j int) bool { return ratios[i].ratio < ratios[j].ratio })
	median := ratios[len(ratios)/2].ratio
	fmt.Fprintf(os.Stderr, "median time ratio vs baseline: %.3f (%d instances)\n", median, len(ratios))
	for i := len(ratios) - 1; i >= 0 && i >= len(ratios)-5; i-- {
		if ratios[i].ratio > 1.2 {
			fmt.Fprintf(os.Stderr, "  slower %.2fx  %s\n", ratios[i].ratio, ratios[i].name)
		}
	}
	if changed > 0 {
		return fmt.Errorf("%d answers changed against the baseline", changed)
	}
	return nil
}
