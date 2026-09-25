#!/usr/bin/env python3
"""Decide every instance of the benchmark corpus with an independent solver and
write corpus/expected.tsv.

The directory names of the corpus encode what the SATLIB family is *said* to be,
which turned out not to be trustworthy: the Beijing family is mixed, not
satisfiable throughout, and 2bitadd_10 is unsatisfiable. This produces ground
truth instead, using CaDiCaL through pysat, so that gomibench -check compares
against something that does not share any assumption with the solver under test.
DIMACS is parsed here rather than by pysat, which rejects the SATLIB files, and
not by the Go implementation either, which is the one being checked.

    pip3 install --user python-sat
    scripts/expected-status.py [corpus-dir] > corpus/expected.tsv

Each instance is decided in its own process so that it can be killed when it runs
too long: pysat's interrupt does not reach CaDiCaL in this version.
"""
import os
import subprocess
import sys
from concurrent.futures import ThreadPoolExecutor

LIMIT_SECONDS = 60
WORKERS = max(2, (os.cpu_count() or 4))


def read_cnf(path):
    """Parse DIMACS as a stream of integers terminated by 0.

    A clause may span several lines, several clauses may share one line, '%' ends
    the formula (the SATLIB convention) and 'c' starts a comment anywhere.
    """
    clauses, clause, header = [], [], None
    with open(path, "r", errors="replace") as f:
        for row in f:
            row = row.strip()
            if not row or row[0] == "c":
                continue
            if row[0] == "%":
                break
            if row[0] == "p":
                fields = row.split()
                header = (int(fields[2]), int(fields[3]))
                continue
            for token in row.split():
                value = int(token)
                if value == 0:
                    if clause:
                        clauses.append(clause)
                        clause = []
                    continue
                clause.append(value)
    if clause:
        clauses.append(clause)
    return header, clauses


def decide_one(path):
    from pysat.solvers import Cadical153

    _, clauses = read_cnf(path)
    solver = Cadical153(bootstrap_with=clauses)
    try:
        result = solver.solve()
    finally:
        solver.delete()
    return "SAT" if result else "UNSAT"


def decide(path):
    try:
        out = subprocess.run(
            [sys.executable, os.path.abspath(__file__), "--one", path],
            capture_output=True, text=True, timeout=LIMIT_SECONDS)
    except subprocess.TimeoutExpired:
        return path, "UNKNOWN"
    status = out.stdout.strip()
    if status not in ("SAT", "UNSAT"):
        return path, "ERROR"
    return path, status


def main():
    if len(sys.argv) >= 3 and sys.argv[1] == "--one":
        print(decide_one(sys.argv[2]))
        return

    root = sys.argv[1] if len(sys.argv) > 1 else os.path.expanduser(
        os.environ.get("GOMISAT_CORPUS", "~/.cache/gomisat/corpus"))
    paths = []
    for dirpath, _, names in os.walk(root):
        for name in names:
            if name.endswith(".cnf"):
                paths.append(os.path.join(dirpath, name))
    paths.sort()

    print("# instance\tstatus")
    print(f"# decided by CaDiCaL through pysat, {LIMIT_SECONDS}s per instance")
    print(f"# {len(paths)} instances, paths relative to the corpus root")
    done = 0
    with ThreadPoolExecutor(max_workers=WORKERS) as pool:
        for path, status in pool.map(decide, paths):
            print(f"{os.path.relpath(path, root)}\t{status}", flush=True)
            done += 1
            if done % 200 == 0:
                print(f"  {done}/{len(paths)}", file=sys.stderr, flush=True)


if __name__ == "__main__":
    main()
