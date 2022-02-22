package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/mitchellh/go-sat"
	"github.com/mitchellh/go-sat/cnf"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func ParseDimacs(b []byte) ([][]int, error) {
	in := bytes.NewBuffer(b)
	s := bufio.NewScanner(in)
	var nlits, nclauses int
	for s.Scan() {
		a := strings.Fields(s.Text())
		if len(a) >= 1 && a[0] == "c" { // comment line
			continue
		}
		if len(a) >= 2 && a[0] == "p" && a[1] == "cnf" {
			nlits, _ = strconv.Atoi(a[2])
			nclauses, _ = strconv.Atoi(a[3])
			log.Println("Number of literals:", nlits)
			log.Println("Number of clauses:", nclauses)
			break
		}
	}
	clauses := make([][]int, 0)
	for s.Scan() {
		a := strings.Fields(s.Text())
		if len(a) >= 2 {
			lits := make([]int, 0, len(a)-1)
			if len(a) >= 1 && a[0] == "c" { // comment line
				continue
			}
			for _, x := range a {
				if v, err := strconv.Atoi(x); err == nil {
					if v != 0 {
						lits = append(lits, int(v))
					}
				} else {
					log.Println("Atoi fails", err)
				}
			}
			clauses = append(clauses, lits)
		}
	}
	if len(clauses) != nclauses {
		log.Println("Did not match the number of clauses generated to the number of clauses defined on the header of DIMCS:", len(clauses))
	}
	return clauses, nil
}

func Mem_used() float64 {
	pid := os.Getpid()
	name := fmt.Sprintf("/proc/%d/statm", pid)
	var value int
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	n, err := fmt.Fscanf(file, "%d", &value)
	n = n +1
	return float64(value) * float64(os.Getpagesize()) / (1024.0*1024.0)
}

func main() {
	fname := os.Args[1]
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := ParseDimacs(buf)

	formula := cnf.NewFormulaFromInts(cs)
	s := sat.New()
	s.AddFormula(formula)
	start := time.Now()
	sat := s.Solve()
	end := time.Now()
	fmt.Printf("Solved: %v\n", sat)
	fmt.Printf("Solution:\n")
	fmt.Printf("computation time : %.8f (sec)\n", (end.Sub(start)).Seconds())
	fmt.Printf("Memory used : %.2f MB\n", Mem_used())
}
