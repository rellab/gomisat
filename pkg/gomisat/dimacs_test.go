package gomisat

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

func TestDimacs01(t *testing.T) {
	cs, _ := ParseDimacs([]byte(`
	p cnf 5 6
	4 5 6 3 0
	-1 2 1 0
	`))
	for _, x := range cs {
		fmt.Println(x)
	}
}

func TestDimacs02(t *testing.T) {
	file, _ := os.Open("../../testdata/aim-100-1_6-no-1.cnf")
	defer file.Close()
	b, _ := io.ReadAll(file)
	cs, _ := ParseDimacs(b)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	result := s.Solve(options)
	fmt.Println("Result", result)
}

func TestDimacs03(t *testing.T) {
	file, _ := os.Open("../../testdata/aim-50-1_6-yes1-4.cnf")
	defer file.Close()
	b, _ := io.ReadAll(file)
	cs, _ := ParseDimacs(b)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	result := s.Solve(options)
	fmt.Println("Result", result)
	fmt.Println("  ", s.assigns)
}

func TestDimacs04(t *testing.T) {
	log.SetOutput(io.Discard)
	file, _ := os.Open("../../testdata/bf0432-007.cnf")
	defer file.Close()
	b, _ := io.ReadAll(file)
	cs, _ := ParseDimacs(b)
	s := NewSolver()
	options := DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	result := s.Solve(options)
	fmt.Println("Result", result)
	fmt.Println("  ", s.assigns)
}

// TestParseDimacsLayouts covers the layouts that occur in the benchmark corpus.
// The line-based parser this replaces silently turned one clause spread over
// several lines into several clauses, which made the whole SATLIB inductive
// inference family look unsatisfiable.
func TestParseDimacsLayouts(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    [][]int64
		vars    int
		wantErr bool
	}{
		{
			name:  "one clause per line",
			input: "p cnf 3 2\n1 -2 0\n2 3 0\n",
			want:  [][]int64{{1, -2}, {2, 3}},
			vars:  3,
		},
		{
			name:  "clause spanning lines",
			input: "p cnf 4 2\n1 -2\n3 0\n-3\n4\n0\n",
			want:  [][]int64{{1, -2, 3}, {-3, 4}},
			vars:  4,
		},
		{
			name:  "several clauses on one line",
			input: "p cnf 3 3\n1 0 -1 2 0 3 0\n",
			want:  [][]int64{{1}, {-1, 2}, {3}},
			vars:  3,
		},
		{
			name:  "percent terminates the formula",
			input: "p cnf 3 1\n1 2 0\n%\n0\n\n",
			want:  [][]int64{{1, 2}},
			vars:  3,
		},
		{
			name:  "comments anywhere",
			input: "c leading\np cnf 2 2\nc between\n1 0\nc more\n-2 0\n",
			want:  [][]int64{{1}, {-2}},
			vars:  2,
		},
		{
			name:  "unterminated last clause",
			input: "p cnf 2 2\n1 0\n-2\n",
			want:  [][]int64{{1}, {-2}},
			vars:  2,
		},
		{
			name:  "crlf",
			input: "c x\r\np cnf 2 1\r\n1 -2 0\r\n",
			want:  [][]int64{{1, -2}},
			vars:  2,
		},
		{
			name:  "stray zero is not the empty clause",
			input: "p cnf 2 1\n1 -2 0\n0\n",
			want:  [][]int64{{1, -2}},
			vars:  2,
		},
		{name: "no header", input: "1 2 0\n", wantErr: true},
		{name: "literals before the header", input: "1 2 0\np cnf 2 1\n", wantErr: true},
		{name: "malformed header", input: "p cnf 2\n1 0\n", wantErr: true},
		{name: "not a literal", input: "p cnf 2 1\n1 x 0\n", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cnf, err := ParseDimacsCNF([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if cnf.NumVars != tt.vars {
				t.Errorf("NumVars = %d, want %d", cnf.NumVars, tt.vars)
			}
			if len(cnf.Clauses) != len(tt.want) {
				t.Fatalf("got %d clauses %v, want %d %v", len(cnf.Clauses), cnf.Clauses, len(tt.want), tt.want)
			}
			for i := range tt.want {
				if len(cnf.Clauses[i]) != len(tt.want[i]) {
					t.Fatalf("clause %d = %v, want %v", i, cnf.Clauses[i], tt.want[i])
				}
				for j := range tt.want[i] {
					if cnf.Clauses[i][j] != tt.want[i][j] {
						t.Fatalf("clause %d = %v, want %v", i, cnf.Clauses[i], tt.want[i])
					}
				}
			}
		})
	}
}

// TestAddCNFCreatesDeclaredVariables checks that a variable the header declares
// but no clause mentions still exists in the solver. It is free, so it does not
// change satisfiability, but it does change the number of assignments, which
// matters once counting arrives.
func TestAddCNFCreatesDeclaredVariables(t *testing.T) {
	cnf, err := ParseDimacsCNF([]byte("p cnf 5 1\n1 -2 0\n"))
	if err != nil {
		t.Fatal(err)
	}
	s := NewSolver()
	options := DefaultSolverOptions()
	s.AddCNF(cnf, options)
	if got := s.NumVars(); got != 5 {
		t.Errorf("NumVars = %d, want 5", got)
	}
	if got := s.Solve(options); got != LTrue {
		t.Errorf("Solve = %v, want T", got)
	}
}
