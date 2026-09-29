# Using gomisat

Three things live in this module: a SAT solver, an exact model counter, and two
ways of stating the problems they answer -- reliability models and finite-domain
constraint problems. This page is the user's guide; [DEVELOPING.md](DEVELOPING.md)
is for working on the code and [../DESIGN.md](../DESIGN.md) records why it is the
way it is.

## Building

```console
$ git clone https://github.com/rellab/gomisat && cd gomisat
$ make build          # bin/gomisat, bin/gomibench
$ make check          # go vet and the regression suite
```

Go 1.24 or later. There are no dependencies outside the standard library.

## The command line

### Solving and counting one file

```console
$ ./bin/gomisat testdata/satlib/sat-uniform-20-91/uf20-01.cnf
c vars 20 clauses 91
c conflicts 6 propagations 58 decisions 10 restarts 1
c learnts 6 blocked 0
c time 0.000091 s
s SATISFIABLE

$ ./bin/gomisat -model file.cnf      # also print the assignment as a DIMACS v-line
$ ./bin/gomisat -count file.cnf      # count the models instead of finding one
$ ./bin/gomisat -timeout 60 file.cnf # give up after a minute
```

The exit status follows the SAT competition convention: 10 satisfiable, 20
unsatisfiable, 0 unknown, 1 error. So a script can branch on it.

Flags worth knowing: `-restart luby|geometric|ema|ema-block` chooses the restart
policy, `-no-lbd` manages learnt clauses by activity alone as MiniSat does,
`-no-decompose` and `-no-cache` turn off the two things that make counting
affordable, and `-cpuprofile` writes a profile.

### Measuring over a corpus

`gomibench` runs the solver over many instances and writes one CSV row each, which
is how a change to the search is judged:

```console
$ make corpus          # fetch the benchmark corpus (about 99 MB, once)
$ make corpus-sweep    # record the results, checking every decided answer
$ make corpus-compare  # re-run and compare against the recorded results
```

The corpus lands in `$GOMISAT_CORPUS`, by default `~/.cache/gomisat/corpus`, not in
the working tree. Answers are checked against `corpus/expected.tsv`, which was
decided by CaDiCaL rather than inferred from the family names -- the names turned
out to be wrong about one instance.

### Comparing encodings

`gomiencode` states the same reliability system three ways and reports what each
costs to count:

```console
$ go run ./cmd/gomiencode -n 20,40,60
$ go run ./cmd/gomiencode -family multistate -n 20,30,40
```

## As a library

### Satisfiability

```go
s := gomisat.NewSolver()
options := gomisat.DefaultSolverOptions()

cnf, err := gomisat.ParseDimacsCNF(data)
if err != nil {
    return err
}
s.AddCNF(cnf, options)

switch s.Solve(options) {
case gomisat.LTrue:
    model := s.Model()                  // map[Var]LBool
    _ = model[gomisat.Var(0)]           // the value of DIMACS variable 1
case gomisat.LFalse:
    // unsatisfiable
}
```

Clauses can also be added directly as DIMACS codes with
`s.AddClauseFromCode([]int64{1, -2, 3}, options)`.

### A sequence of queries

One solver can answer many questions. Assumptions are replaced on every call, and a
negative answer comes with the subset of them that caused it:

```go
for _, assumptions := range queries {
    if s.SolveWithAssumptions(assumptions, options) == gomisat.LFalse {
        core := s.UnsatCore()  // a subset of assumptions, unsatisfiable on its own
        _ = core
    }
}
```

### Counting

```go
count, stats := s.CountModels(options, gomisat.DefaultCountOptions())
fmt.Println(count, stats.Decisions, stats.CacheHits)
```

With weights, the count becomes a probability:

```go
weights := gomisat.NewWeights(s.NumVars())
for v := 0; v < s.NumVars(); v++ {
    weights.SetProbability(gomisat.Var(v), 0.9)
}
value, _ := s.WeightedCount(options, gomisat.DefaultCountOptions(), weights)
```

The accumulator is a `big.Float`, because the probability of a long series of
components leaves the range of a `float64` and an answer that has silently
underflowed to zero is worse than no answer.

**Two rules.** Count on a solver that has not solved anything: the counter
propagates with the problem clauses alone and a learnt clause can join two
components that must stay independent. And make sure the auxiliary variables of
your encoding are determined by the variables being counted, or the same solution
is counted twice -- see the warning under *Encodings* below.

### A study: many counting queries over one model

```go
study := gomisat.NewStudy(s, options, gomisat.DefaultCountOptions(), weights)
for i := 0; i < n; i++ {
    value, _ := study.Count(failed[i])   // condition on component i having failed
    _ = value
}
fmt.Println(study.Totals().CacheHits)
```

The component cache is kept between queries, which is worth about a factor of four
on the sequences measured. Queries are expressed as assumptions, not as edits to the
formula: that is what keeps a cached component valid. If the weights change, call
`study.ClearCache`.

## Reliability models

`pkg/reliability` builds the structure function of a coherent system and its
reliability is the weighted count of it.

```go
m := reliability.New()

pumps := m.Events("pump", 3)
for _, p := range pumps {
    m.SetProbability(p, 0.95)
}
m.Assert(m.AtLeast(2, pumps...))   // two of the three must work

fmt.Println(m.Reliability())       // 0.992750
```

Components with several states use the one-hot encoding when probabilities are
involved:

```go
levels := m.MultiStateOneHot("valve", 4)
m.SetStateProbabilities(levels, []float64{0.05, 0.15, 0.3, 0.5})
working := m.AtLeastState(levels, 2)   // in state 2 or better
```

`m.MultiState` is the other encoding, one variable per level meaning "this state or
better". It makes a threshold a single literal, which is convenient when thresholds
are what a study varies, but it cannot carry a state distribution on independent
literal weights. Use one-hot for probabilities and the order encoding for counting
configurations.

## Constraint problems

`pkg/csp` states finite-domain problems -- integer variables, linear constraints,
Boolean structure -- and solves as well as counts them.

```go
m := csp.New()
x, y := m.IntVarRange(0, 10), m.IntVarRange(0, 10)

m.Add(csp.LeZero(csp.NewSum(map[*csp.IntVar]int{x: 2, y: 3}, -12))) // 2x + 3y <= 12
m.Add(csp.GeZero(csp.NewSum(map[*csp.IntVar]int{x: 1, y: -1}, 0)))  // x >= y

if sol, ok := m.Solve(); ok {
    fmt.Println(sol.Int(x), sol.Int(y))
}
count, _ := m.Count()   // 13
```

A sum is a map from variables to coefficients plus a constant, and `LeZero`,
`GeZero`, `EqZero` and `NeZero` compare it against zero. `csp.And`, `csp.Or`,
`csp.Imp`, `csp.Iff` combine constraints; `m.NewBool` adds a Boolean variable.

A Boolean variable cannot appear in a linear sum. Use an integer variable of domain
0..1 where you want to count how many of something hold, and tie it to a condition
with `Iff` if it is standing in for one.

Solving only? Set `m.Definitional = false` for a smaller, faster encoding. It is not
the default because counting needs the larger one and would be quietly wrong without
it.

## Encodings: the one thing that will bite you

Counting demands more of an encoding than solving does. An auxiliary variable
introduced for a subformula must be **determined** by the original variables; if it
can take either value while the subformula holds, the same solution is counted more
than once and nothing complains.

It does not always go wrong, which is what makes it dangerous. In the encoder this
module imported, a disjunction of two mutually exclusive branches came out right, a
disjunction of two overlapping ones was inflated by a factor of 1.3, and three of
them by 1.65. The first case anyone tries passes.

So: if you write an encoding by hand, count a small instance and compare against
enumeration. Both model builders here do that in their own tests, and it is the only
reason their encodings can be trusted.

## Where things are

| | |
| --- | --- |
| `pkg/gomisat` | the solver and the counter |
| `pkg/reliability` | structure functions of coherent systems |
| `pkg/csp` | finite-domain constraint problems |
| `cmd/gomisat` | solve or count one file |
| `cmd/gomibench` | sweep a corpus, check answers, compare against a baseline |
| `cmd/gomiencode` | compare encodings of one model by counting cost |
| `testdata/satlib` | 2185 instances, committed, used by the tests |
| `corpus/` | the manifest and known answers of the larger corpus |
| `bench/` | recorded measurements |
